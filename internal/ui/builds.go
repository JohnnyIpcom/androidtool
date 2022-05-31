package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
)

type builds struct {
	app       fyne.App
	parent    fyne.Window
	aabClient *aabclient.Client
	aapt      *aapt.AAPT
	useCache  bool

	buildList *BuildList
}

func uiBuilds(app fyne.App, parent fyne.Window, aabClient *aabclient.Client, aapt *aapt.AAPT) *builds {
	return &builds{
		app:       app,
		parent:    parent,
		aabClient: aabClient,
		aapt:      aapt,
	}
}

func (b *builds) buildUI() *fyne.Container {
	b.buildList = NewBuildList(b.aabClient, b.aapt, b.parent)

	useCachedDataCheck := widget.NewCheck("Use cached data", func(checked bool) {
		b.useCache = checked
	})

	loadAPK := widget.NewButtonWithIcon("Load *.apk", assets.APKIcon, b.onLoadAPK)
	loadAAB := widget.NewButtonWithIcon("Load *.aab", assets.AABIcon, b.onLoadAAB)

	return container.NewBorder(
		nil,
		container.NewGridWithColumns(
			2,
			widget.NewCard(
				"",
				"",
				container.NewVBox(
					layout.NewSpacer(),
					loadAPK,
				),
			),
			widget.NewCard(
				"",
				"",
				container.NewVBox(
					useCachedDataCheck,
					loadAAB,
				),
			),
		),
		nil,
		nil,
		container.NewMax(
			widget.NewCard(
				"",
				"Builds:",
				b.buildList,
			),
		),
	)
}

func (b *builds) onLoadAPK() {
	fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			GetApp().ShowError(err, nil, b.parent)
			return
		}

		if file == nil {
			return
		}

		b.buildList.LoadAPK(file.URI().Path())
	}, b.parent)

	fopenDialog.Resize(DialogSize(b.parent))
	fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".apk"}))
	fopenDialog.Show()
}

func (b *builds) onLoadAAB() {
	fopenDialog := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			GetApp().ShowError(err, nil, b.parent)
			return
		}

		if file == nil {
			return
		}

		b.buildList.LoadAAB(file.URI().Path(), b.useCache)
	}, b.parent)

	fopenDialog.Resize(DialogSize(b.parent))
	fopenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".aab"}))
	fopenDialog.Show()
}

func (b *builds) tabItem() *container.TabItem {
	return &container.TabItem{Text: "Builds", Icon: assets.BuildsTabIcon, Content: b.buildUI()}
}
