package ui

import (
	"context"
	"fmt"
	"image/color"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

// InstallAPK installs an APK file to a device.
func InstallAPK(client *adbclient.Client, serial string, file fyne.URIReadCloser, parent fyne.Window) {
	bar := NewProgressBar(parent)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Installation", "Cancel", container.NewMax(rect, bar), parent)
	d.Show()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	onError := func(err error) {
		bar.Failed()
		ShowError(err, d.Hide, parent)
	}

	device, err := client.GetDevice(serial)
	if err != nil {
		onError(err)
		return
	}

	apkPath := client.GetInstallPath()
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
	ShowInformation("Installation result", result, parent)
}

// InstallAAB installs an AAB file to a device and optionally signs it with a keystore.
func InstallAAB(client *aabclient.Client, serial string, file fyne.URIReadCloser, keystore *aabclient.KeystoreConfig, parent fyne.Window) {
	pbari := widget.NewProgressBarInfinite()
	label := widget.NewLabel("Build APKs...")
	label.Alignment = fyne.TextAlignCenter

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Installation", "Cancel", container.NewMax(rect, pbari, label), parent)
	d.Show()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	onError := func(out string, err error) {
		label.SetText("Failed")
		ShowError(fmt.Errorf("%s\n%s", err.Error(), out), d.Hide, parent)
	}

	dir, fullName := filepath.Split(file.URI().Path())
	nameWithoutExt := strings.TrimSuffix(fullName, filepath.Ext(fullName))

	apksFile := filepath.Join(dir, nameWithoutExt+".apks")
	out, err := client.BuildAPKs(ctx, file.URI().Path(), apksFile, serial, keystore)

	if err != nil {
		onError(string(out), err)
		return
	}

	label.SetText("Installing APKs...")
	out, err = client.InstallAPKs(ctx, apksFile, serial)
	if err != nil {
		onError(string(out), err)
		return
	}

	d.Hide()
	ShowInformation("Installation result", "Success!", parent)
}
