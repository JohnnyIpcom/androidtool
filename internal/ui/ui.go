package ui

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/logger/logrus"
)

const (
	// DefaultLogPath is the default log file path.
	DefaultLogPath = "./androidtool.log"
)

func SetupUI(app fyne.App, parent fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	log := logrus.New(prefs.StringWithFallback("log_path", DefaultLogPath))

	adbPort := prefs.IntWithFallback("adb_port", adbclient.DefaultPort)
	adbClient, err := adbclient.NewClient(adbPort, log)
	if err != nil {
		log.Fatal(err)
	}

	bundletoolVersion := prefs.StringWithFallback("bundletool_version", aabclient.BundleToolDefaultVersion)
	aabClient, err := aabclient.NewClient(bundletoolVersion, log)
	if err != nil {
		log.Fatal(err)
	}

	javaSettings := prefs.StringWithFallback("bundletool_java_settings", aabclient.BundleToolJavaSettings)
	aabClient.SetBundleToolJavaSettings(javaSettings)

	ctx, cancel := context.WithCancel(context.Background())
	if err := adbClient.Start(ctx); err != nil {
		log.Fatal(err)
	}

	if err := aabClient.Start(ctx); err != nil {
		log.Fatal(err)
	}

	aapt, err := aapt.New(log)
	if err != nil {
		log.Fatal(err)
	}

	parent.SetOnClosed(func() {
		cancel()
		adbClient.Stop()
		aabClient.Stop()
	})

	return &container.AppTabs{Items: []*container.TabItem{
		uiMain(app, parent, adbClient, aabClient).tabItem(),
		uiBuilds(app, parent, aabClient, aapt).tabItem(),
		uiSettings(app, parent, adbClient, aabClient, log).tabItem(),
		uiAbout(adbClient).tabItem(),
	}}
}
