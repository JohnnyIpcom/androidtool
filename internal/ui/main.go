package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
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

	installButton := widget.NewButton("Install", m.onInstall)
	installButton.SetIcon(assets.InstallIcon)

	return container.NewVBox(
		widget.NewLabel("Devices:"),
		deviceList,
		installButton,
	)
}

func (m *main) onInstall() {
	fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, m.parent)
			return
		}

		if file == nil {
			return
		}

		switch file.URI().Extension() {
		case "apk":
			go InstallAPK(m.client, file, m.parent)
		case "aab":
			go InstallAAB(m.client, file, m.parent)
		default:
			go func() {
				defer file.Close()
				dialog.ShowError(fmt.Errorf("unsupported file type: %s", file.URI().Extension()), m.parent)
			}()
		}
	}, m.parent)

	fopenDialog.Resize(DialogSize(m.parent))
	fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".apk", ".aab"}))
	fopenDialog.Show()
}

func (m *main) tabItem() *container.TabItem {
	return &container.TabItem{Text: "Main", Icon: assets.MainTabIcon, Content: m.buildUI()}
}
