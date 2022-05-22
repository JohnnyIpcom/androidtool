package ui

import (
	"context"
	"io"
	"os"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// FileSaver is a helper struct for saving files.
type FileSaver struct {
	mu     sync.RWMutex
	reader io.Reader
	cancel context.CancelFunc
	active bool
}

// NewFileSaver creates a new file saver.
func NewFileSaver(reader io.Reader) (*FileSaver, error) {
	return &FileSaver{
		reader: reader,
		active: false,
	}, nil
}

// Start start asynchronous saving of the file.
func (f *FileSaver) Start(path string) error {
	if f.isActive() {
		return nil
	}

	f.setActive(true)

	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	go func() {
		defer file.Close()

		buffer := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return

			default:
				n, err := f.reader.Read(buffer)
				if err != nil {
					if err == io.EOF {
						return
					}

					return
				}

				_, err = file.Write(buffer[:n])
				if err != nil {
					return
				}
			}
		}
	}()

	return nil
}

// Stop stops saving of the file.
func (f *FileSaver) Stop() {
	if f.cancel != nil {
		f.cancel()
	}

	f.setActive(false)
}

func (f *FileSaver) isActive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.active
}

func (f *FileSaver) setActive(active bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.active = active
}

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
