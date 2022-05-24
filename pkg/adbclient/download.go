package adbclient

import (
	"context"
	"io"
	"os"

	adb "github.com/zach-klippenstein/goadb"
)

type downloadOptions struct {
	progressFunc progressFunc
}

// DownloadOption is an option for downloading file.
type DownloadOption interface {
	apply(*downloadOptions) error
}

type progressDownloadOption struct {
	progressFunc progressFunc
}

func (o progressDownloadOption) apply(opts *downloadOptions) error {
	opts.progressFunc = o.progressFunc
	return nil
}

// WithDownloadProgress sets a progress function for downloading file.
func WithDownloadProgress(f func(sentBytes int64, totalBytes int64)) DownloadOption {
	return progressDownloadOption{f}
}

// Download downloads a file from the device.
func (c *Client) Download(ctx context.Context, device *Device, src, dst string, opts ...DownloadOption) error {
	c.log.Infof("Downloading %s to %s...", src, dst)

	options := downloadOptions{}
	for _, o := range opts {
		if err := o.apply(&options); err != nil {
			return err
		}
	}

	d := c.adb.Device(adb.DeviceWithSerial(device.Serial))

	size, err := diskUsage(d, src)
	if err != nil {
		return err
	}

	c.log.Debugf("Downloading %d bytes", size)

	r, err := d.OpenRead(src)
	if err != nil {
		return err
	}

	defer r.Close()

	file, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer file.Close()

	total := 0
	_, err = io.Copy(file, readerFunc(func(b []byte) (int, error) {
		select {
		case <-ctx.Done():
			c.log.Debug("Download canceled")
			return 0, ctx.Err()

		default:
			n, err := r.Read(b)
			if err != nil {
				return 0, err
			}

			total += n
			if options.progressFunc != nil {
				options.progressFunc(int64(total), int64(size))
			}

			return n, err
		}
	}))

	return err
}
