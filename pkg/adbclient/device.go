package adbclient

import (
	"fmt"
	"strconv"
	"strings"

	adb "github.com/zach-klippenstein/goadb"
)

type DeviceState int8

const (
	StateInvalid DeviceState = iota
	StateUnauthorized
	StateDisconnected
	StateOffline
	StateOnline
)

func (state DeviceState) String() string {
	switch state {
	case StateUnauthorized:
		return "unauthorized"
	case StateDisconnected:
		return "disconnected"
	case StateOffline:
		return "offline"
	case StateOnline:
		return "online"
	default:
		return "invalid"
	}
}

// DisplayParams is the display parameters.
type DisplayParams struct {
	Width   int
	Height  int
	Density int
}

// Device is a wrapper around an adb.Device.
type Device struct {
	Serial     string
	Product    string
	Model      string
	DeviceInfo string
	USB        string
	Display    DisplayParams
	Release    int
	SDK        int
	ABI        string
	EGLVersion string
	State      DeviceState
}

// NewDevice creates a new Device from an adb.Device.
func NewDevice(client *Client, device *adb.Device) (*Device, error) {
	deviceInfo, err := device.DeviceInfo()
	if err != nil {
		return nil, err
	}

	deviceState, err := device.State()
	if err != nil {
		return nil, err
	}

	sRelease, _ := getProp(device, "ro.build.version.release")
	iRelease, _ := strconv.ParseInt(sRelease, 10, 64)

	sSdk, _ := getProp(device, "ro.build.version.sdk")
	iSdk, _ := strconv.ParseInt(sSdk, 10, 64)

	sABI, _ := getProp(device, "ro.product.cpu.abi")

	sEGLVersion, _ := getProp(device, "ro.hardware.egl")

	sSize, _ := wm(device, "size")
	aSize := strings.Split(strings.Trim(sSize, " \n"), "x")

	iWidth, _ := strconv.ParseInt(aSize[0], 10, 64)
	iHeight, _ := strconv.ParseInt(aSize[1], 10, 64)

	sDensity, _ := wm(device, "density")
	iDensity, _ := strconv.ParseInt(strings.Trim(sDensity, " \n"), 10, 64)

	return &Device{
		Serial:     deviceInfo.Serial,
		Product:    deviceInfo.Product,
		Model:      deviceInfo.Model,
		DeviceInfo: deviceInfo.DeviceInfo,
		USB:        deviceInfo.Usb,
		State:      DeviceState(deviceState),
		Release:    int(iRelease),
		SDK:        int(iSdk),
		ABI:        sABI,
		EGLVersion: sEGLVersion,
		Display: DisplayParams{
			Width:   int(iWidth),
			Height:  int(iHeight),
			Density: int(iDensity),
		},
	}, nil
}

// SetState sets the state of the device.
func (d *Device) SetState(deviceState DeviceState) {
	d.State = deviceState
}

func (d *Device) String() string {
	return fmt.Sprintf("%s (%s)", d.Serial, d.Model)
}

// DeviceStateChangedEvent represents a device state transition.
type DeviceStateChangedEvent struct {
	Serial string
	State  DeviceState
}

// NewDeviceStateChangedEvent creates a new DeviceStateChangedEvent from an adb.DeviceStateChangedEvent.
func NewDeviceStateChangedEvent(event adb.DeviceStateChangedEvent) DeviceStateChangedEvent {
	return DeviceStateChangedEvent{
		Serial: event.Serial,
		State:  DeviceState(event.NewState),
	}
}
