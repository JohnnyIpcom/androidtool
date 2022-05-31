package ui

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

func DeviceInfo(client *adbclient.Client, device *adbclient.Device, parent fyne.Window) {
	rows := [][]string{
		{"Serial", device.Serial},
		{"Product", device.Product},
		{"Model", device.Model},
		{"Device Info", device.DeviceInfo},
		{"USB", device.USB},
		{"Display", device.Display.String()},
		{"Release", device.Release},
		{"SDK", strconv.FormatInt(int64(device.SDK), 10)},
		{"ABI", device.ABI},
		{"EGL Version", device.EGLVersion},
	}

	table := widget.NewTable(
		func() (int, int) {
			return len(rows), 2
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(rows[i.Row][i.Col])
		},
	)
	table.SetColumnWidth(0, 100)
	table.SetColumnWidth(1, 200)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(310, 300))

	d := dialog.NewCustom("Device Info", "OK", container.NewMax(rect, table), parent)
	d.Show()
}
