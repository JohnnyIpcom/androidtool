package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

func Send(client *adbclient.Client, device *adbclient.Device, parent fyne.Window) {
	sendEntry := widget.NewEntry()
	sendEntry.SetPlaceHolder("Enter text or absolute url")

	sendTextButton := widget.NewButtonWithIcon("Send text", assets.SendIcon, nil)
	sendLinkButton := widget.NewButtonWithIcon("Send link", assets.SendIcon, nil)

	d := dialog.NewCustom(
		"Send",
		"Cancel",
		container.NewVBox(
			sendEntry,
			container.NewCenter(
				container.NewHBox(
					sendTextButton,
					sendLinkButton,
				),
			),
		),
		parent,
	)

	sendTextButton.OnTapped = func() {
		text := sendEntry.Text
		if text == "" {
			return
		}

		sendEntry.SetText("")

		if err := client.Input(
			device,
			adbclient.InputSourceDefault,
			adbclient.InputCommandText,
			text,
		); err != nil {
			GetApp().ShowError(err, d.Hide, parent)
		}
	}

	sendLinkButton.OnTapped = func() {
		link := sendEntry.Text
		if link == "" {
			return
		}

		sendEntry.SetText("")
		if err := client.SendLink(device, link); err != nil {
			GetApp().ShowError(err, d.Hide, parent)
		}
	}

	d.Resize(fyne.NewSize(400, 200))
	d.Show()
}
