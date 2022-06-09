package ui

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/c2h5oh/datasize"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

// ZeroReader is a reader that reads zeros.
type zeroReader struct {
}

func (r *zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}

	return len(p), nil
}

func Zeroing(client *adbclient.Client, device *adbclient.Device, parent fyne.Window) {
	zeroingPathEntry := widget.NewEntry()
	zeroingPathEntry.SetText("/sdcard/zeroing%04d.dat")

	freeSpace, _ := client.GetFreeSpace(device)

	zeroingSizeEntry := widget.NewEntry()
	zeroingSizeEntry.SetText(strconv.FormatUint(freeSpace, 10))

	zeroingProgress := NewProgressBar(parent)
	zeroingPathButton := widget.NewButtonWithIcon("Zeroing", assets.ZeroingIcon, nil)

	d := dialog.NewCustom(
		"Zeroing free space",
		"Cancel",
		container.NewVBox(
			container.NewGridWithColumns(
				2,
				widget.NewLabelWithStyle("Path template:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				zeroingPathEntry,
				widget.NewLabelWithStyle("Zeroing size:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				zeroingSizeEntry,
			),
			zeroingProgress,
			container.NewCenter(
				zeroingPathButton,
			),
		),
		parent,
	)

	ctx, cancel := context.WithCancel(context.Background())
	d.SetOnClosed(cancel)

	zeroingPathButton.OnTapped = func() {
		var size datasize.ByteSize
		if err := size.UnmarshalText([]byte(zeroingSizeEntry.Text)); err != nil {
			GetApp().ShowError(err, d.Hide, parent)
			return
		}

		const chunkSize uint64 = 1024 * 1024 * 1024 // 1GB
		numFullChunks := size.Bytes() / chunkSize   // count full size chunks
		lastChunkSize := size.Bytes() % chunkSize   // count last chunk size(can be 0)

		for i := uint64(0); i < numFullChunks+1; i++ {
			path := fmt.Sprintf(zeroingPathEntry.Text, i)

			limit := chunkSize
			if i == numFullChunks {
				limit = lastChunkSize
			}

			zeroingProgress.TextFormatter = func() string {
				delta := float32(zeroingProgress.Max - zeroingProgress.Min)
				ratio := float32(zeroingProgress.Value-zeroingProgress.Min) / delta

				return fmt.Sprintf("(%d/%d) Zeroing %s [%s]", i+1, numFullChunks+1, filepath.Base(path), strconv.Itoa(int(ratio*100))+"%")
			}

			if limit != 0 {
				reader := io.LimitReader(&zeroReader{}, int64(limit))
				if err := client.Upload(ctx, device, reader, limit, path, zeroingProgress.WithUploadProgress()); err != nil {
					GetApp().ShowError(err, d.Hide, parent)
					return
				}
			}
		}

		d.Hide()
	}

	d.Resize(fyne.NewSize(500, 0))
	d.Show()
}
