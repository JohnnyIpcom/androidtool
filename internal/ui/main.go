package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

type main struct {
	app    fyne.App
	parent fyne.Window
	client *adbclient.Client
}

func uiMain(app fyne.App, parent fyne.Window, client *adbclient.Client) *main {
	return &main{
		app:    app,
		parent: parent,
		client: client,
	}
}

func (m *main) buildUI() *fyne.Container {
	deviceList := NewDeviceList(m.parent, m.client)
	deviceList.Resize(fyne.NewSize(477, 200))

	installButton := widget.NewButton("Install", func() {
		fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, m.parent)
				return
			}

			if file == nil {
				return
			}

			defer file.Close()
			if err := m.client.Install(file); err != nil {
				dialog.ShowError(err, m.parent)
				return
			}
		}, m.parent)

		fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".apk", ".aab"}))
		fopenDialog.Show()
	})

	return container.NewVBox(
		widget.NewLabel("Devices:"),
		deviceList,
		installButton,
	)
}

func (m *main) tabItem() *container.TabItem {
	return &container.TabItem{Text: "Main", Icon: theme.MailSendIcon(), Content: m.buildUI()}
}
