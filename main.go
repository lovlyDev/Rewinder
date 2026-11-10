package main

import (
	"context"
	"embed"
	"os"

	"Rewinder/internal/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	services.SetupLogFile()
	app := NewApp()

	err := wails.Run(&options.App{
		Title:             "Rewinder",
		Frameless:         true,
		MinWidth:          760,
		MinHeight:         500,
		Width:             1024,
		Height:            768,
		StartHidden:       services.StartedByAutostart(os.Args[1:]),
		HideWindowOnClose: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 12, A: 255},
		Windows: &windows.Options{
			// The UI is a few static panels; software rendering is plenty and
			// saves WebView2's whole GPU process.
			WebviewGpuIsDisabled: true,
		},
		OnStartup:  app.startup,
		OnShutdown: app.shutdown,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			app.HideWindow()
			return true
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
