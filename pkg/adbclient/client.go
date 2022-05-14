package adbclient

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/johnnyipcom/androidtool/pkg/logger"
	adb "github.com/zach-klippenstein/goadb"
)

const (
	DefaultInstallPath = "/data/local/tmp/app.apk"
	DefaultPort        = adb.AdbPort
)

// Client is a ui wrapper around the adb client.
type Client struct {
	adb    *adb.Adb
	log    logger.Logger
	events chan DeviceStateChangedEvent
	port   int

	installPath string
}

// New creates a new client.
func NewClient(port int, log logger.Logger) (*Client, error) {
	config := adb.ServerConfig{Port: port}

	innerLog := log.WithField("component", "ADBClient")
	innerLog.Infof("Creating ADB client on port %d", port)

	adb, err := adb.NewWithConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		adb:    adb,
		log:    innerLog,
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

type progressFunc func(sentBytes int64, totalBytes int64)

type uploadOptions struct {
	progressFunc progressFunc
}

// UploadOption is an option for uploading file.
type UploadOption interface {
	apply(*uploadOptions) error
}

type progressUploadOption struct {
	progressFunc progressFunc
}

func (o progressUploadOption) apply(opts *uploadOptions) error {
	opts.progressFunc = o.progressFunc
	return nil
}

func WithProgress(f func(sentBytes int64, totalBytes int64)) UploadOption {
	return progressUploadOption{f}
}

type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) {
	return rf(p)
}

// Upload uploads a file to the device.
func (c *Client) Upload(ctx context.Context, device *Device, src, dst string, opts ...UploadOption) error {
	c.log.Infof("Uploading to %s...", dst)

	var options uploadOptions
	for _, opt := range opts {
		err := opt.apply(&options)
		if err != nil {
			return nil
		}
	}

	file, err := os.Open(src)
	if err != nil {
		return err
	}

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	w, err := c.adb.Device(adb.DeviceWithSerial(device.Serial)).OpenWrite(dst, os.FileMode(0664), time.Now())
	if err != nil {
		return err
	}

	total := 0
	defer w.Close()
	_, err = io.Copy(w, readerFunc(func(b []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()

		default:
			n, err := file.Read(b)
			if err != nil {
				return 0, err
			}

			total += n
			if options.progressFunc != nil {
				options.progressFunc(int64(total), fi.Size())
			}
			return n, err
		}
	}))

	return err
}

// SetInstallPath sets the path to install the apk.
func (c *Client) SetInstallPath(path string) {
	c.installPath = path
}

func (c *Client) GetInstallPath() string {
	if c.installPath == "" {
		return DefaultInstallPath
	}

	return c.installPath
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
func parseKeyVal(pair string) (string, string) {
	split := strings.Split(pair, ":")
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

	_, value := parseKeyVal(strings.Trim(result, " \n"))
	return value, nil
}

func wm(device *adb.Device, prop string) (string, error) {
	result, err := device.RunCommand("wm", prop)
	if err != nil {
		return "", err
	}

	_, value := parseKeyVal(strings.Trim(result, " \n"))
	return value, nil
}

// GetProp returns a property of the device.
func (c *Client) GetProp(device *Device, prop string) (string, error) {
	c.log.Info("Getting %s...", prop)
	return getProp(c.adb.Device(adb.DeviceWithSerial(device.Serial)), prop)
}
