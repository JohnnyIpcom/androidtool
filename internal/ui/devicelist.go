package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/internal/storage"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/generic"
)

// DeviceItem is a single device item
type DeviceItem struct {
	*adbclient.Device

	check      *widget.Check
	logs       *widget.Button
	screenshot *widget.Button
	video      *widget.Button
	send       *widget.Button
	delete     *widget.Button
}

// DeviceList is a list of devices
type DeviceList struct {
	widget.List

	client   *adbclient.Client
	storage  *storage.Storage
	items    *generic.Slice[*DeviceItem]
	selected *DeviceItem
	parent   fyne.Window
}

// Length returns the number of items in the list
func (d *DeviceList) Length() int {
	return d.items.Len()
}

// CreateItem creates a new empty entry in the list
func (d *DeviceList) CreateItem() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		container.NewHBox(
			widget.NewCheck("", nil),
			widget.NewIcon(assets.StatusIcons["invalid"]),
			widget.NewLabel("<SERIAL>"),
		),
		container.NewHBox(
			widget.NewLabel("INVALID"),
			widget.NewButtonWithIcon("", assets.LogsIcon, nil),
			widget.NewButtonWithIcon("", assets.ScreenshotIcon, nil),
			widget.NewButtonWithIcon("", assets.VideoIcon, nil),
			widget.NewButtonWithIcon("", assets.SendIcon, nil),
			widget.NewButtonWithIcon("", assets.DeleteIcon, nil),
		),
	)
}

// UpdateItem updates the item in the list
func (d *DeviceList) UpdateItem(id int, item fyne.CanvasObject) {
	container := item.(*fyne.Container)

	deviceItem := d.items.Load(id)
	deviceName := fmt.Sprintf("%s (%s)", strings.ToUpper(deviceItem.Serial), deviceItem.Model)
	container.Objects[0].(*fyne.Container).Objects[1].(*widget.Icon).SetResource(assets.StatusIcons[deviceItem.State.String()])
	container.Objects[0].(*fyne.Container).Objects[2].(*widget.Label).SetText(deviceName)
	container.Objects[1].(*fyne.Container).Objects[0].(*widget.Label).SetText(strings.ToUpper(deviceItem.State.String()))
	deviceItem.check = container.Objects[0].(*fyne.Container).Objects[0].(*widget.Check)
	deviceItem.check.OnChanged = func(checked bool) {
		d.OnCheckChanged(id, checked)
	}

	deviceItem.logs = container.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)
	deviceItem.logs.OnTapped = func() {
		go Logs(d.client, deviceItem.Device, d.parent)
	}

	deviceItem.screenshot = container.Objects[1].(*fyne.Container).Objects[2].(*widget.Button)
	deviceItem.screenshot.OnTapped = func() {
		go Screenshot(d.client, deviceItem.Device, d.parent)
	}

	deviceItem.video = container.Objects[1].(*fyne.Container).Objects[3].(*widget.Button)
	deviceItem.video.OnTapped = func() {
		go Video(d.client, deviceItem.Device, d.parent)
	}

	deviceItem.send = container.Objects[1].(*fyne.Container).Objects[4].(*widget.Button)
	deviceItem.send.OnTapped = func() {
		go Send(d.client, deviceItem.Device, d.parent)
	}

	deviceItem.delete = container.Objects[1].(*fyne.Container).Objects[5].(*widget.Button)
	deviceItem.delete.OnTapped = func() {
		d.OnDelete(id)
	}

	if deviceItem.Device.State == adbclient.StateOnline {
		deviceItem.logs.Enable()
		deviceItem.screenshot.Enable()
		deviceItem.video.Enable()
		deviceItem.send.Enable()
	} else {
		deviceItem.logs.Disable()
		deviceItem.screenshot.Disable()
		deviceItem.video.Disable()
		deviceItem.send.Disable()
	}

	// If no device is selected, select the first one
	if d.selected == nil {
		deviceItem.check.SetChecked(true)
	}

	// If the currently selected device is not online and
	// this device is online, select this device
	if d.selected != nil && d.selected.State != adbclient.StateOnline && deviceItem.State == adbclient.StateOnline {
		deviceItem.check.SetChecked(true)
	}
}

