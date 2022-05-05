package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

// DialogSize returns a suitable dialog size.
func DialogSize(parent fyne.Window) fyne.Size {
	size := parent.Canvas().Size()
	return fyne.NewSize(size.Width*0.8, size.Height*0.8)
}

// ShowConfirmation shows a confirmation dialog with appropriate size
func ShowInformation(title, message string, parent fyne.Window) {
	confirm := dialog.NewInformation(title, message, parent)
	confirm.Resize(DialogSize(parent))
	confirm.Show()
}
