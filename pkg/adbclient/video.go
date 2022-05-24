package adbclient

import (
	"fmt"
	"strings"
	"time"
)

type videoOptions struct {
	duration time.Duration
	bitrate  int
}

func (o videoOptions) String() string {
	return fmt.Sprintf("duration:%s bitrate:%d", o.duration, o.bitrate)
}

func (o videoOptions) Options() []string {
	var options []string
	if o.duration != 0 {
		options = append(options, fmt.Sprintf("--time-limit %d", int(o.duration.Seconds())))
	}

	if o.bitrate != 0 {
		options = append(options, fmt.Sprintf("--bit-rate %d", o.bitrate))
	}

	return options
}

// VideoOption is an option for video recording.
type VideoOption interface {
	apply(*videoOptions) error
}

type videoDurationOption struct {
	duration time.Duration
}

func (o videoDurationOption) apply(opts *videoOptions) error {
	opts.duration = o.duration
	return nil
}

type videoBitrateOption struct {
	bitrate int
}

func (o videoBitrateOption) apply(opts *videoOptions) error {
	opts.bitrate = o.bitrate
	return nil
}

// WithVideoDuration sets the duration for video recording.
func WithVideoDuration(duration time.Duration) VideoOption {
	return videoDurationOption{
		duration: duration,
	}
}

// WithVideoBitrate sets the bitrate for video recording.
func WithVideoBitrate(bitrate int) VideoOption {
	return videoBitrateOption{
		bitrate: bitrate,
	}
}

// Video takes a video from the device.
func (c *Client) Video(device *Device, width int, height int, path string, opts ...VideoOption) error {
	c.log.Info("Recording video...")

	options := videoOptions{
		duration: 180 * time.Second, // 3 minutes
		bitrate:  20000000,          // 20Mbps
	}

	for _, opt := range opts {
		err := opt.apply(&options)
		if err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("screenrecord --verbose --size %dx%d %s %s", width, height, strings.Join(options.Options(), " "), path)
	resp, err := c.runCommand(device, cmd)
	if err != nil {
		return err
	}

	c.log.Debugf("Got response: %s", resp)
	return nil
}
