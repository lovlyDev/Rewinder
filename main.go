package main

import (
	"context"
	"embed"
	"os"

	"Rewinder/internal/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	services.SetupLogFile()
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Rewinder",
		Width:             1024,
		Height:            768,
		StartHidden:       services.StartedByAutostart(os.Args[1:]),
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			return false
		},
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "rewinder-7c1f3f2e-app-time-machine",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				app.onSecondInstance()
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
