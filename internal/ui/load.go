package ui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/ui/util"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
	"github.com/johnnyipcom/androidtool/pkg/apk"
)

type APKInfo struct {
	APK  *apk.APK
	Icon image.Image
}

func LoadAPK(path string, parent fyne.Window) (*APKInfo, error) {
	pbari := widget.NewProgressBarInfinite()
	label := widget.NewLabel("Parsing APK...")
	label.Alignment = fyne.TextAlignCenter

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Loading APK info", "Cancel", container.NewMax(rect, pbari, label), parent)
	d.Show()

	onError := func(out string, err error) {
		label.SetText("Failed")
		GetApp().ShowError(fmt.Errorf("%s\n%s", err.Error(), out), d.Hide, parent)
	}

	pkg, err := apk.NewAPK(path)
	if err != nil {
		onError("", err)
		return nil, err
	}

	icon, err := pkg.Icon(apk.ScreenHDPI)
	if err != nil {
		onError("", err)
		return nil, err
	}

	d.Hide()
	return &APKInfo{
		APK:  pkg,
		Icon: icon,
	}, nil
}

type AABInfo struct {
	APKsPath     string
	APKPath      string
	UnpackedPath string
	APK          *apk.APK
	Icon         image.Image
}

func LoadAAB(client *aabclient.Client, aapt *aapt.AAPT, path string, useCachedData bool, parent fyne.Window) (*AABInfo, error) {
	bar := NewProgressBar(parent)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Loading AAB info", "Cancel", container.NewMax(rect, bar), parent)
	d.Show()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	onError := func(out string, err error) {
		bar.SetText("Failed")
		GetApp().ShowError(fmt.Errorf("%s\n%s", err.Error(), out), d.Hide, parent)
	}

	dir, fullName := filepath.Split(path)
	nameWithoutExt := strings.TrimSuffix(fullName, filepath.Ext(fullName))

	bar.SetText("Build APKs...")
	apksFile := filepath.Join(dir, nameWithoutExt+".apks")

	if !useCachedData || !IsFileExists(apksFile) {
		// unpack aab to apks
		out, err := client.BuildAPKs(ctx, path, apksFile, "", false, nil)
		if err != nil {
			onError(string(out), err)
			return nil, err
		}
	}

	bar.SetText("Build universal APK...")
	apkFile := filepath.Join(dir, nameWithoutExt+".apk")

	if !useCachedData || !IsFileExists(apkFile) {
		// build universal apks
		out, err := client.BuildAPKs(ctx, path, filepath.Join(dir, "universal.apks"), "", true, nil)
		if err != nil {
			onError(string(out), err)
			return nil, err
		}

		defer func() {
			if err := os.Remove(filepath.Join(dir, "universal.apks")); err != nil {
				onError("", err)
			}
		}()

		var once sync.Once
		progressFunc := func(current, total uint64) {
			once.Do(func() { bar.Max = float64(total) })
			bar.SetValue(float64(current))
		}

		if err := util.UnzipFile(ctx, filepath.Join(dir, "universal.apks"), "universal.apk", apkFile, progressFunc); err != nil {
			onError("", err)
			return nil, err
		}
	}

	bar.SetText("")

	destDir := filepath.Join(dir, nameWithoutExt)
	if !useCachedData || !IsDirExists(destDir) {
		var once sync.Once
		progressFunc := func(current, total uint64) {
			once.Do(func() { bar.Max = float64(total) })
			bar.SetValue(float64(current))
		}

		if err := util.Unzip(ctx, apksFile, destDir, progressFunc); err != nil {
			onError("", err)
			return nil, err
		}
	}

	pkg, err := apk.NewAPK(apkFile)
	if err != nil {
		onError("", err)
		return nil, err
	}

	icon, err := pkg.Icon(apk.ScreenHDPI)
	if err != nil {
		onError("", err)
		return nil, err
	}

	d.Hide()
	return &AABInfo{
		APKsPath:     apksFile,
		UnpackedPath: destDir,
		APK:          pkg,
		Icon:         icon,
	}, nil
}
