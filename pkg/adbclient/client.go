package adbclient

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/johnnyipcom/androidtool/pkg/logger"
	adb "github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/wire"
)

const (
	DefaultInstallPath = "/data/local/tmp/app.apk"
	DefaultVideoPath   = "/sdcard/video.mp4"
	DefaultPort        = adb.AdbPort
)

// Client is a ui wrapper around the adb client.
type Client struct {
	adb    *adb.Adb
	log    logger.Logger
	dialer *dialer
	events chan DeviceStateChangedEvent
	port   int

	propertyMu  sync.RWMutex
	installPath string
	videoPath   string
}

// New creates a new client.
func NewClient(port int, log logger.Logger) (*Client, error) {
	dialer := &dialer{}
	config := adb.ServerConfig{Dialer: dialer, Port: port}

	innerLog := log.WithField("component", "ADBClient")
	innerLog.Infof("Creating ADB client on port %d", port)

	adb, err := adb.NewWithConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		adb:    adb,
		log:    innerLog,
		dialer: dialer,
		port:   port,
		events: make(chan DeviceStateChangedEvent),
	}, nil
}

// Start starts client.
func (c *Client) Start(ctx context.Context) error {
	c.log.Info("Starting ADB server...")
	if err := c.adb.StartServer(); err != nil {
		return err
	}

	go c.deviceWatcher(ctx)
	return nil
}

// Kill kills the client.
func (c *Client) Stop() {
	c.log.Info("Stopping ADB client...")
	if err := c.adb.KillServer(); err != nil {
		c.log.Fatal(err)
	}
}

// Version returns the version of the client.
func (c *Client) Version() string {
	return "0.0.1"
}

// ServerVersion returns the version of the ADB server.
func (c *Client) ServerVersion() int {
	ver, err := c.adb.ServerVersion()
	if err != nil {
		return -1
	}

	return ver
}

func (c *Client) Port() int {
	return c.port
}

// deviceWatcher watches for device state changes.
func (c *Client) deviceWatcher(ctx context.Context) {
	c.log.Info("Starting device watcher...")
	defer close(c.events)

	watcher := c.adb.NewDeviceWatcher()
	defer watcher.Shutdown()

	for {
		select {
		case <-ctx.Done():
			c.log.Debug("Device watcher stopped.")
			return

		case event := <-watcher.C():
			c.log.Infof("Device %s changed state to %s", event.Serial, event.NewState)
			c.events <- NewDeviceStateChangedEvent(event)
		}
	}
}

// DeviceWatcher returns a channel that receives device state changes.
func (c *Client) DeviceWatcher() <-chan DeviceStateChangedEvent {
	return c.events
}

// Device returns a device with the given serial.
func (c *Client) GetDevice(serial string) (*Device, error) {
	return NewDevice(c, c.adb.Device(adb.DeviceWithSerial(serial)))
}

func (c *Client) GetAnyOnlineDevice() (*Device, error) {
	deviceInfos, err := c.adb.ListDevices()
	if err != nil {
		return nil, err
	}

	for _, deviceInfo := range deviceInfos {
		device := c.adb.Device(adb.DeviceWithSerial(deviceInfo.Serial))

		state, err := device.State()
		if err != nil {
			continue
		}

		if adb.StateOnline == state {
			return NewDevice(c, device)
		}
	}

	return nil, fmt.Errorf("no online device found")
}

// SetInstallPath sets the path to install the apk.
func (c *Client) SetInstallPath(path string) {
	c.propertyMu.Lock()
	defer c.propertyMu.Unlock()

	c.installPath = path
}

func (c *Client) GetInstallPath() string {
	c.propertyMu.RLock()
	defer c.propertyMu.RUnlock()

	if c.installPath == "" {
		return DefaultInstallPath
	}

	return c.installPath
}

func (c *Client) SetVideoPath(path string) {
	c.propertyMu.Lock()
	defer c.propertyMu.Unlock()

	c.videoPath = path
}

