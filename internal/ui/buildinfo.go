package ui

import (
	"image"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/apk"
)

func BuildInfoAPK(pkg *apk.APK, icon image.Image, parent fyne.Window) {
	rows := [][]string{
		{"Identifier", pkg.Identifier()},
		{"Label", pkg.Label()},
		{"Version name", pkg.VersionName()},
		{"Version code", strconv.FormatInt(int64(pkg.VersionCode()), 10)},
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

	table.SetColumnWidth(0, 120)
	table.SetColumnWidth(1, 180)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(310, 200))

	image := canvas.NewImageFromImage(icon)
	image.SetMinSize(fyne.NewSize(192, 192))

	d := dialog.NewCustom(
		"APK Info",
		"OK",
		container.NewVBox(
			container.NewCenter(
				image,
			),
			container.NewMax(
				rect,
				table,
			),
		),
		parent,
	)
	d.Show()
}
