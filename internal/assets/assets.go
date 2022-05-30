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
//go:generate fyne bundle -package assets -o bundled.go -append icon_logs.svg
//go:generate fyne bundle -package assets -o bundled.go -append icon_screenshot.svg
//go:generate fyne bundle -package assets -o bundled.go -append icon_video.svg
//go:generate fyne bundle -package assets -o bundled.go -append icon_send.svg
//go:generate fyne bundle -package assets -o bundled.go -append icon_builds.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_apk.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_aab.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_abi.svg
//go:generate fyne bundle -package assets -o bundled.go -append icon_manifest.svg

// IconApp is the icon for the application
var AppIcon = resourceIconappPng

// MainTabIcon is the icon for the main tab
var MainTabIcon = resourceIconmainPng

// BuildsTabIcon is the icon for the builds tab
var BuildsTabIcon = resourceIconbuildsPng

// AboutTabIcon is the icon for the about tab
var AboutTabIcon = resourceIconaboutPng

// SettingsTabIcon is the icon for the settings tab
var SettingsTabIcon = resourceIconsettingsPng

// InstallIcon is the icon for the install button
var InstallIcon = resourceIconinstallPng

// LogsIcon is the icon for the logs button
var LogsIcon = resourceIconlogsSvg

// ScreenshotIcon is the icon for the screenshot button
var ScreenshotIcon = resourceIconscreenshotSvg

// VideoIcon is the icon for the video button
var VideoIcon = resourceIconvideoSvg

// SendIcon is the icon for the send button
var SendIcon = resourceIconsendSvg

// IconAPK is the icon for the APK file
var APKIcon = resourceIconapkPng

// IconAAB is the icon for the AAB file
var AABIcon = resourceIconaabPng

// IconABI is the icon for the ABI file
var ABIIcon = resourceIconabiSvg

// IconManifest is the icon for the manifest button
var ManifestIcon = resourceIconmanifestSvg

// StatusIcons are the icons for the status of the device
var StatusIcons map[string]*fyne.StaticResource = map[string]*fyne.StaticResource{
	"online":       resourceIconconnectedPng,
	"disconnected": resourceIcondisconnectedPng,
	"offline":      resourceIconofflinePng,
	"unauthorized": resourceIconunauthorizedPng,
	"invalid":      resourceIconunknownPng,
}
