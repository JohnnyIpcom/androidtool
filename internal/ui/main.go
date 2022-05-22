package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

type main struct {
	app       fyne.App
	parent    fyne.Window
	adbclient *adbclient.Client
	aabclient *aabclient.Client

	deviceList             *DeviceList
	useCustomKeystoreCheck *widget.Check
}

func uiMain(app fyne.App, parent fyne.Window, adbclient *adbclient.Client, aabClient *aabclient.Client) *main {
	return &main{
		app:       app,
		parent:    parent,
		adbclient: adbclient,
		aabclient: aabClient,
	}
}

func (m *main) buildUI() *fyne.Container {
	m.deviceList = NewDeviceList(m.adbclient, m.parent)

	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(0, 75))

	installAPKButton := widget.NewButton("Install *.apk", m.onInstallAPK)
	installAPKButton.SetIcon(assets.InstallIcon)

	m.useCustomKeystoreCheck = widget.NewCheck("Use custom keystore", m.onUseCustomKeystoreChecked)

	installAABButton := widget.NewButton("Install *.aab", m.onInstallAAB)
	installAABButton.SetIcon(assets.InstallIcon)

	return container.NewBorder(
		nil,
		container.NewGridWithColumns(
			2,
			widget.NewCard(
				"",
				"",
				container.NewVBox(
					layout.NewSpacer(),
					installAPKButton,
				),
			),
			widget.NewCard(
				"",
				"",
				container.NewVBox(
					m.useCustomKeystoreCheck,
					installAABButton,
				),
			),
		),
		nil,
		nil,
		container.NewMax(
			widget.NewCard(
				"",
				"Devices:",
				m.deviceList,
			),
		),
	)
}

func (m *main) onUseCustomKeystoreChecked(checked bool) {
}

func (m *main) onInstallAPK() {
	fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			ShowError(err, nil, m.parent)
			return
		}

		if file == nil {
			return
		}

		device, err := m.deviceList.SelectedDevice()
		if err != nil {
			ShowError(err, nil, m.parent)
			return
		}

		go func() {
			defer file.Close()
			InstallAPK(m.adbclient, device.Serial, file, m.parent)
		}()
	}, m.parent)

	fopenDialog.Resize(DialogSize(m.parent))
	fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".apk"}))
	fopenDialog.Show()
}

func (m *main) onInstallAAB() {
	fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			ShowError(err, nil, m.parent)
			return
		}

		if file == nil {
			return
		}

		device, err := m.deviceList.SelectedDevice()
		if err != nil {
			ShowError(err, nil, m.parent)
			return
		}

		go func() {
			defer file.Close()
			InstallAAB(m.aabclient, device.Serial, file, m.getCustomKeystore(), m.parent)
		}()
	}, m.parent)

	fopenDialog.Resize(DialogSize(m.parent))
	fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".aab"}))
	fopenDialog.Show()
}

func (m *main) getCustomKeystore() *aabclient.KeystoreConfig {
	if !m.useCustomKeystoreCheck.Checked {
		return nil
	}

	keystoreChan := make(chan *aabclient.KeystoreConfig)
	go func() {
		defaultKeystore := aabclient.NewDefaultKeystoreConfig("./debug.keystore")

		keystorePathEntry := widget.NewEntry()
		keystorePathEntry.SetText(defaultKeystore.KeystorePath)

		keystorePathButton := widget.NewButtonWithIcon("Select", theme.FileIcon(), func() {
			fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
				if err != nil {
					ShowError(err, nil, m.parent)
					return
				}

				if file == nil {
					return
				}

				defer file.Close()
				keystorePathEntry.SetText(file.URI().Path())
			}, m.parent)

			fopenDialog.Resize(DialogSize(m.parent))
			fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".keystore"}))
			fopenDialog.Show()
		})

		keystorePassEntry := widget.NewPasswordEntry()
		keystorePassEntry.SetText(defaultKeystore.KeystorePass)

		keyAliasEntry := widget.NewEntry()
		keyAliasEntry.SetText(defaultKeystore.KeyAlias)

		keyPassEntry := widget.NewPasswordEntry()
		keyPassEntry.SetText(defaultKeystore.KeyPass)

		form := dialog.NewForm("Custom keystore config", "Confirm", "Dismiss", []*widget.FormItem{
			{Text: "Keystore path:", Widget: container.New(&alignToRightLayout{}, keystorePathEntry, keystorePathButton)},
			{Text: "Keystore pass:", Widget: keystorePassEntry},
			{Text: "Key alias:", Widget: keyAliasEntry},
			{Text: "Key pass:", Widget: keyPassEntry},
		}, func(submitted bool) {
			keystore := &aabclient.KeystoreConfig{
				KeystorePath: keystorePathEntry.Text,
				KeystorePass: keystorePassEntry.Text,
				KeyAlias:     keyAliasEntry.Text,
				KeyPass:      keyPassEntry.Text,
			}
			if !submitted || keystore.KeystorePath == keystorePathEntry.PlaceHolder {
				keystoreChan <- nil
			} else {
				keystoreChan <- keystore
			}
			close(keystoreChan)
		}, m.parent)

		form.Resize(fyne.Size{Width: m.parent.Canvas().Size().Width * 0.8, Height: 0})
		form.Show()
	}()

	return <-keystoreChan
}

func (m *main) tabItem() *container.TabItem {
	return &container.TabItem{Text: "Main", Icon: assets.MainTabIcon, Content: m.buildUI()}
}
