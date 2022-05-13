package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// dialogSize returns a suitable dialog size.
func DialogSize(parent fyne.Window) fyne.Size {
	size := parent.Canvas().Size()
	return fyne.NewSize(size.Width*0.8, size.Height*0.8)
}

// ShowConfirmation shows a confirmation dialog with appropriate size
func ShowInformation(title, message string, parent fyne.Window) {
	dc := dialog.NewInformation(title, message, parent)
	dc.Resize(fyne.NewSize(500, 200))
	dc.Show()
}

// ShowError shows a dialog over the specified window for an application error.
func ShowError(err error, closed func(), parent fyne.Window) {
	label := widget.NewLabel(err.Error())
	label.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(label)

	de := dialog.NewCustom("Error", "OK", container.NewMax(scroll), parent)
	de.Resize(fyne.NewSize(500, 200))
	if closed != nil {
		de.SetOnClosed(closed)
	}
	de.Show()
}

// NewBoldLabel creates a new label with bold text style.
func NewBoldLabel(text string) *widget.Label {
	return &widget.Label{Text: text, TextStyle: fyne.TextStyle{Bold: true}}
}
