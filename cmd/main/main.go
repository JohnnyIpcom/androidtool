package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/internal/ui"
)

func main() {
	a := app.NewWithID("com.johnnyipcom.androidtool")
	a.SetIcon(assets.AppIcon)

	w := a.NewWindow("G5 Android tool")
	w.SetContent(ui.SetupUI(a, w))
	w.Resize(fyne.NewSize(700, 400))
	w.ShowAndRun()
}
