package ui

import (
	"context"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/johnnyipcom/androidtool/internal/assets"
	"github.com/johnnyipcom/androidtool/internal/storage"
	"github.com/johnnyipcom/androidtool/pkg/aabclient"
	"github.com/johnnyipcom/androidtool/pkg/aapt"
	"github.com/johnnyipcom/androidtool/pkg/adbclient"
	"github.com/johnnyipcom/androidtool/pkg/logger"
	"github.com/johnnyipcom/androidtool/pkg/logger/logrus"
)

const (
	// DefaultLogPath is the default log file path.
	DefaultLogPath = "./androidtool.log"
)

var (
	once     sync.Once
	instance *App
)

// Application is the application.
type App struct {
	app       fyne.App
	window    fyne.Window
	storage   *storage.Storage
	adbClient *adbclient.Client
	aabClient *aabclient.Client
	aapt      *aapt.AAPT
	log       logger.Logger
}

func GetApp() *App {
	once.Do(func() {
		a := app.NewWithID("com.johnnyipcom.androidtool")
		a.SetIcon(assets.AppIcon)

		w := a.NewWindow("Android tool")
		instance = &App{
			app:    a,
			window: w,
		}
	})

	return instance
}

// Run runs the application.
func (a *App) Run() {
	icon := canvas.NewImageFromResource(assets.AppIcon)
	icon.SetMinSize(fyne.NewSize(128, 128))

	driver := a.app.Driver().(desktop.Driver)
	splash := driver.CreateSplashWindow()
	splash.SetContent(
		container.NewVBox(
			layout.NewSpacer(),
			widget.NewLabelWithStyle("Loading...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			layout.NewSpacer(),
			icon,
			layout.NewSpacer(),
			widget.NewProgressBarInfinite(),
			layout.NewSpacer(),
		),
	)

	uiChan := make(chan fyne.CanvasObject)
	go func() {
		uiChan <- a.setupUI()
		close(uiChan)
	}()

	go func() {
		a.window.SetContent(<-uiChan)
		a.window.CenterOnScreen()
		a.window.Resize(fyne.NewSize(800, 500))
		a.window.Show()

		splash.Close()
	}()

	splash.ShowAndRun()
}

func (a *App) setupUI() fyne.CanvasObject {
	prefs := a.app.Preferences()

	log := logrus.New(prefs.StringWithFallback("log_path", DefaultLogPath))

	a.log = log.WithField("component", "ui")
	a.log.Info("Starting up")

	var err error
	a.storage, err = storage.NewStorage(prefs.StringWithFallback("storage_path", storage.DefaultStoragePath), log)
	if err != nil {
		a.log.Fatal(err)
	}

	adbPort := prefs.IntWithFallback("adb_port", adbclient.DefaultPort)
	a.adbClient, err = adbclient.NewClient(adbPort, log)
	if err != nil {
		log.Fatal(err)
	}

	bundletoolVersion := prefs.StringWithFallback("bundletool_version", aabclient.BundleToolDefaultVersion)
	a.aabClient, err = aabclient.NewClient(bundletoolVersion, log)
	if err != nil {
		log.Fatal(err)
	}

	javaSettings := prefs.StringWithFallback("bundletool_java_settings", aabclient.BundleToolJavaSettings)
	a.aabClient.SetBundleToolJavaSettings(javaSettings)

	ctx, cancel := context.WithCancel(context.Background())
	if err := a.adbClient.Start(ctx); err != nil {
		log.Fatal(err)
	}

	if err := a.aabClient.Start(ctx); err != nil {
		log.Fatal(err)
	}

	a.aapt, err = aapt.New(a.log)
	if err != nil {
		log.Fatal(err)
	}

	a.window.SetOnClosed(func() {
		cancel()
		a.adbClient.Stop()
		a.aabClient.Stop()
		a.storage.Close()
	})

	return &container.AppTabs{Items: []*container.TabItem{
		uiMain(a.app, a.window, a.adbClient, a.aabClient, a.storage).tabItem(),
		uiBuilds(a.app, a.window, a.aabClient, a.aapt).tabItem(),
		uiSettings(a.app, a.window, a.adbClient, a.aabClient, a.storage, a.log).tabItem(),
		uiAbout(a.adbClient).tabItem(),
	}}
}

// ShowConfirmation shows a confirmation dialog with appropriate size
func (a *App) ShowInformation(title, message string, parent fyne.Window) {
	a.log.Info(message)

	dc := dialog.NewInformation(title, message, parent)
	dc.Resize(fyne.NewSize(500, 200))
	dc.Show()
}

// ShowError shows a dialog over the specified window for an application error.
func (a *App) ShowError(err error, closed func(), parent fyne.Window) {
	a.log.WithStackParams(true, 3).Error(err)

	label := widget.NewLabel(err.Error())
	label.Wrapping = fyne.TextWrapWord

	scroll := container.NewVScroll(label)

	de := dialog.NewCustom("Error", "OK", container.NewMax(scroll), parent)
	de.Resize(fyne.NewSize(500, 200))
	if closed != nil {
		de.SetOnClosed(closed)
	}

	de.Show()
}
