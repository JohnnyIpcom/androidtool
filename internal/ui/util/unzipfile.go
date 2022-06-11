package util

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileUnzipper struct {
	reader *zip.ReadCloser
	file   *zip.File
	size   uint64
	total  uint64
	dest   string
}

func NewFileUnzipper(zipPath string, filePath string, dest string) (*FileUnzipper, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}

	d, err := filepath.Abs(dest)
	if err != nil {
		return nil, err
	}

	var file *zip.File
	for _, f := range reader.File {
		if f.Name == filePath {
			file = f
			break
		}
	}

	if file == nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if file.Mode().IsDir() {
		return nil, fmt.Errorf("file is a directory: %s", filePath)
	}

	return &FileUnzipper{
		reader: reader,
		file:   file,
		size:   file.UncompressedSize64,
		dest:   d,
	}, nil
}

func (u *FileUnzipper) Close() error {
	return u.reader.Close()
}

func (u FileUnzipper) GetUncompressedSize() uint64 {
	return u.size
}

func (u *FileUnzipper) UnzipFile(ctx context.Context, progress progressFunc) error {
	progress(0, u.size)

	destFile, err := os.OpenFile(u.dest, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, u.file.Mode())
	if err != nil {
		return err
	}

	defer destFile.Close()

	zipFile, err := u.file.Open()
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

func UnzipFile(ctx context.Context, source string, file string, dest string, progressFunc progressFunc) error {
	// unzip apks to dir
	unzip, err := NewFileUnzipper(source, file, dest)
	if err != nil {
		return err
	}

	defer unzip.Close()

	if err := unzip.UnzipFile(ctx, progressFunc); err != nil {
		unzip.Close()
		return err
	}

	return nil
}
