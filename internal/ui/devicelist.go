package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/generic"
)

// DeviceItem is a single device item
type DeviceItem struct {
	device     *adbclient.Device
	check      *widget.Check
	logs       *widget.Button
	screenshot *widget.Button
}

// DeviceList is a list of devices
type DeviceList struct {
	widget.List

	client   *adbclient.Client
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
		),
	)
}

// UpdateItem updates the item in the list
func (d *DeviceList) UpdateItem(id int, item fyne.CanvasObject) {
	container := item.(*fyne.Container)

	deviceItem := d.items.Load(id)
	container.Objects[0].(*fyne.Container).Objects[1].(*widget.Icon).SetResource(assets.StatusIcons[deviceItem.device.State.String()])
	container.Objects[0].(*fyne.Container).Objects[2].(*widget.Label).SetText(deviceItem.device.String())
	container.Objects[1].(*fyne.Container).Objects[0].(*widget.Label).SetText(strings.ToUpper(deviceItem.device.State.String()))
	deviceItem.check = container.Objects[0].(*fyne.Container).Objects[0].(*widget.Check)
	deviceItem.check.OnChanged = func(checked bool) {
		d.OnCheckChanged(id, checked)
	}

	deviceItem.logs = container.Objects[1].(*fyne.Container).Objects[1].(*widget.Button)
	deviceItem.logs.OnTapped = func() {
		go Logs(d.client, deviceItem.device, d.parent)
	}

	deviceItem.screenshot = container.Objects[1].(*fyne.Container).Objects[2].(*widget.Button)
	deviceItem.screenshot.OnTapped = func() {
		go Screenshot(d.client, deviceItem.device, d.parent)
	}

	if deviceItem.device.State == adbclient.StateOnline {
		deviceItem.logs.Enable()
		deviceItem.screenshot.Enable()
	} else {
		deviceItem.logs.Disable()
		deviceItem.screenshot.Disable()
	}

	// Auto-select first device
	if d.selected == nil {
		deviceItem.check.SetChecked(true)
	}
}

// OnSelected is called when the user selects an item
func (d *DeviceList) OnSelected(id int) {
	go DeviceInfo(d.client, d.items.Load(id).device, d.parent)
	d.Unselect(id)
}

// OnCheckChanged is called when the user checks or unchecks a device
func (d *DeviceList) OnCheckChanged(id int, checked bool) {
	deviceItem := d.items.Load(id)
	if checked {
		// Unselect all other items
		d.items.Each(func(i int, item *DeviceItem) bool {
			if item.device.Serial != deviceItem.device.Serial {
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
			if item.device.Serial == event.Serial {
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

			d.items.Store(&DeviceItem{device: newDevice})
			d.Refresh()
			continue
		}

		oldItem.device.SetState(event.State)
		d.Refresh()
	}
}

// SelectDevice selects a device
func (d *DeviceList) SelectedDevice() (*adbclient.Device, error) {
	if d.selected == nil {
		return nil, fmt.Errorf("no device selected")
	}

	if d.selected.device.State != adbclient.StateOnline {
		return nil, fmt.Errorf("device is not online")
	}

	return d.selected.device, nil
}

// NewDeviceList creates a new device list
func NewDeviceList(client *adbclient.Client, parent fyne.Window) *DeviceList {
	d := &DeviceList{
		parent: parent,
		client: client,
		items:  generic.NewSlice[*DeviceItem](),
	}

	d.List.Length = d.Length
	d.List.CreateItem = d.CreateItem
	d.List.UpdateItem = d.UpdateItem
	d.List.OnSelected = d.OnSelected
	d.ExtendBaseWidget(d)

	go d.deviceWatcher()
	return d
}
