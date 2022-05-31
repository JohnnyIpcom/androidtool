package ui

import (
	"context"
	"fmt"
	"image/color"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/c2h5oh/datasize"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
)

func showSizes(min uint64, max uint64, parent fyne.Window) {
	minSize := datasize.ByteSize(min)
	maxSize := datasize.ByteSize(max)
	minSizeLabel := widget.NewLabel(fmt.Sprintf("%s (%d)", minSize.HumanReadable(), minSize.Bytes()))
	maxSizeLabel := widget.NewLabel(fmt.Sprintf("%s (%d)", maxSize.HumanReadable(), maxSize.Bytes()))

	dialog.ShowCustom(
		"ABI Info",
		"Close",
		container.NewVBox(
			widget.NewCard(
				"",
				"Sizes:",
				container.NewGridWithColumns(
					2,
					widget.NewLabel("Min:"),
					minSizeLabel,
					widget.NewLabel("Max:"),
					maxSizeLabel,
				),
			),
		),
		parent,
	)
}

func APKSizes(apkPath string, parent fyne.Window) {
	stat, err := os.Stat(apkPath)
	if err != nil {
		GetApp().ShowError(err, nil, parent)
		return
	}

	showSizes(uint64(stat.Size()), uint64(stat.Size()), parent)
}

func AABSizes(client *aabclient.Client, apksPath string, parent fyne.Window) {
	bar := widget.NewProgressBarInfinite()

	label := widget.NewLabel("Parsing APKs...")
	label.Alignment = fyne.TextAlignCenter

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Loading info", "Cancel", container.NewMax(rect, bar, label), parent)
	d.Show()

	label.SetText("Parsing APKs...")

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	onError := func(out string, err error) {
		label.SetText("Failed")
		GetApp().ShowError(fmt.Errorf("%s\n%s", err.Error(), out), d.Hide, parent)
	}

	min, max, err := client.GetMinMaxSizes(ctx, apksPath)
	if err != nil {
		onError(err.Error(), err)
		return
	}

	d.Hide()
	showSizes(min, max, parent)
}
