package aabclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/johnnyipcom/androidtool/pkg/logger"
)

const (
	BundleToolDefaultVersion = "1.4.0"
	BundleToolReleasePath    = "https://github.com/google/bundletool/releases/download/"
	BundleToolReleaseFile    = "bundletool-all.jar"
	BundleToolJavaSettings   = "-Dfile.encoding=UTF-8"
)

// BundleTool is a wrapper around the bundletool jar.
type BundleTool struct {
	client       *http.Client
	version      string
	javaSettings string
	downloadURL  *url.URL
	downloadPath string
	log          logger.Logger
}

// NewBundleTool creates a new BundleTool instance.
func NewBundleTool(version string, log logger.Logger) (*BundleTool, error) {
	innerLog := log.WithField("module", "BundleTool")
	innerLog.Info("Creating AAB client...")

	downloadPath, downloadURL, err := getDownloadPathAndURLForVersion(version)
	if err != nil {
		return nil, err
	}

	return &BundleTool{
		client:       http.DefaultClient,
		version:      version,
		downloadURL:  downloadURL,
		downloadPath: downloadPath,
		log:          innerLog,
	}, nil
}

func getDownloadPathAndURLForVersion(version string) (string, *url.URL, error) {
	tempPath := filepath.Join(os.TempDir(), "bundletool", version)

	baseURL, err := url.Parse(BundleToolReleasePath)
	if err != nil {
		return "", nil, err
	}

	filename := fmt.Sprintf("bundletool-all-%s.jar", version)
	downloadURL, err := baseURL.Parse(path.Join(version, filename))
	if err != nil {
		return "", nil, err
	}

	return filepath.Join(tempPath, filename), downloadURL, nil
}

// SetVersion sets the version of bundletool to use. If the version is not installed, it will be downloaded.
func (b *BundleTool) SetVersion(version string) error {
	b.log.Infof("Setting version to %s...", version)

	downloadPath, downloadURL, err := getDownloadPathAndURLForVersion(version)
	if err != nil {
		return err
	}

	if !isInstalled(downloadPath) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := download(ctx, b.client, downloadURL, downloadPath)
		if err != nil {
			return err
		}
	}

	b.version = version
	b.downloadURL = downloadURL
	b.downloadPath = downloadPath
	b.log.Infof("Version set to %s.", version)
	return nil
}

// Version returns the version of bundletool that is being used.
func (b *BundleTool) Version() string {
	return b.version
}

func isInstalled(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsInstalled returns true if bundletool is installed.
func (b *BundleTool) IsInstalled() bool {
	b.log.Infof("Checking bundletool v%s on %s...", b.version, b.downloadPath)
	installed := isInstalled(b.downloadPath)
	if installed {
		b.log.Infof("Bundletool v%s is installed.", b.version)
	} else {
		b.log.Infof("Bundletool v%s is not installed.", b.version)
	}

	return installed
}

// SetJavaSettings sets the java settings to use when running bundletool.
func (b *BundleTool) SetJavaSettings(settings string) {
	b.javaSettings = settings
}

// JavaSettings returns the java settings to use when running bundletool.
func (b *BundleTool) JavaSettings() string {
	if b.javaSettings == "" {
		return BundleToolJavaSettings
	}

	return b.javaSettings
}

type progressFunc func(sentBytes int64)

type progressReader struct {
	reader       io.Reader
	progressFunc progressFunc
}

func (p *progressReader) Read(b []byte) (n int, err error) {
	n, err = p.reader.Read(b)
	if err != nil {
		return n, err
	}

	if p.progressFunc != nil {
		p.progressFunc(int64(n))
	}

	return n, err
}

type downloadOptions struct {
	progressFunc progressFunc
}

// DownloadOption is a functional option for downloading bundletool.
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

// WithProgress returns a DownloadOption that will report progress to the given progressFunc.
func WithProgress(f func(sentBytes int64)) DownloadOption {
	return progressDownloadOption{f}
}

func download(ctx context.Context, client *http.Client, url *url.URL, dst string, opts ...DownloadOption) error {
	var options downloadOptions
	for _, opt := range opts {
		err := opt.apply(&options)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}

	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	downloadDir := filepath.Dir(dst)
	if _, err := os.Stat(downloadDir); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(downloadDir, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	file, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer file.Close()

	var reader io.Reader = resp.Body
	if options.progressFunc != nil {
		reader = &progressReader{
			reader:       resp.Body,
			progressFunc: options.progressFunc,
		}
	}

	_, err = io.Copy(file, reader)
	return err
}

// Download downloads bundletool to the given path.
func (b *BundleTool) Download(ctx context.Context, opts ...DownloadOption) error {
	b.log.Infof("Downloading bundletool v%s from %s to %s...", b.version, b.downloadURL.String(), b.downloadPath)
	err := download(ctx, b.client, b.downloadURL, b.downloadPath, opts...)
	if err != nil {
		return err
	}

	b.log.Infof("Download complete.")
	return err
}

func (b *BundleTool) Run(ctx context.Context, args ...string) ([]byte, error) {
	b.log.Infof("Running bundletool v%s with args %v...", b.version, args)

	var command []string
	if b.javaSettings != "" {
		command = append(command, b.javaSettings)
	}
	command = append(command, "-jar", b.downloadPath)
	command = append(command, args...)

	cmd := exec.CommandContext(ctx, "java", command...)
	b.log.Debugf("Running command: %s", cmd.String())
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			scanner := bufio.NewScanner(bytes.NewReader(ee.Stderr))
			for scanner.Scan() {
				b.log.Error(scanner.Text())
			}

			return ee.Stderr, ee
		}

		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			b.log.Error(scanner.Text())
		}

		return out, err
	}

	b.log.Info("Bundletool run complete.")
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		b.log.Debug(scanner.Text())
	}

	return out, nil
}
