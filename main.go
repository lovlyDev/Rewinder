package main

import (
	"context"
	"embed"
	"log"
	"os"

	"Rewinder/internal/services"
	"Rewinder/internal/update"
	"Rewinder/internal/winapi"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	services.SetupLogFile()
	waitForPreviousInstance(os.Args[1:])
	if exe, err := os.Executable(); err == nil {
		update.CleanupOld(exe) // leftovers of a previous self-update
	}
	app := NewApp()
	app.startHidden = services.StartedByAutostart(os.Args[1:])

	// "Run as administrator": start the elevated copy and leave. If the user says
	// no to UAC, carry on without rights; the settings page explains it.
	if app.preferences.RunAsAdmin && !winapi.SelfElevated() {
		if err := restartElevated(); err == nil {
			return
		} else {
			log.Printf("rewinder: not elevated: %v", err)
			app.elevationDeclined = true
		}
	}
	if err := services.SyncAutostart(app.preferences.RunAsAdmin && winapi.SelfElevated()); err != nil {
		log.Printf("rewinder: autostart: %v", err)
	}

	err := wails.Run(&options.App{
		Title:             "Rewinder",
		Frameless:         true,
		MinWidth:          760,
		MinHeight:         500,
		Width:             1024,
		Height:            768,
		StartHidden:       app.startHidden,
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
