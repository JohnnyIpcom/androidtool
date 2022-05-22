package adbclient

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"strings"
	"time"

	"github.com/johnnyipcom/androidtool/pkg/logger"
	adb "github.com/zach-klippenstein/goadb"
	"github.com/zach-klippenstein/goadb/wire"
)

const (
	DefaultInstallPath = "/data/local/tmp/app.apk"
	DefaultPort        = adb.AdbPort
)

// Client is a ui wrapper around the adb client.
type Client struct {
	adb    *adb.Adb
	log    logger.Logger
	dialer *dialer
	events chan DeviceStateChangedEvent
	port   int

	installPath string
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

func (c *Client) runCommand(device *Device, cmd string, args ...string) ([]byte, error) {
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

type logcatOptions struct {
	tag      string
	pid      int
	priority LogcatPriority
}

func (o logcatOptions) String() string {
	return fmt.Sprintf("tag:%s pid:%d priority:%s", o.tag, o.pid, o.priority)
}

func (o logcatOptions) Options() []string {
	var options []string
	if o.pid != 0 {
		options = append(options, fmt.Sprintf("--pid %d", o.pid))
	}

	// '*' by itself means '*:D' and <tag> by itself means <tag>:V.
	// If no '*' filterspec or -s on command line, all filter defaults to '*:V'.
	//  eg: '*:S <tag>' prints only <tag>, '<tag>:S' suppresses all <tag> log messages.
	if o.tag != "" {
		if o.priority != Verbose {
			options = append(options, fmt.Sprintf("-s %s:%s", o.tag, o.priority))
		} else {
			options = append(options, fmt.Sprintf("-s %s", o.tag))
		}
	} else {
		if o.priority != Debug {
			options = append(options, fmt.Sprintf("-s *:%s", o.priority))
		} else {
			options = append(options, "-s *")
		}
	}

	return options
}

// LogcatOption is an option for logcat.
type LogcatOption interface {
	apply(*logcatOptions) error
}

type logcatTagOption struct {
	tag string
}

func (o logcatTagOption) apply(opts *logcatOptions) error {
	opts.tag = o.tag
	return nil
}

type logcatPidOption struct {
	pid int
}

func (o logcatPidOption) apply(opts *logcatOptions) error {
	opts.pid = o.pid
	return nil
}

type logcatPriorityOption struct {
	priority LogcatPriority
}

func (o logcatPriorityOption) apply(opts *logcatOptions) error {
	opts.priority = o.priority
	return nil
}

func WithLogcatTag(tag string) LogcatOption {
	return logcatTagOption{tag}
}

func WithLogcatPid(pid int) LogcatOption {
	return logcatPidOption{pid}
}

func WithLogcatPriority(priority LogcatPriority) LogcatOption {
	return logcatPriorityOption{priority}
}

func (c *Client) ClearLogcat(device *Device) error {
	c.log.Info("Clearing logcat...")
	resp, err := c.runCommand(device, "logcat -c")
	if err != nil {
		return err
	}

	c.log.Debugf("Got response: %s", resp)
	return nil
}

// Logcat returns a logcat watcher.
func (c *Client) Logcat(device *Device, opts ...LogcatOption) (*LogcatWatcher, error) {
	c.log.Info("Getting logcat...")

	var options logcatOptions
	for _, opt := range opts {
		err := opt.apply(&options)
		if err != nil {
			return nil, err
		}
	}

	conn, err := c.sendCommand(device, fmt.Sprintf("logcat -v threadtime %s", strings.Join(options.Options(), " ")))
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &LogcatWatcher{
		reader: c.dialer.reader,
		conn:   conn,
		log:    c.log.WithField("device", device.Serial),
	}, nil
}

func (c *Client) Screenshot(device *Device) (image.Image, error) {
	c.log.Info("Taking screenshot...")
	resp, err := c.runCommand(device, "screencap -p")
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(resp))
	if err != nil {
		return nil, err
	}

	return img, nil
}
