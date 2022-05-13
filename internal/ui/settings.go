package ui

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	appearance "fyne.io/fyne/v2/cmd/fyne_settings/settings"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

type settings struct {
	app       fyne.App
	parent    fyne.Window
	adbClient *adbclient.Client
	aabClient *aabclient.Client
	prefs     fyne.Preferences

	installPathEntry       *widget.Entry
	adbPortEntry           *widget.Entry
	bundletoolVersionEntry *widget.Entry
}

func uiSettings(app fyne.App, parent fyne.Window, adbClient *adbclient.Client, aabClient *aabclient.Client) *settings {
	return &settings{
		app:       app,
		parent:    parent,
		adbClient: adbClient,
		aabClient: aabClient,
		prefs:     app.Preferences(),
	}
}

func (s *settings) onInstallPathSubmitted(path string) {
	s.prefs.SetString("install_path", path)
	s.adbClient.SetInstallPath(path)
}

func (s *settings) onADBPortSubmitted(port string) {
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return
	}

	if s.adbClient.Port() == portInt {
		return
	}

	s.prefs.SetInt("adb_port", portInt)
	ShowInformation("ADB port changed", "You must restart the application for the new port to take effect.", s.parent)
}

func (s *settings) onBundleToolVersionSubmitted(version string) {
	if s.aabClient.BundleToolVersion() == version {
		return
	}

	pbar := widget.NewProgressBarInfinite()
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(fyne.NewSize(200, 0))

	d := dialog.NewCustom("Setting bundletool version", "Cancel", container.NewMax(rect, pbar), s.parent)
	d.Show()

	go func() {
		if err := s.aabClient.SetBundleToolVersion(version); err != nil {
			d.Hide()
			s.prefs.SetString("bundletool_version", s.aabClient.BundleToolVersion())
			s.bundletoolVersionEntry.SetText(s.aabClient.BundleToolVersion())
			dialog.ShowError(err, s.parent)
			return
		}

		s.prefs.SetString("bundletool_version", s.aabClient.BundleToolVersion())
		s.bundletoolVersionEntry.SetText(s.aabClient.BundleToolVersion())
		d.Hide()
	}()
}

func (s *settings) applyPreferences() {
	path := s.prefs.StringWithFallback("install_path", adbclient.DefaultInstallPath)
	s.installPathEntry.SetText(path)

	port := s.prefs.IntWithFallback("adb_port", adbclient.DefaultPort)
	s.adbPortEntry.SetText(strconv.FormatInt(int64(port), 10))

	version := s.prefs.StringWithFallback("bundletool_version", aabclient.BundleToolDefaultVersion)
	s.bundletoolVersionEntry.SetText(version)
}

func (s *settings) buildADBUI() fyne.CanvasObject {
	s.installPathEntry = &widget.Entry{
		PlaceHolder: adbclient.DefaultInstallPath,
		OnSubmitted: s.onInstallPathSubmitted,
	}

	s.adbPortEntry = &widget.Entry{
		PlaceHolder: strconv.FormatInt(adbclient.DefaultPort, 10),
		OnSubmitted: s.onADBPortSubmitted,

		Validator: func(s string) error {
			port, err := strconv.Atoi(s)
			if err != nil {
				return err
			}

			if port < 1 || port > 65535 {
				return fmt.Errorf("port must be between 1 and 65535")
			}

			return nil
		},
	}

	return container.NewVBox(
		container.NewGridWithColumns(2,
			NewBoldLabel("Install path:"),
			s.installPathEntry,
		),
		widget.NewAccordion(
			widget.NewAccordionItem(
				"Advanced",
				container.NewGridWithColumns(
					2,
					NewBoldLabel("ADB port:"),
					s.adbPortEntry,
				),
			),
		),
	)
}

func (s *settings) buildBundleToolUI() fyne.CanvasObject {
	s.bundletoolVersionEntry = &widget.Entry{
		PlaceHolder: aabclient.BundleToolDefaultVersion,
		OnSubmitted: s.onBundleToolVersionSubmitted,

		Validator: func(s string) error {
			if s == "" {
				return fmt.Errorf("version cannot be empty")
			}

			regexp := regexp.MustCompile(`^[0-9]+(\.[0-9]+)+(\.[0-9]+)?$`)
			if !regexp.MatchString(s) {
				return fmt.Errorf("version must be a valid version number")
			}

			return nil
		},
	}

	return container.NewVBox(
		container.NewGridWithColumns(2,
			NewBoldLabel("Bundletool version:"),
			s.bundletoolVersionEntry,
		),
	)
}

func (s *settings) buildUI() *fyne.Container {
	adbUI := s.buildADBUI()
	bundletoolUI := s.buildBundleToolUI()
	interfaceContainer := appearance.NewSettings().LoadAppearanceScreen(s.parent)

	s.applyPreferences()

	return container.NewVBox(
		widget.NewCard(
			"User Interface",
			"",
			interfaceContainer,
		),
		widget.NewCard(
			"Android Debug Bridge",
			"",
			adbUI,
		),
		widget.NewCard(
			"Bundletool",
			"",
			bundletoolUI,
		),
	)
}

func (s *settings) tabItem() *container.TabItem {
	return &container.TabItem{Text: "Settings", Icon: assets.SettingsTabIcon, Content: s.buildUI()}
}