func (c *Client) GetVideoPath() string {
	c.propertyMu.RLock()
	defer c.propertyMu.RUnlock()

	if c.videoPath == "" {
		return DefaultVideoPath
	}

	return c.videoPath
}

// Install installs a package to the device.
func (c *Client) Install(device *Device, apkPath string) (string, error) {
	c.log.Infof("Installing %s...", apkPath)

	result, err := c.adb.Device(adb.DeviceWithSerial(device.Serial)).RunCommand("pm", "install", "-r", apkPath)
	c.log.Debug(result)
	if err != nil {
		return "", err
	}

	return result, nil
}

// parseKeyVal parses a key:val pair and returns key, val.
func parseKeyVal(pair string, sep string) (string, string) {
	split := strings.Split(pair, sep)
	switch len(split) {
	case 1:
		return "", split[0]
	case 2:
		return split[0], split[1]
	default:
		return "", ""
	}
}

func getProp(device *adb.Device, prop string) (string, error) {
	result, err := device.RunCommand("getprop", prop)
	if err != nil {
		return "", err
	}

	_, value := parseKeyVal(strings.Trim(result, " \n"), ":")
	return value, nil
}

func wm(device *adb.Device, prop string) (string, error) {
	result, err := device.RunCommand("wm", prop)
	if err != nil {
		return "", err
	}

	_, value := parseKeyVal(strings.Trim(result, " \n"), ":")
	return value, nil
}

func diskUsage(device *adb.Device, path string) (int, error) {
	result, err := device.RunCommand("du", "-b", path)
	if err != nil {
		return 0, err
	}

	key, _ := parseKeyVal(strings.Trim(result, " \n"), "\t")
	return strconv.Atoi(key)
}

// GetProp returns a property of the device.
func (c *Client) GetProp(device *Device, prop string) (string, error) {
	c.log.Info("Getting %s...", prop)
	return getProp(c.adb.Device(adb.DeviceWithSerial(device.Serial)), prop)
}

// dialDevice returns a connection to the device.
func (c *Client) dialDevice(device *Device) (*wire.Conn, error) {
	conn, err := c.adb.Dial()
	if err != nil {
		return nil, err
	}

	req := fmt.Sprintf("host:transport:%s", device.Serial)
	if err := wire.SendMessageString(conn, req); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := conn.ReadStatus(req); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

// sendCommand sends a command to the device anc checks the status of the command.
func (c *Client) sendCommand(device *Device, cmd string) (*wire.Conn, error) {
	conn, err := c.dialDevice(device)
	if err != nil {
		return nil, err
	}

	req := fmt.Sprintf("shell:%s", cmd)
	c.log.Debugf("Sending command: %s", req)
	if err := wire.SendMessageString(conn, req); err != nil {
		conn.Close()
		return nil, err
	}

	status, err := conn.ReadStatus(req)
	if err != nil {
		conn.Close()
		return nil, err
	}

	c.log.Debugf("Got status: %s", status)
	return conn, nil
}

// runCommand runs a command on the device.
func (c *Client) runCommand(device *Device, cmd string, args ...string) ([]byte, error) {
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	conn, err := c.sendCommand(device, cmd)
	if err != nil {
		return nil, err
	}

	defer conn.Close()
	result, err := conn.ReadUntilEof()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// RemoveFile removes a file from the device.
func (c *Client) RemoveFile(device *Device, path string) error {
	c.log.Infof("Removing %s...", path)

	resp, err := c.runCommand(device, "rm -f -v", path)
	if err != nil {
		return err
	}

	c.log.Debug(string(resp))
	return nil
}

// SendLink start a browser and send a link to the device.
func (c *Client) SendLink(device *Device, link string) error {
	c.log.Infof("Sending link %s...", link)

	_, err := url.ParseRequestURI(link)
	if err != nil {
		return err
	}

	resp, err := c.runCommand(device, "am", "start", "-a", "android.intent.action.VIEW", "-d", link)
	if err != nil {
		return err
	}

	c.log.Debug(string(resp))
	return nil
}
