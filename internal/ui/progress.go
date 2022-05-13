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

func (p *ProgressBar) WithProgress() adbclient.UploadOption {
	once := sync.Once{}
	return adbclient.WithProgress(func(sentBytes int64, totalBytes int64) {
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

// Failed sets the text to indicate a failure.
func (p *ProgressBar) Install() {
	p.TextFormatter = func() string { return "Installing..." }
	p.Refresh()
}

// Failed sets the text to indicate a failure.
func (p *ProgressBar) Failed() {
	p.TextFormatter = func() string { return "Failed" }
	p.Refresh()
}

// NewProgressBar creates a new ProgressBar widget.
func NewProgressBar(parent fyne.Window) *ProgressBar {
	p := &ProgressBar{}
	p.ExtendBaseWidget(p)
	return p
}
