package adbclient

import (
	"fmt"
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
func (c *Client) Screenshot(device *Device, path string, opts ...ScreenshotOption) error {
	c.log.Info("Taking screenshot...")

	options := screenshotOptions{}
	for _, o := range opts {
		if err := o.apply(&options); err != nil {
			return err
		}
	}

	resp, err := c.runCommand(device, fmt.Sprintf("screencap %s %s", strings.Join(options.Options(), " "), path))
	if err != nil {
		return err
	}

	c.log.Debugf("Got response: %s", resp)
	return nil
}
