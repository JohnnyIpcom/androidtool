package ui

import (
	"context"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

func Install(client *adbclient.Client, file fyne.URIReadCloser, parent fyne.Window) {
	bar := NewProgressBar(parent)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Preparing file to install", "Cancel", container.NewMax(rect, bar), parent)
	d.Show()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	onError := func(err error) {
		bar.Failed()

		de := dialog.NewError(err, parent)
		de.SetOnClosed(d.Hide)
		de.Show()
	}

	device, err := client.GetAnyOnlineDevice()
	if err != nil {
		onError(err)
		return
	}

	apkPath := "/data/local/tmp/install.apk"
	if err := client.Upload(ctx, device, file.URI().Path(), apkPath, bar.WithProgress()); err != nil {
		onError(err)
		return
	}

	bar.Install()
	result, err := client.Install(ctx, device, apkPath)

	if err != nil {
		onError(err)
		return
	}

	d.Hide()
	dialog.ShowInformation("Installation result", result, parent)
}
