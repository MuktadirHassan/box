package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

func main() {
	application := NewApp()
	if err := wails.Run(&options.App{
		Title:  "Box",
		Width:  960,
		Height: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: application.startup,
		Bind:      []interface{}{application},
	}); err != nil {
		log.Fatal(err)
	}
}
