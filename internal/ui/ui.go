package ui

import (
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

func SetupUI(app fyne.App, parent fyne.Window) fyne.CanvasObject {
	prefs := app.Preferences()

	adbPort := prefs.IntWithFallback("adb_port", adbclient.DefaultPort)
	adbClient, err := adbclient.NewClient(adbPort, log.Default())
	if err != nil {
		log.Fatal(err)
	}

	bundletoolVersion := prefs.StringWithFallback("bundletool_version", aabclient.BundleToolDefaultVersion)
	aabClient, err := aabclient.NewClient(bundletoolVersion, log.Default())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := adbClient.Start(ctx); err != nil {
		log.Fatal(err)
	}

	if err := aabClient.Start(ctx); err != nil {
		log.Fatal(err)
	}

	parent.SetOnClosed(func() {
		cancel()
		adbClient.Stop()
		aabClient.Stop()
	})

	return &container.AppTabs{Items: []*container.TabItem{
		uiMain(app, parent, adbClient, aabClient).tabItem(),
		uiSettings(app, parent, adbClient, aabClient).tabItem(),
		uiAbout(adbClient).tabItem(),
	}}
}
