package ui

import (
	"context"
	"fmt"
	"image/color"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"github.com/johnnyipcom/androidtool/internal/ui/util"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
	"github.com/johnnyipcom/androidtool/pkg/generic"
)

type AABInfo struct {
	APKsPath string
	ABIList  []string
	MinSize  uint64
	MaxSize  uint64
}

func LoadAAB(client *aabclient.Client, aapt *aapt.AAPT, path string, useCachedData bool, parent fyne.Window) (*AABInfo, error) {
	bar := NewProgressBar(parent)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Loading info", "Cancel", container.NewMax(rect, bar, rect), parent)
	d.Show()

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	onError := func(out string, err error) {
		bar.SetText("Failed")
		ShowError(fmt.Errorf("%s\n%s", err.Error(), out), d.Hide, parent)
	}

	dir, fullName := filepath.Split(path)
	nameWithoutExt := strings.TrimSuffix(fullName, filepath.Ext(fullName))

	bar.SetText("Build APKs...")
	apksFile := filepath.Join(dir, nameWithoutExt+".apks")

	if !useCachedData || !IsFileExists(apksFile) {
		// unpack aab to apks
		out, err := client.BuildAPKs(ctx, path, apksFile, "", nil)
		if err != nil {
			onError(string(out), err)
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

	bar.SetText("Parsing APKs...")

	abiSet := generic.NewSet[string]()
	if err := filepath.Walk(destDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if strings.HasSuffix(info.Name(), ".apk") {
			abis, err := aapt.GetNativeCodeABIs(path)
			if err != nil {
				return err
			}

			for _, abi := range abis {
				if len(abi) > 0 {
					abiSet.Store(abi)
				}
			}
		}

		return nil
	}); err != nil {
		onError("", err)
		return nil, err
	}

	bar.SetText("Calculating sizes...")
	min, max, err := client.GetMinMaxSizes(ctx, apksFile)
	if err != nil {
		onError("", err)
		return nil, err
	}

	d.Hide()
	return &AABInfo{
		APKsPath: apksFile,
		ABIList:  abiSet.Keys(),
		MinSize:  min,
		MaxSize:  max,
	}, nil
}
