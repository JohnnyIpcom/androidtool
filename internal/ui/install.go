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
	"github.com/johnnyipcom/androidtool/pkg/generic"
)

var ErrorsMap = map[string]string{
	"INSTALL_FAILED_UPDATE_INCOMPATIBLE":             "The app is already installed. If you want to update, make sure that you use the same keystore as in the previous version",
	"INSTALL_FAILED_DUPLICATE_PACKAGE":               "The version code of the installed application is higher than the version code of the application that you are installing",
	"INSTALL_FAILED_INSUFFICIENT_STORAGE":            "Not enough free space on the connected device",
	"INSTALL_FAILED_USER_RESTRICTED":                 "Run the installation again and confirm the installation on the device screen",
	"INSTALL_PARSE_FAILED_INCONSISTENT_CERTIFICATES": "Try to install with keystore",
	"INSTALL_FAILED_OLDER_SDK":                       "Device OS Version is not supported. Check Manifest File",
	"INSTALL_PARSE_FAILED_NO_CERTIFICATES":           "Missing file build.keystore",
}

func humanizeError(out string) string {
	errorsMap := generic.NewMapFromMap(ErrorsMap)

	result := out
	errorsMap.Each(func(key, value string) bool {
		if strings.Contains(out, key) {
			result = value
			return false
		}

		return true
	})

	return result
}

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
		bar.SetText("Failed")
		ShowError(err, d.Hide, parent)
	}

	device, err := client.GetDevice(serial)
	if err != nil {
		onError(err)
		return
	}

	apkPath := client.GetInstallPath()
	if err := client.Upload(ctx, device, file.URI().Path(), apkPath, bar.WithUploadProgress()); err != nil {
		onError(err)
		return
	}

	defer func() {
		if err := client.RemoveFile(device, apkPath); err != nil {
			ShowError(err, nil, parent)
		}
	}()

	bar.SetText("Installing APK...")

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
		onError(humanizeError(string(out)), err)
		return
	}

	d.Hide()
	ShowInformation("Installation result", "Success!", parent)
}
