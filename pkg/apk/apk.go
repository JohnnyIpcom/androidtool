package apk

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"io"
	"io/fs"
	"os"

	"github.com/shogo82148/androidbinary"
	"github.com/shogo82148/androidbinary/apk"
)

type ScreenDPI uint8

const (
	ScreenLDPI ScreenDPI = iota
	ScreenMDPI
	ScreenHDPI
	ScreenXHDPI
	ScreenXXHDPI
	ScreenXXXHDPI
	ScreenNoDPI
	ScreenTVDPI
)

func (dpi ScreenDPI) String() string {
	switch dpi {
	case ScreenLDPI:
		return "ldpi"
	case ScreenMDPI:
		return "mdpi"
	case ScreenHDPI:
		return "hdpi"
	case ScreenXHDPI:
		return "xhdpi"
	case ScreenXXHDPI:
		return "xxhdpi"
	case ScreenXXXHDPI:
		return "xxxhdpi"
	case ScreenNoDPI:
		return "nodpi"
	case ScreenTVDPI:
		return "tvdpi"
	default:
		return "unknown"
	}
}

type APK struct {
	file     *os.File
	fs       fs.FS
	manifest apk.Manifest
	table    *androidbinary.TableFile
	size     int64
}

func NewAPK(path string) (*APK, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			file.Close()
		}
	}()

	fi, err := file.Stat()
	if err != nil {
		return nil, err
	}

	reader, err := zip.NewReader(file, fi.Size())
	if err != nil {
		return nil, err
	}

	apk := &APK{
		file: file,
		fs:   reader,
		size: fi.Size(),
	}

	err = apk.parseResources()
	if err != nil {
		return nil, err
	}

	err = apk.parseManifest()
	if err != nil {
		return nil, err
	}

	return apk, nil
}

func (a *APK) Close() error {
	if a.file != nil {
		return a.file.Close()
	}

	return nil
}

func (a *APK) File(path string) (data []byte, err error) {
	file, err := a.fs.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return io.ReadAll(file)
}

func (a *APK) Image(path string) (image.Image, error) {
	imgData, err := a.File(path)
	if err != nil {
		return nil, err
	}

	m, _, err := image.Decode(bytes.NewReader(imgData))
	return m, err
}

func (a *APK) parseManifest() error {
	xmlData, err := a.File("AndroidManifest.xml")
	if err != nil {
		return err
	}

	xmlFile, err := androidbinary.NewXMLFile(bytes.NewReader(xmlData))
	if err != nil {
		return err
	}

	return xmlFile.Decode(&a.manifest, a.table, nil)
}

func (a *APK) parseResources() (err error) {
	resData, err := a.File("resources.arsc")
	if err != nil {
		return
	}

	a.table, err = androidbinary.NewTableFile(bytes.NewReader(resData))
	return
}

func (a *APK) Label() string {
	return a.manifest.App.Label.MustString()
}

func (a *APK) VersionCode() int32 {
	return a.manifest.VersionCode.MustInt32()
}

func (a *APK) VersionName() string {
	return a.manifest.VersionName.MustString()
}

func (a *APK) Identifier() string {
	return a.manifest.Package.MustString()
}

func (a *APK) Icon(dpi ScreenDPI) (image.Image, error) {
	var dpiMap map[ScreenDPI]uint16 = map[ScreenDPI]uint16{
		ScreenLDPI:    120,
		ScreenMDPI:    160,
		ScreenHDPI:    240,
		ScreenXHDPI:   320,
		ScreenXXHDPI:  480,
		ScreenXXXHDPI: 640,
		ScreenNoDPI:   0,
		ScreenTVDPI:   213,
	}

	density, ok := dpiMap[dpi]
	if !ok {
		return nil, fmt.Errorf("invalid dpi")
	}

	iconPath, err := a.manifest.App.Icon.WithResTableConfig(&androidbinary.ResTableConfig{
		Density: density,
	}).String()
	if err != nil {
		return nil, err
	}

	if androidbinary.IsResID(iconPath) {
		return nil, fmt.Errorf("unable to convert icon-id to icon path")
	}

	return a.Image(iconPath)
}

func (a *APK) Size() int64 {
	return a.size
}
