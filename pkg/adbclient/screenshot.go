package adbclient

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"
)

type screenshotOptions struct {
	asPng bool
}

func (o screenshotOptions) String() string {
	return fmt.Sprintf("asPng:%t", o.asPng)
}

func (o screenshotOptions) Options() []string {
	var options []string
	if o.asPng {
		options = append(options, "-p")
	}

	return options
}

type ScreenshotOption interface {
	apply(*screenshotOptions) error
}

type screenshotAsPngOption struct{}

func (o screenshotAsPngOption) apply(opts *screenshotOptions) error {
	opts.asPng = true
	return nil
}

func WithScreenshotAsPng() ScreenshotOption {
	return screenshotAsPngOption{}
}

// Screenshot takes a screenshot of the device.
func (c *Client) Screenshot(device *Device, opts ...ScreenshotOption) (image.Image, error) {
	c.log.Info("Taking screenshot...")

	options := screenshotOptions{}
	for _, o := range opts {
		if err := o.apply(&options); err != nil {
			return nil, err
		}
	}

	resp, err := c.runCommand(device, fmt.Sprintf("screencap %s", strings.Join(options.Options(), " ")))
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(bytes.NewReader(resp))
	if err != nil {
		return nil, err
	}

	return img, nil
}
