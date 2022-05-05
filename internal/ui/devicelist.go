package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/generic"
)

type DeviceItem struct {
	device *adbclient.Device
}

type DeviceList struct {
	widget.List

	client *adbclient.Client
	items  *generic.Slice[*DeviceItem]
	window fyne.Window
}

func (d *DeviceList) Length() int {
	return d.items.Len()
}

func (d *DeviceList) CreateItem() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		nil,
		container.NewHBox(
			widget.NewIcon(assets.StatusIcons["invalid"]),
			widget.NewLabel(""),
		),
		widget.NewLabel(""),
	)
}

func (d *DeviceList) UpdateItem(id int, item fyne.CanvasObject) {
	container := item.(*fyne.Container)

	device := d.items.Load(id).device
	container.Objects[0].(*fyne.Container).Objects[0].(*widget.Icon).SetResource(assets.StatusIcons[device.State.String()])
	container.Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(device.Serial)
	container.Objects[1].(*widget.Label).SetText(strings.ToUpper(device.State.String()))
}

func (d *DeviceList) OnSelected(id int) {
	d.Unselect(id)
}

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

func NewDeviceList(window fyne.Window, client *adbclient.Client) *DeviceList {
	d := &DeviceList{
		window: window,
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
