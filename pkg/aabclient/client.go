package aabclient

import (
	"context"
	"fmt"
	"log"
)

// Client is a client for the Android Asset Packaging Tool.
type Client struct {
	tool *BundleTool
	log  *log.Logger
}

// NewClient creates a new AAB client.
func NewClient(version string, l *log.Logger) (*Client, error) {
	innerLog := log.New(l.Writer(), "[AABClient]  ", 0)
	innerLog.SetFlags(log.LstdFlags | log.Lshortfile)

	innerLog.Println("Creating AAB client...")
	tool, err := NewBundleTool(version, innerLog)
	if err != nil {
		return nil, err
	}

	return &Client{
		tool: tool,
		log:  innerLog,
	}, nil
}

// Start starts the AAB client.
func (c *Client) Start(ctx context.Context) error {
	c.log.Println("Starting AAB client...")

	// If the bundle tool is not installed, install it.
	if !c.tool.IsInstalled() {
		err := c.tool.Download(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stop stops the AAB client.
func (c *Client) Stop() {
	c.log.Println("Stopping AAB client...")
}

// SetBundleToolVersion sets the version of the bundle tool and downloads this version.
func (c *Client) SetBundleToolVersion(version string) error {
	return c.tool.SetVersion(version)
}

// BundleToolVersion returns the version of used bundle tool.
func (c *Client) BundleToolVersion() string {
	return c.tool.Version()
}

// KeystoreConfig returns the keystore configuration.
type KeystoreConfig struct {
	KeystorePath string // Path to the keystore.
	KeystorePass string // Password for the keystore.
	KeyAlias     string // Alias for the key.
	KeyPass      string // Password for the key.
}

// NewDefaultKeystoreConfig returns a new default keystore configuration.
func NewDefaultKeystoreConfig(path string) *KeystoreConfig {
	return &KeystoreConfig{
		KeystorePath: path,
		KeystorePass: "not specified",
		KeyAlias:     "android",
		KeyPass:      "not specified",
	}
}

// BuildAPKs unpacks the given AAB file and builds the universal APKs file bundle.
func (c *Client) BuildAPKs(ctx context.Context, aabPath string, apksPath string, serial string, keystore *KeystoreConfig) ([]byte, error) {
	c.log.Println("Building APKs...")
	var args []string = []string{
		"build-apks",
		"--overwrite",
		fmt.Sprintf("--bundle=%s", aabPath),
		fmt.Sprintf("--output=%s", apksPath),
	}

	if serial != "" {
		args = append(args,
			"--connected-device",
			fmt.Sprintf("--device-id=%s", serial),
		)
	}

	if keystore != nil {
		args = append(
			args,
			fmt.Sprintf("--ks=%s", keystore.KeystorePath),
			fmt.Sprintf("--ks-pass=pass:%s", keystore.KeystorePass),
			fmt.Sprintf("--ks-key-alias=%s", keystore.KeyAlias),
			fmt.Sprintf("--key-pass=pass:%s", keystore.KeyPass),
		)
	}

	return c.tool.Run(ctx, args...)
}

// InstallAPKs installs the given APKs file bundle on the device.
func (c *Client) InstallAPKs(ctx context.Context, apksPath string, serial string) ([]byte, error) {
	c.log.Println("Installing APKs...")
	args := []string{
		"install-apks",
		fmt.Sprintf("--apks=%s", apksPath),
	}

	if serial != "" {
		args = append(args,
			fmt.Sprintf("--device-id=%s", serial),
		)
	}

	return c.tool.Run(ctx, args...)
}
