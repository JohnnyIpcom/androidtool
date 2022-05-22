package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/internal/ui"
)

func main() {
	a := app.NewWithID("com.johnnyipcom.androidtool")
	a.SetIcon(assets.AppIcon)

	w := a.NewWindow("G5 Android tool")

	icon := canvas.NewImageFromResource(assets.AppIcon)
	icon.SetMinSize(fyne.NewSize(128, 128))

	driver := a.Driver().(desktop.Driver)
	splash := driver.CreateSplashWindow()
	splash.SetContent(
		container.NewVBox(
			layout.NewSpacer(),
			widget.NewLabelWithStyle("Loading...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			layout.NewSpacer(),
			icon,
			layout.NewSpacer(),
			widget.NewProgressBarInfinite(),
			layout.NewSpacer(),
		),
	)

	uiChan := make(chan fyne.CanvasObject)
	go func() {
		uiChan <- ui.SetupUI(a, w)
		close(uiChan)
	}()

	go func() {
		w.SetContent(<-uiChan)
		w.CenterOnScreen()
		w.Resize(fyne.NewSize(800, 500))
		w.Show()

		splash.Close()
	}()

	splash.ShowAndRun()
}
