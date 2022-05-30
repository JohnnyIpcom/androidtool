package ui

import (
	"context"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
)

func APKManifest(aapt *aapt.AAPT, path string, parent fyne.Window) {
	manifestPathEntry := widget.NewEntry()
	manifestPathEntry.SetText("./manifest.txt")

	manifestPathButton := widget.NewButtonWithIcon("Select", theme.DocumentSaveIcon(), func() {
		fsaveDialog := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				return
			}

			if file == nil {
				return
			}

			defer file.Close()
			manifestPathEntry.SetText(file.URI().Path())
		}, parent)

		fsaveDialog.SetFileName("manifest.txt")
		fsaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
		fsaveDialog.Resize(DialogSize(parent))
		fsaveDialog.Show()
	})

	d := dialog.NewCustomConfirm(
		"Manifest",
		"Save",
		"Close",
		container.New(&alignToRightLayout{}, manifestPathEntry, manifestPathButton),
		func(b bool) {
			if !b || manifestPathEntry.Text == "" {
				return
			}

			// Save manifest to file.
			file, err := os.OpenFile(manifestPathEntry.Text, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				ShowError(err, nil, parent)
				return
			}

			manifest, err := aapt.GetManifest(path)
			if err != nil {
				ShowError(err, nil, parent)
				return
			}

			defer file.Close()
			if _, err := file.Write(manifest); err != nil {
				ShowError(err, nil, parent)
				return
			}
		},
		parent,
	)

	d.Resize(fyne.NewSize(400, 0))
	d.Show()
}

func AABManifest(client *aabclient.Client, aabFile string, parent fyne.Window) {
	manifestPathEntry := widget.NewEntry()
	manifestPathEntry.SetText("./AndroidManifest.xml")

	manifestPathButton := widget.NewButtonWithIcon("Select", theme.DocumentSaveIcon(), func() {
		fsaveDialog := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				return
			}

			if file == nil {
				return
			}

			defer file.Close()
			manifestPathEntry.SetText(file.URI().Path())
		}, parent)

		fsaveDialog.SetFileName("AndroidManifest.xml")
		fsaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".xml"}))
		fsaveDialog.Resize(DialogSize(parent))
		fsaveDialog.Show()
	})

	d := dialog.NewCustomConfirm(
		"Manifest",
		"Save",
		"Close",
		container.New(&alignToRightLayout{}, manifestPathEntry, manifestPathButton),
		func(b bool) {
			if !b || manifestPathEntry.Text == "" {
				return
			}

			// Save manifest to file.
			file, err := os.OpenFile(manifestPathEntry.Text, os.O_WRONLY|os.O_CREATE, 0644)
			if err != nil {
				ShowError(err, nil, parent)
				return
			}

			manifest, err := client.GetManifest(context.Background(), aabFile)
			if err != nil {
				ShowError(err, nil, parent)
				return
			}

			defer file.Close()
			if _, err := file.Write(manifest); err != nil {
				ShowError(err, nil, parent)
				return
			}
		},
		parent,
	)

	d.Resize(fyne.NewSize(400, 0))
	d.Show()
}
