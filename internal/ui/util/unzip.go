package util

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Unzipper struct {
	reader *zip.ReadCloser
	size   uint64
	total  uint64
	dest   string
}

func NewUnzipper(zipPath string, dest string) (*Unzipper, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}

	d, err := filepath.Abs(dest)
	if err != nil {
		return nil, err
	}

	var size uint64
	for _, f := range reader.File {
		size += f.UncompressedSize64
	}

	return &Unzipper{
		size:   size,
		reader: reader,
		dest:   d,
	}, nil
}

func (u *Unzipper) Close() error {
	return u.reader.Close()
}

type progressFunc func(current, total uint64)

func (u Unzipper) GetUncompressedSize() uint64 {
	return u.size
}

func (u *Unzipper) Unzip(ctx context.Context, progress progressFunc) error {
	for _, file := range u.reader.File {
		progress(0, u.size)
		if err := u.unzipFile(ctx, file, u.dest, progress); err != nil {
			return err
		}
	}

	return nil
}

type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) {
	return rf(p)
}

func (u *Unzipper) unzipFile(ctx context.Context, f *zip.File, dest string, progress progressFunc) error {
	// Check if file paths are not vulnerable to Zip Slip
	filePath := filepath.Join(dest, f.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	if f.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
			return err
		}

		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
		return err
	}

	destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}

	defer destFile.Close()

	zipFile, err := f.Open()
	if err != nil {
		return err
	}

	defer zipFile.Close()

	_, err = io.Copy(destFile, readerFunc(func(b []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()

		default:
			n, err := zipFile.Read(b)
			if err != nil {
				return n, err
			}

			u.total += uint64(n)
			progress(u.total, u.size)

			return n, nil
		}
	}))

	return err
}

func Unzip(ctx context.Context, source string, dest string, progressFunc progressFunc) error {
	// unzip apks to dir
	unzip, err := NewUnzipper(source, dest)
	if err != nil {
		return err
	}

	defer unzip.Close()

	if err := unzip.Unzip(ctx, progressFunc); err != nil {
		unzip.Close()
		return err
	}

	return nil
}
