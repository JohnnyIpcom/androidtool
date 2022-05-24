package ui

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

// ProgressBar is a widget that shows the progress of an install.
type ProgressBar struct {
	widget.ProgressBar
}

func (p *ProgressBar) WithUploadProgress() adbclient.UploadOption {
	once := sync.Once{}
	return adbclient.WithUploadProgress(func(sentBytes int64, totalBytes int64) {
		once.Do(func() { p.Max = float64(totalBytes) })
		p.SetValue(float64(sentBytes))
	})
}

func (p *ProgressBar) WithDownloadProgress() adbclient.DownloadOption {
	once := sync.Once{}
	return adbclient.WithDownloadProgress(func(sentBytes int64, totalBytes int64) {
		once.Do(func() { p.Max = float64(totalBytes) })
		p.SetValue(float64(sentBytes))
	})
}

func (p *ProgressBar) WithMax(max float64) aabclient.DownloadOption {
	return aabclient.WithProgress(func(sentBytes int64) {
		p.SetValue(p.Value + float64(sentBytes))
	})
}

// Done sets the value to max to indicate that it is finished.
func (p *ProgressBar) Done() {
	p.SetValue(p.Max)
}

// SetText sets the text of the ProgressBar. Send an empty string to hide the text and return to the default.
func (p *ProgressBar) SetText(text string) {
	if len(text) > 0 {
		p.TextFormatter = func() string { return text }
	} else {
		p.TextFormatter = nil
	}

	p.Refresh()
}

// NewProgressBar creates a new ProgressBar widget.
func NewProgressBar(parent fyne.Window) *ProgressBar {
	p := &ProgressBar{}
	p.ExtendBaseWidget(p)
	return p
}
