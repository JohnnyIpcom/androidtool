package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

const (
	// DefaultVideoPath is the default video file path.
	DefaultVideoPath = "./video.mp4"
)

func Video(client *adbclient.Client, device *adbclient.Device, parent fyne.Window) {
	progressBar := NewProgressBar(parent)

	videoPathEntry := widget.NewEntry()
	videoPathEntry.SetText(DefaultVideoPath)

	videoPathButton := widget.NewButton("Select", func() {
		fsaveDialog := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				return
			}

			if file == nil {
				return
			}

			defer file.Close()
			videoPathEntry.SetText(file.URI().Path())
		}, parent)

		fsaveDialog.SetFileName("video.mp4")
		fsaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp4"}))
		fsaveDialog.Resize(DialogSize(parent))
		fsaveDialog.Show()
	})

	videoDurationEntry := widget.NewEntry()
	videoDurationEntry.SetText("1m30s")
	videoDurationEntry.Validator = func(s string) error {
		if s == "" {
			return nil
		}

		duration, err := time.ParseDuration(s)
		if err != nil {
			return err
		}

		if duration < time.Second || duration > 3*time.Minute {
			return fmt.Errorf("duration must be between 1s and 3m")
		}

		return nil
	}

	makeVideoButton := widget.NewButtonWithIcon("Video", assets.VideoIcon, nil)

	d := dialog.NewCustom(
		"Video",
		"Close",
		container.NewBorder(
			container.New(&alignToRightLayout{}, videoPathEntry, videoPathButton),
			container.NewMax(
				container.NewBorder(
					progressBar,
					nil,
					widget.NewLabel("Duration:"),
					makeVideoButton,
					videoDurationEntry,
				),
			),
			nil,
			nil,
			container.NewCenter(
				widget.NewLabel("Video player will be available in one of the next versions."),
			),
		),
		parent,
	)

	onError := func(err error) {
		progressBar.SetText("Failed")
		GetApp().ShowError(err, d.Hide, parent)
	}

	makeVideoButton.OnTapped = func() {
		makeVideoButton.Disable()
		defer makeVideoButton.Enable()

		progressBar.SetText("Recording...")

		duration, err := time.ParseDuration(videoDurationEntry.Text)
		if err != nil {
			onError(err)
			return
		}

		width, height := device.Display.Width, device.Display.Height
		if width < height {
			width, height = height, width
		}

		videoPath := client.GetVideoPath()
		if err := client.Video(device, width, height, videoPath, adbclient.WithVideoDuration(duration)); err != nil {
			onError(err)
			return
		}

		defer func() {
			if err := client.RemoveFile(device, videoPath); err != nil {
				onError(err)
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		progressBar.SetText("")
		if err := client.DownloadFile(ctx, device, videoPath, videoPathEntry.Text, progressBar.WithDownloadProgress()); err != nil {
			onError(err)
			return
		}

		progressBar.SetText("Done")
	}

	d.Resize(DialogSize(parent))
	d.Show()
}
