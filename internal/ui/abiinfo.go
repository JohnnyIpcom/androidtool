package ui

import (
	"fmt"
	"image/color"
	"io/fs"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
	"github.com/johnnyipcom/androidtool/pkg/generic"
)

func showABIInfo(abis []string, parent fyne.Window) {
	data := binding.BindStringList(&abis)
	list := widget.NewListWithData(
		data,
		func() fyne.CanvasObject {
			return widget.NewLabel("<ABI>")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			o.(*widget.Label).Bind(i.(binding.String))
		},
	)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(310, 200))

	dialog.ShowCustom(
		"ABI Info",
		"Close",
		container.NewVBox(
			widget.NewCard(
				"",
				"ABIs:",
				container.NewMax(
					rect,
					list,
				),
			),
		),
		parent,
	)
}

func APKABIInfo(aapt *aapt.AAPT, apkPath string, parent fyne.Window) {
	abis, err := aapt.GetNativeCodeABIs(apkPath)
	if err != nil {
		ShowError(err, nil, parent)
		return
	}

	showABIInfo(abis, parent)
}

func AABABIInfo(client *aabclient.Client, aapt *aapt.AAPT, unpackedPath string, parent fyne.Window) {
	bar := widget.NewProgressBarInfinite()

	label := widget.NewLabel("Parsing APKs...")
	label.Alignment = fyne.TextAlignCenter

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Loading info", "Cancel", container.NewMax(rect, bar, label), parent)
	d.Show()

	label.SetText("Parsing APKs...")

	onError := func(out string, err error) {
		label.SetText("Failed")
		ShowError(fmt.Errorf("%s\n%s", err.Error(), out), d.Hide, parent)
	}

	abiSet := generic.NewSet[string]()
	if err := filepath.Walk(unpackedPath, func(path string, info fs.FileInfo, err error) error {
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
		return
	}

	d.Hide()
	showABIInfo(abiSet.Keys(), parent)
}
