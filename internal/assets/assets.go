package assets

import "fyne.io/fyne/v2"

//go:generate fyne bundle -package assets -o bundled.go icon_app.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_connected.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_disconnected.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_offline.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_unauthorized.png
//go:generate fyne bundle -package assets -o bundled.go -append icon_unknown.png

// IconApp is the icon for the application
var AppIcon = resourceIconappPng

// StatusIcons are the icons for the status of the device
var StatusIcons map[string]*fyne.StaticResource = map[string]*fyne.StaticResource{
	"online":       resourceIconconnectedPng,
	"disconnected": resourceIcondisconnectedPng,
	"offline":      resourceIconofflinePng,
	"unauthorized": resourceIconunauthorizedPng,
	"invalid":      resourceIconunknownPng,
}
