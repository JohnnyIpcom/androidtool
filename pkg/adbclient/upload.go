package adbclient

import (
	"context"
	"io"
	"os"
	"time"

	adb "github.com/zach-klippenstein/goadb"
)

type progressFunc func(sentBytes int64, totalBytes int64)

type uploadOptions struct {
	progressFunc progressFunc
}

// UploadOption is an option for uploading file.
type UploadOption interface {
	apply(*uploadOptions) error
}

type progressUploadOption struct {
	progressFunc progressFunc
}

func (o progressUploadOption) apply(opts *uploadOptions) error {
	opts.progressFunc = o.progressFunc
	return nil
}

func WithUploadProgress(f func(sentBytes int64, totalBytes int64)) UploadOption {
	return progressUploadOption{f}
}

type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) {
	return rf(p)
}

// Upload uploads a file to the device.
func (c *Client) Upload(ctx context.Context, device *Device, src, dst string, opts ...UploadOption) error {
	c.log.Infof("Uploading to %s...", dst)

	var options uploadOptions
	for _, opt := range opts {
		err := opt.apply(&options)
		if err != nil {
			return nil
		}
	}

	file, err := os.Open(src)
	if err != nil {
		return err
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}

	c.log.Debugf("Uploading %d bytes...", fi.Size())

	w, err := c.adb.Device(adb.DeviceWithSerial(device.Serial)).OpenWrite(dst, os.FileMode(0664), time.Now())
	if err != nil {
		return err
	}

	defer w.Close()

	total := 0
	_, err = io.Copy(w, readerFunc(func(b []byte) (int, error) {
		select {
		case <-ctx.Done():
			c.log.Debug("Upload canceled")
			return 0, ctx.Err()

		default:
			n, err := file.Read(b)
			if err != nil {
				return 0, err
			}

			total += n
			if options.progressFunc != nil {
				options.progressFunc(int64(total), fi.Size())
			}

			return n, err
		}
	}))

	return err
}
