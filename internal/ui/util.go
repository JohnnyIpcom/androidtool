package ui

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// dialogSize returns a suitable dialog size.
func DialogSize(parent fyne.Window) fyne.Size {
	size := parent.Canvas().Size()
	return fyne.NewSize(size.Width*0.8, size.Height*0.8)
}

// NewBoldLabel creates a new label with bold text style.
func NewBoldLabel(text string) *widget.Label {
	return &widget.Label{Text: text, TextStyle: fyne.TextStyle{Bold: true}}
}

// IsFileExists returns true if the specified file exists.
func IsFileExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && !stat.IsDir()
}

// IsDirExists returns true if the specified directory exists.
func IsDirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}
