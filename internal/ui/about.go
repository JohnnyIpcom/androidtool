package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

const version = "0.0.1"

type about struct {
	client       *adbclient.Client
	icon         *canvas.Image
	nameLabel    *widget.Label
	spacerLabel  *widget.Label
	versionLabel *widget.Label
}

func uiAbout(client *adbclient.Client) *about {
	return &about{
		client: client,
	}
}

func (a *about) buildUI() *fyne.Container {
	a.icon = canvas.NewImageFromResource(assets.AppIcon)
	a.icon.SetMinSize(fyne.NewSize(256, 256))

	a.nameLabel = NewBoldLabel("G5 Android Tool")
	a.spacerLabel = NewBoldLabel("-")
	a.versionLabel = NewBoldLabel("v" + version + " (client v" + a.client.Version() + ")")

	spacer := &layout.Spacer{}
	return container.NewVBox(
		spacer,
		container.NewHBox(spacer, a.icon, spacer),
		container.NewHBox(
			spacer,
			a.nameLabel,
			a.spacerLabel,
			a.versionLabel,
			spacer,
		),
		spacer,
	)
}

func (a *about) tabItem() *container.TabItem {
	return &container.TabItem{Text: "About", Icon: assets.AboutTabIcon, Content: a.buildUI()}
}
