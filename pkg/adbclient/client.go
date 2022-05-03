package adbclient

import (
	"context"
	"fmt"
	"io"
	"log"

	adb "github.com/zach-klippenstein/goadb"
)

// Client is a ui wrapper around the adb client.
type Client struct {
	adb    *adb.Adb
	log    *log.Logger
	events chan DeviceStateChangedEvent
}

// New creates a new client.
func NewClient(l *log.Logger) (*Client, error) {
	innerLog := log.New(l.Writer(), "[Client] ", 0)
	innerLog.SetFlags(log.LstdFlags | log.Lshortfile)

	innerLog.Println("Creating ADB client...")
	adb, err := adb.New()
	if err != nil {
		return nil, err
	}

	return &Client{
		adb:    adb,
		log:    innerLog,
		events: make(chan DeviceStateChangedEvent),
	}, nil
}

// Start starts client.
func (c *Client) Start(ctx context.Context) error {
	c.log.Println("Starting ADB server...")
	if err := c.adb.StartServer(); err != nil {
		return err
	}

	go c.deviceWatcher(ctx)
	return nil
}

// Kill kills the client.
func (c *Client) Stop() {
	c.log.Println("Stopping ADB client...")
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

// deviceWatcher watches for device state changes.
func (c *Client) deviceWatcher(ctx context.Context) {
	c.log.Println("Starting device watcher...")
	defer close(c.events)

	watcher := c.adb.NewDeviceWatcher()
	defer watcher.Shutdown()

	for {
		select {
		case <-ctx.Done():
			c.log.Println("Device watcher stopped.")
			return

		case event := <-watcher.C():
			c.log.Println("Device state changed:", event.Serial, event.NewState)
			c.events <- NewDeviceStateChangedEvent(event)
		}
	}
}

func (c *Client) DeviceWatcher() <-chan DeviceStateChangedEvent {
	return c.events
}

func (c *Client) Device(serial string) (*Device, error) {
	return NewDevice(c.adb.Device(adb.DeviceWithSerial(serial)))
}

// Install installs the given APK file.
func (c *Client) Install(r io.Reader) error {
	deviceInfos, err := c.adb.ListDevices()
	if err != nil {
		return err
	}

	var validDevice *adb.Device
	for _, deviceInfo := range deviceInfos {
		device := c.adb.Device(adb.DeviceWithSerial(deviceInfo.Serial))

		state, err := device.State()
		if err != nil {
			continue
		}

		if adb.StateOnline == state {
			validDevice = device
			break
		}
	}

	if validDevice == nil {
		return fmt.Errorf("no valid device found")
	}

	return nil
}