// OnSelected is called when the user selects an item
func (d *DeviceList) OnSelected(id int) {
	go DeviceInfo(d.client, d.items.Load(id).Device, d.parent)
	d.Unselect(id)
}

// OnDelete is called when the user deletes an item
func (d *DeviceList) OnDelete(id int) {
	deviceItem := d.items.Load(id)
	if deviceItem.check.Checked {
		d.selected = nil

		// Find another online device to select
		d.items.Each(func(i int, item *DeviceItem) bool {
			if item.Device.State == adbclient.StateOnline && i != id {
				item.check.SetChecked(true)
				return false
			}

			return true
		})

		// If no online devices, select first device
		if d.selected == nil {
			d.items.Each(func(i int, item *DeviceItem) bool {
				if i != id {
					item.check.SetChecked(true)
					return false
				}

				return true
			})
		}
	}

	d.storage.DeleteDevice(deviceItem.Serial)
	d.items.Delete(id)
	d.Refresh()
}

// OnCheckChanged is called when the user checks or unchecks a device
func (d *DeviceList) OnCheckChanged(id int, checked bool) {
	deviceItem := d.items.Load(id)
	if checked {
		// Unselect all other items
		d.items.Each(func(i int, item *DeviceItem) bool {
			if item.Serial != deviceItem.Serial {
				if item.check != nil {
					// can't use item.check.SetChecked(false) because it will trigger OnCheckChanged again
					item.check.Checked = false
					item.check.Refresh()
				}
			}

			return true
		})

		d.selected = deviceItem
	} else {
		// reselect current device, one and only one device must be selected
		deviceItem.check.Checked = true
		deviceItem.check.Refresh()
	}
}

// deviceWatcher watches for device changes
func (d *DeviceList) deviceWatcher() {
	for event := range d.client.DeviceWatcher() {
		var oldItem *DeviceItem
		d.items.Each(func(i int, item *DeviceItem) bool {
			if item.Serial == event.Serial {
				oldItem = item
				return false
			}
			return true
		})

		// if item is not found, create new item
		if oldItem == nil {
			newDevice, err := d.client.GetDevice(event.Serial)
			if err != nil {
				continue
			}

			d.storage.SaveDevice(newDevice)
			d.items.Store(
				&DeviceItem{
					Device: newDevice,
				},
			)
			d.Refresh()
			continue
		}

		if oldItem.Device.State == adbclient.StateInvalid {
			// if device is invalid, refresh it
			oldItem.Device, _ = d.client.GetDevice(event.Serial)
			d.storage.SaveDevice(oldItem.Device)
		}

		oldItem.SetState(event.State)
		d.Refresh()
	}
}

// SelectDevice selects a device
func (d *DeviceList) SelectedDevice() (*adbclient.Device, error) {
	if d.selected == nil {
		return nil, fmt.Errorf("no device selected")
	}

	if d.selected.State != adbclient.StateOnline {
		return nil, fmt.Errorf("device is not online")
	}

	return d.selected.Device, nil
}

// NewDeviceList creates a new device list
func NewDeviceList(client *adbclient.Client, storage *storage.Storage, parent fyne.Window) (*DeviceList, error) {
	d := &DeviceList{
		parent:  parent,
		client:  client,
		storage: storage,
		items:   generic.NewSlice[*DeviceItem](),
	}

	d.List.Length = d.Length
	d.List.CreateItem = d.CreateItem
	d.List.UpdateItem = d.UpdateItem
	d.List.OnSelected = d.OnSelected
	d.ExtendBaseWidget(d)

	devices, err := storage.GetDevices()
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		d.items.Store(
			&DeviceItem{
				Device: device,
			},
		)
	}

	go d.deviceWatcher()
	return d, nil
}
