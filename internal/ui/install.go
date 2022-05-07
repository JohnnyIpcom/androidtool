package ui

import (
	"context"
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

func InstallAPK(client *adbclient.Client, file fyne.URIReadCloser, parent fyne.Window) {
	defer file.Close()

	bar := NewProgressBar(parent)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Installation", "Cancel", container.NewMax(rect, bar), parent)
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
	result, err := client.Install(device, apkPath)

	if err != nil {
		onError(err)
		return
	}

	d.Hide()
	dialog.ShowInformation("Installation result", result, parent)
}

func InstallAAB(client *adbclient.Client, file fyne.URIReadCloser, parent fyne.Window) {
	dialog.ShowError(fmt.Errorf("not implemented"), parent)
}
