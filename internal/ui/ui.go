package ui

import (
	"context"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
)

func SetupUI(app fyne.App, parent fyne.Window) fyne.CanvasObject {
	client, err := adbclient.NewClient(log.Default())
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	if err := client.Start(ctx); err != nil {
		log.Fatal(err)
	}

	parent.SetOnClosed(func() {
		cancel()
		client.Stop()
	})

	return &container.AppTabs{Items: []*container.TabItem{
		uiMain(app, parent, client).tabItem(),
		uiAbout(client).tabItem(),
	}}
}
