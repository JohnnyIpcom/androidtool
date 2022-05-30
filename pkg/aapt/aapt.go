package aapt

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/johnnyipcom/androidtool/pkg/logger"
)

// AAPTConfig is the configuration for the AAPT client.
type Config struct {
	// Path to the aapt executable. If empty, the ANDROID_HOME environment variable will be searched.
	PathToAAPT string
}

// AAPT is the client for the Android Asset Packaging Tool.
type AAPT struct {
	path string
	log  logger.Logger
}

func New(log logger.Logger) (*AAPT, error) {
	return NewWithConfig(Config{}, log)
}

func NewWithConfig(config Config, log logger.Logger) (*AAPT, error) {
	innerLog := log.WithField("component", "aapt")
	innerLog.Info("Creating aapt client")

	pathToAAPT := config.PathToAAPT
	if config.PathToAAPT == "" {
		home := os.Getenv("ANDROID_HOME")
		if home != "" {
			toolsDir := filepath.Join(home, "build-tools")
			files, err := ioutil.ReadDir(toolsDir)
			if err != nil {
				return nil, err
			}

			// Find the newest build-tools directory.
			newest := files[0]
			for _, f := range files {
				if f.Name() > newest.Name() {
					newest = f
				}
			}

			pathToAAPT = filepath.Join(toolsDir, newest.Name(), "aapt")
		} else {
			return nil, fmt.Errorf("path to aapt not specified")
		}
	}

	innerLog.Debug("Using aapt at: ", pathToAAPT)
	return &AAPT{
		path: pathToAAPT,
		log:  innerLog,
	}, nil
}

type Filter func(line string) bool

func runCommand(cmd *exec.Cmd, filter Filter) ([]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, out)
	}

	if filter != nil {
		lines := strings.Split(string(out), "\n")
		out = []byte{}
		for _, line := range lines {
			if filter(line) {
				out = append(out, line...)
				out = append(out, '\n')
			}
		}
	}

	return out, nil
}

// DumpBadging dumps the badging information for the given APK.
func (a *AAPT) DumpBadging(apkPath string, filter Filter) ([]byte, error) {
	a.log.Infof("Dumping badging for %s", apkPath)
	cmd := exec.Command(a.path, "dump", "badging", apkPath)
	return runCommand(cmd, filter)
}

// GetNativeCodeABIs returns the native code ABIs for the given APK.
func (a *AAPT) GetNativeCodeABIs(apkPath string) ([]string, error) {
	a.log.Infof("Getting native code ABIs for %s", apkPath)
	out, err := a.DumpBadging(apkPath, func(line string) bool {
		return strings.HasPrefix(line, "native-code: ")
	})

	if err != nil {
		return nil, err
	}

	var line string
	line = strings.TrimPrefix(string(out), "native-code: ")
	line = strings.TrimSuffix(line, "\r\n")

	var archs []string
	for _, line := range strings.Split(line, " ") {
		archs = append(archs, strings.Trim(line, "'"))
	}

	a.log.Debugf("Native code ABIs: %v", archs)
	return archs, nil
}

// GetManifest returns the manifest for the given APK.
func (a *AAPT) GetManifest(apkPath string) ([]byte, error) {
	a.log.Infof("Getting manifest for %s", apkPath)
	cmd := exec.Command(a.path, "dump", "xmltree", apkPath, "AndroidManifest.xml")
	return runCommand(cmd, nil)
}
