package adbclient

import (
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

// Device is a wrapper around an adb.Device.
type Device struct {
	Serial string
	Model  string
	State  DeviceState
}

// NewDevice creates a new Device from an adb.Device.
func NewDevice(device *adb.Device) (*Device, error) {
	deviceInfo, err := device.DeviceInfo()
	if err != nil {
		return nil, err
	}

	deviceState, err := device.State()
	if err != nil {
		return nil, err
	}

	return &Device{
		Serial: deviceInfo.Serial,
		Model:  deviceInfo.Model,
		State:  DeviceState(deviceState),
	}, nil
}

// SetState sets the state of the device.
func (d *Device) SetState(deviceState DeviceState) {
	d.State = deviceState
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
