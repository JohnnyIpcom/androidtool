package assets

import "fyne.io/fyne/v2"

//go:generate fyne bundle -package assets -o bundled.go icon_app.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_connected.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_disconnected.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_offline.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_unauthorized.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_unknown.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_about.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_settings.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_main.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_install.png

// IconApp is the icon for the application
var AppIcon = resourceIconappPng

// MainTabIcon is the icon for the main tab
var MainTabIcon = resourceIconmainPng

// AboutTabIcon is the icon for the about tab
var AboutTabIcon = resourceIconaboutPng

// SettingsTabIcon is the icon for the settings tab
var SettingsTabIcon = resourceIconsettingsPng

// InstallIcon is the icon for the install button
var InstallIcon = resourceIconinstallPng

// StatusIcons are the icons for the status of the device
var StatusIcons map[string]*fyne.StaticResource = map[string]*fyne.StaticResource{
	"online":       resourceIconconnectedPng,
	"disconnected": resourceIcondisconnectedPng,
	"offline":      resourceIconofflinePng,
	"unauthorized": resourceIconunauthorizedPng,
	"invalid":      resourceIconunknownPng,
}
