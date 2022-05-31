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
	Width   int `json:"width"`
	Height  int `json:"height"`
	Density int `json:"density"`
}

func (p DisplayParams) String() string {
	return fmt.Sprintf("%dx%d (%d dpi)", p.Width, p.Height, p.Density)
}

// Device is a wrapper around an adb.Device.
type Device struct {
	Serial     string        `json:"serial"`
	Product    string        `json:"product"`
	Model      string        `json:"model"`
	DeviceInfo string        `json:"device_info"`
	USB        string        `json:"usb"`
	Display    DisplayParams `json:"display"`
	Release    string        `json:"release"`
	SDK        int           `json:"sdk"`
	ABI        string        `json:"abi"`
	EGLVersion string        `json:"egl_version"`
	State      DeviceState   `json:"-"`
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
		Release:    sRelease,
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

// String implements the fmt.Stringer interface.
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
