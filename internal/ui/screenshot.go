package ui

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

const (
	// DefaultScreenshotPath is the default screenshot file path.
	DefaultScreenshotPath = "./screenshot.png"
)

func previewImage(size fyne.Size, color color.Color, textColor color.Color, text string) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, int(size.Width), int(size.Height)))
	for y := 0; y < int(size.Height); y++ {
		for x := 0; x < int(size.Width); x++ {
			img.Set(x, y, color)
		}
	}

	face := basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(textColor),
		Face: face,
	}

	textSize := drawer.MeasureString(text)
	drawer.Dot = fixed.Point26_6{
		X: fixed.I(int(size.Width/2)) - textSize/2,
		Y: fixed.I(int(size.Height/2) - face.Height/2),
	}

	drawer.DrawString(text)
	return img
}

type ScreenshotImage struct {
	widget.BaseWidget

	min fyne.Size
	src image.Image
	dst *canvas.Image
}

func NewScreenshotImageFromReader(r io.Reader) (*ScreenshotImage, error) {
	img := &ScreenshotImage{}
	img.ExtendBaseWidget(img)

	img.dst = &canvas.Image{}
	img.dst.FillMode = canvas.ImageFillStretch

	return img, img.LoadFromReader(r)
}

func NewScreenshotImageFromImage(src image.Image) (*ScreenshotImage, error) {
	img := &ScreenshotImage{}
	img.ExtendBaseWidget(img)

	img.dst = &canvas.Image{}
	img.dst.FillMode = canvas.ImageFillContain

	return img, img.LoadFromImage(src)
}

func (img *ScreenshotImage) CreateRenderer() fyne.WidgetRenderer {
	return &screenshotImageRenderer{
		screenshot: img,
	}
}

func (img *ScreenshotImage) LoadFromReader(r io.Reader) error {
	img.dst.Image = nil
	img.dst.Refresh()

	src, _, err := image.Decode(r)
	if err != nil {
		return err
	}

	return img.LoadFromImage(src)
}

func (img *ScreenshotImage) LoadFromImage(src image.Image) error {
	img.src = src

	img.dst.Image = src
	img.dst.FillMode = canvas.ImageFillContain
	img.dst.Refresh()
	return nil
}

func (img *ScreenshotImage) MinSize() fyne.Size {
	return img.min
}

func (img *ScreenshotImage) SetMinSize(min fyne.Size) {
	img.min = min
}

type screenshotImageRenderer struct {
	screenshot *ScreenshotImage
}

var _ fyne.WidgetRenderer = &screenshotImageRenderer{}

func (r *screenshotImageRenderer) Destroy() {
}

func (r *screenshotImageRenderer) Layout(size fyne.Size) {
	r.screenshot.dst.Resize(size)
}

func (r *screenshotImageRenderer) MinSize() fyne.Size {
	return r.screenshot.min
}

func (r *screenshotImageRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.screenshot.dst}
}

func (r *screenshotImageRenderer) Refresh() {
	r.screenshot.dst.Refresh()
}

func Screenshot(client *adbclient.Client, device *adbclient.Device, parent fyne.Window) {
	width, height := float32(device.Display.Width), float32(device.Display.Height)
	if width < height {
		width, height = height, width
	}

	img := previewImage(fyne.NewSize(width, height), color.Black, color.White, fmt.Sprintf("%dx%d", device.Display.Width, device.Display.Height))
	screenshotImage, _ := NewScreenshotImageFromImage(img)
	screenshotImage.SetMinSize(fyne.NewSize(256*width/height, 256))

	screenshotPathEntry := widget.NewEntry()
	screenshotPathEntry.SetText(DefaultScreenshotPath)

	screenshotPathButton := widget.NewButtonWithIcon("Select", theme.DocumentSaveIcon(), func() {
		fsaveDialog := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				return
			}

			if file == nil {
				return
			}

			defer file.Close()
			screenshotPathEntry.SetText(file.URI().Path())
		}, parent)

		fsaveDialog.SetFileName("screenshot.png")
		fsaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png"}))
		fsaveDialog.Resize(DialogSize(parent))
		fsaveDialog.Show()
	})

	makeScreenshotButton := widget.NewButtonWithIcon("Screenshot", assets.ScreenshotIcon, nil)

	d := dialog.NewCustom(
		"Screenshot",
		"Close",
		container.NewBorder(
			container.New(&alignToRightLayout{}, screenshotPathEntry, screenshotPathButton),
			container.NewCenter(makeScreenshotButton),
			nil,
			nil,
			container.NewMax(screenshotImage),
		),
		parent,
	)

	makeScreenshotButton.OnTapped = func() {
		screenshotPath := client.GetScreenshotPath()
		if err := client.Screenshot(device, screenshotPath, adbclient.WithScreenshotAsPng()); err != nil {
			GetApp().ShowError(err, d.Hide, parent)
			return
		}

		defer func() {
			if err := client.RemoveFile(device, screenshotPath); err != nil {
				GetApp().ShowError(err, d.Hide, parent)
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := client.Download(ctx, device, screenshotPath, screenshotPathEntry.Text); err != nil {
			GetApp().ShowError(err, d.Hide, parent)
			return
		}

		f, err := os.Open(screenshotPathEntry.Text)
		if err != nil {
			GetApp().ShowError(err, d.Hide, parent)
			return
		}

		defer f.Close()

		image, _, err := image.Decode(f)
		if err != nil {
			GetApp().ShowError(err, d.Hide, parent)
			return
		}

		screenshotImage.LoadFromImage(image)
	}

	d.Resize(DialogSize(parent))
	d.Show()
}
