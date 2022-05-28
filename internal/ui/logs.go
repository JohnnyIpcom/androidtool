package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/internal/ui/util"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

const (
	// DefaultLogcatLogPath is the default log file path.
	DefaultLogcatLogPath = "./logcat.log"
)

func Logs(client *adbclient.Client, device *adbclient.Device, parent fyne.Window) {
	logsPathEntry := widget.NewEntry()
	logsPathEntry.SetText(DefaultLogcatLogPath)

	rect := canvas.NewRectangle(theme.BackgroundColor())
	rect.SetMinSize(fyne.NewSize(310, 0))

	logsPathButton := widget.NewButtonWithIcon("Select", theme.DocumentSaveIcon(), func() {
		fsaveDialog := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				return
			}

			if file == nil {
				return
			}

			defer file.Close()
			logsPathEntry.SetText(file.URI().Path())
		}, parent)

		fsaveDialog.SetFileName("logcat.log")
		fsaveDialog.Resize(DialogSize(parent))
		fsaveDialog.Show()
	})

	logsStartButton := widget.NewButtonWithIcon("Start", assets.LogsIcon, nil)
	logsStopButton := widget.NewButtonWithIcon("Stop", theme.MediaStopIcon(), nil)
	logsStopButton.Disable()

	dialog := dialog.NewCustom(
		"Logs",
		"Close",
		container.NewMax(
			rect,
			container.NewVBox(
				widget.NewCard(
					"",
					"",
					container.New(&alignToRightLayout{}, logsPathEntry, logsPathButton),
				),
				container.NewCenter(
					container.NewHBox(
						logsStartButton,
						logsStopButton,
					),
				),
			),
		),
		parent,
	)

	logcat, err := client.Logcat(device)
	if err != nil {
		ShowError(err, nil, parent)
		return
	}

	dialog.SetOnClosed(func() {
		logcat.Close()
	})

	fileSaver, err := util.NewFileSaver(logcat)
	if err != nil {
		ShowError(err, nil, parent)
		return
	}

	logsStartButton.OnTapped = func() {
		if err := client.ClearLogcat(device); err != nil {
			ShowError(err, nil, parent)
			return
		}

		if err := fileSaver.Start(logsPathEntry.Text); err != nil {
			ShowError(err, nil, parent)
			return
		}

		logsStartButton.Disable()
		logsStopButton.Enable()
	}

	logsStopButton.OnTapped = func() {
		logsStartButton.Enable()
		logsStopButton.Disable()
		fileSaver.Stop()
	}

	dialog.Show()
}
