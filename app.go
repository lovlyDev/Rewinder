package main

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/services"
	"Rewinder/internal/settings"
	"Rewinder/internal/winapi"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	svc         *services.Services
	start       sync.Once
	settingsMu  sync.Mutex
	preferences settings.Settings
	windowMu    sync.Mutex
	overlay     bool
	previous    struct {
		w, h, x, y int
		maximized  bool
	}
}

func NewApp() *App {
	s, err := settings.Load(settings.Directory())
	if err != nil {
		log.Printf("settings: %v", err)
	}
	return &App{preferences: s}
}

// OverlayRequest tells the frontend which app the user was in when they pressed the hotkey.
type OverlayRequest struct {
	AppID string `json:"appID"`
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.start.Do(func() {
		a.svc = services.New(services.Dependencies{
			Hotkey:   a.preferences.Hotkey,
			Language: a.preferenceLanguage(),
			EmitEvent: func(name string, data any) {
				runtime.EventsEmit(ctx, name, data)
				if name == "onExitRequested" {
					a.ExitApp()
				}
			},
			OnOverlayRequested: func(appID string) {
				// Show last: Wails' show path is what moves keyboard focus into the
				// WebView, and changing window styles afterwards can lose it again.
				runtime.EventsEmit(ctx, "onOverlayOpenRequested", OverlayRequest{AppID: appID})
				a.prepareOverlay()
				runtime.EventsEmit(ctx, "onWindowVisibilityChanged", true)
				runtime.WindowShow(ctx)
				// Windows often refuses Wails' own SetForegroundWindow (focus-stealing
				// guard); then the overlay is visible but keys still go to the other app.
				if !winapi.WaitAndForceForeground(uint32(os.Getpid()), time.Second) {
					log.Printf("rewinder: could not bring the overlay to the foreground")
					return
				}
				// Now that the window is active, show again: Wails' SetFocus is what hands
				// keyboard focus to the WebView, and it only works for the active window.
				runtime.WindowShow(ctx)
			},
			ShowTimelineWindow: a.ShowTimelineWindow,
		})
		a.svc.Start(ctx)
	})
}

func (a *App) shutdown(ctx context.Context) {
	if a.svc != nil {
		a.svc.Stop()
	}
}

// onSecondInstance brings the running window forward when Rewinder is started again.
func (a *App) onSecondInstance() {
	a.ShowTimelineWindow()
}

var errNotReady = errors.New("backend.unavailable")

func (a *App) GetApps() ([]ipcapi.AppSummary, error) {
	if a.svc == nil {
		return nil, errNotReady
	}
	return a.svc.GetApps(), nil
}

func (a *App) GetTimeline(appID string) ([]ipcapi.SnapshotMeta, error) {
	if a.svc == nil {
		return nil, errNotReady
	}
	return a.svc.GetTimeline(appID), nil
}

func (a *App) Restore(appID string, snapshotID string) error {
	if a.svc == nil {
		return errNotReady
	}
	return a.svc.Restore(appID, snapshotID)
}

func (a *App) PauseTracking(appID *string) error {
	if a.svc == nil {
		return errNotReady
	}
	a.svc.PauseTracking(appID)
	return nil
}

func (a *App) ResumeTracking(appID *string) error {
	if a.svc == nil {
		return errNotReady
	}
	a.svc.ResumeTracking(appID)
	return nil
}

func (a *App) ShowTimelineWindow() {
	if a.ctx == nil {
		return
	}
	a.restoreGeometry()
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", true)
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
	runtime.EventsEmit(a.ctx, "onShowTimelineRequested", nil)
}

// HideOverlay restores the normal geometry before hiding the window.
func (a *App) HideOverlay() {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", false)
	runtime.WindowHide(a.ctx)
	a.restoreGeometry()
}

func (a *App) ExitApp() {
	if a.svc != nil {
		a.svc.Stop()
	}
	os.Exit(0)
}

// LogUI lets the frontend write to the app log.
func (a *App) LogUI(msg string) {
	log.Printf("ui: %s", msg)
}

func (a *App) GetAutostart() bool {
	return services.IsAutostartEnabled()
}

func (a *App) SetAutostart(enabled bool) error {
	return services.SetAutostart(enabled)
}

// Wails 2.11 SetPos takes a work-area-relative offset, while Pos is absolute.
// Keep all saved coordinates absolute and convert at this one boundary.
func (a *App) setAbsolutePosition(x, y int) {
	work := winapi.MonitorWork(winapi.OwnWindow())
	runtime.WindowSetPosition(a.ctx, x-int(work.Left), y-int(work.Top))
}
func (a *App) prepareOverlay() {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	runtime.WindowUnminimise(a.ctx)
	if !a.overlay {
		a.previous.maximized = runtime.WindowIsMaximised(a.ctx)
		runtime.WindowUnmaximise(a.ctx)
		a.previous.w, a.previous.h = runtime.WindowGetSize(a.ctx)
		a.previous.x, a.previous.y = runtime.WindowGetPosition(a.ctx)
		a.overlay = true
	} else {
		runtime.WindowUnmaximise(a.ctx)
	}
	runtime.WindowSetMinSize(a.ctx, 0, 0)
	if work, ok := winapi.CursorMonitorWork(); ok {
		// Move to the target monitor first, so SetSize uses that monitor's DPI.
		a.setAbsolutePosition(int(work.Left)+int(work.Right-work.Left)/2-380, int(work.Top)+int(work.Bottom-work.Top)/2-120)
		runtime.WindowSetSize(a.ctx, 760, 240)
		rect := winapi.WindowRect(winapi.OwnWindow())
		a.setAbsolutePosition(int(work.Left)+int(work.Right-work.Left-rect.Right+rect.Left)/2, int(work.Top)+int(work.Bottom-work.Top-rect.Bottom+rect.Top)/2)
	} else {
		runtime.WindowSetSize(a.ctx, 760, 240)
		runtime.WindowCenter(a.ctx)
	}
	runtime.WindowSetAlwaysOnTop(a.ctx, true)
}
func (a *App) restoreGeometry() {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	runtime.WindowSetAlwaysOnTop(a.ctx, false)
	if !a.overlay {
		return
	}
	a.setAbsolutePosition(a.previous.x, a.previous.y)
	runtime.WindowSetSize(a.ctx, a.previous.w, a.previous.h)
	runtime.WindowSetMinSize(a.ctx, 760, 500)
	if a.previous.maximized {
		runtime.WindowMaximise(a.ctx)
	}
	a.overlay = false
}
func (a *App) preferenceLanguage() string {
	lang := a.preferences.Language
	if lang == "system" {
		return winapi.SystemLanguage()
	}
	return lang
}
func (a *App) MinimiseWindow() {
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", false)
	runtime.WindowMinimise(a.ctx)
}
func (a *App) HideWindow() {
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", false)
	runtime.WindowHide(a.ctx)
}

func (a *App) GetWindowVisible() bool {
	return winapi.OwnWindowVisible() && !runtime.WindowIsMinimised(a.ctx)
}

func (a *App) GetAppSummary(appID string) (*ipcapi.AppSummary, error) {
	if a.svc == nil {
		return nil, errNotReady
	}
	for _, app := range a.svc.GetApps() {
		if app.AppID == appID {
			return &app, nil
		}
	}
	return nil, nil
}
