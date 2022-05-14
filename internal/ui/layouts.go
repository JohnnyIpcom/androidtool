package ui

import "fyne.io/fyne/v2"

type alignToRightLayout struct{}

var _ fyne.Layout = &alignToRightLayout{}

func (l *alignToRightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	topRight := fyne.NewPos(size.Width, 0)

	var width float32
	for i := len(objects) - 1; i >= 1; i-- {
		width += objects[i].MinSize().Width

		topRight = topRight.Subtract(fyne.NewSize(objects[i].MinSize().Width, 0))
		objects[i].Move(topRight)
		objects[i].Resize(fyne.Size{Width: objects[i].MinSize().Width, Height: size.Height})
	}

	if len(objects) > 0 {
		objects[0].Move(fyne.NewPos(0, 0))
		objects[0].Resize(fyne.Size{Width: size.Width - width, Height: size.Height})
	}
}

func (l *alignToRightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	var minHeight, totalWidth float32
	for _, object := range objects {
		if object.MinSize().Height > minHeight {
			minHeight = object.MinSize().Height
		}

		totalWidth += object.MinSize().Width
	}

	return fyne.NewSize(totalWidth, minHeight)
}
