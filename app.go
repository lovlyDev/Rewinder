package main

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/services"
	"Rewinder/internal/settings"
	"Rewinder/internal/winapi"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx               context.Context
	svc               *services.Services
	start             sync.Once
	settingsMu        sync.Mutex
	preferences       settings.Settings
	startHidden       bool        // launched by autostart, window not shown
	elevationDeclined bool        // "run as administrator" is on but UAC was declined at start
	hiddenOnce        atomic.Bool // the window has been hidden or minimised at least once
	// overlayFromWindow: the overlay was opened from the main window, so closing it returns there
	overlayFromWindow atomic.Bool
	windowMu          sync.Mutex
	overlay           bool
	previous          struct {
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
			OnOverlayRequested: a.OpenOverlay,
			ShowTimelineWindow: a.ShowTimelineWindow,
		})
		a.svc.Start(ctx)
		go a.autoCheckUpdate()
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
	if a.overlayFromWindow.Swap(false) {
		a.ShowTimelineWindow()
		return
	}
	a.hiddenOnce.Store(true)
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
	a.settingsMu.Lock()
	elevated := a.preferences.RunAsAdmin && winapi.SelfElevated()
	a.settingsMu.Unlock()
	if err := services.SetAutostart(enabled, elevated); err != nil {
		log.Printf("rewinder: autostart: %v", err)
		if err.Error() == "autostart.admin" {
			return err
		}
		return errors.New("saveFailed")
	}
	return nil
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

// OpenOverlay turns the window into the compact quick-restore overlay for appID.
// The hotkey calls it, and so does the "Quick restore" button in the main window.
func (a *App) OpenOverlay(appID string) {
	if a.ctx == nil {
		return
	}
	// Show last: Wails' show path is what moves keyboard focus into the
	// WebView, and changing window styles afterwards can lose it again.
	// Opened from the main window: closing the overlay goes back there. Opened by
	// the hotkey from another app: closing just hides Rewinder again.
	a.overlayFromWindow.Store(a.GetWindowVisible())
	runtime.EventsEmit(a.ctx, "onOverlayOpenRequested", OverlayRequest{AppID: appID})
	a.prepareOverlay()
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", true)
	runtime.WindowShow(a.ctx)
	// Windows often refuses Wails' own SetForegroundWindow (focus-stealing
	// guard); then the overlay is visible but keys still go to the other app.
	if !winapi.WaitAndForceForeground(uint32(os.Getpid()), time.Second) {
		log.Printf("rewinder: could not bring the overlay to the foreground")
		return
	}
	// Now that the window is active, show again: Wails' SetFocus is what hands
	// keyboard focus to the WebView, and it only works for the active window.
	runtime.WindowShow(a.ctx)
	winapi.FocusContent(winapi.VisibleWindowOf(uint32(os.Getpid())))
}

// FocusOverlay is called by the overlay once it is rendered. The UI is not drawn
// while the window is hidden, so the focus handed over when the window was shown
// can land before the overlay exists; activating again here makes keys reliable.
func (a *App) FocusOverlay() {
	if a.ctx == nil {
		return
	}
	if !winapi.WaitAndForceForeground(uint32(os.Getpid()), 500*time.Millisecond) {
		log.Printf("rewinder: overlay could not take the foreground")
		return
	}
	runtime.WindowShow(a.ctx)
	winapi.FocusContent(winapi.VisibleWindowOf(uint32(os.Getpid())))
}

func (a *App) MinimiseWindow() {
	a.hiddenOnce.Store(true)
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", false)
	runtime.WindowMinimise(a.ctx)
}
func (a *App) HideWindow() {
	a.hiddenOnce.Store(true)
	runtime.EventsEmit(a.ctx, "onWindowVisibilityChanged", false)
	runtime.WindowHide(a.ctx)
}

func (a *App) GetWindowVisible() bool {
	if winapi.OwnWindowVisible() {
		return !runtime.WindowIsMinimised(a.ctx)
	}
	// The UI asks this while starting up, before Wails has shown the window. A
	// normal launch is about to become visible; only an autostart launch stays hidden.
	return !a.startHidden && !a.hiddenOnce.Load()
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

// GetSnapshotLayout returns monitors and window rectangles for the snapshot preview.
func (a *App) GetSnapshotLayout(appID, snapshotID string) (*ipcapi.SnapshotLayout, error) {
	if a.svc == nil {
		return nil, errNotReady
	}
	l, err := a.svc.SnapshotLayout(appID, snapshotID)
	if err != nil {
		code, _ := ipcapi.ErrorCode(err)
		return nil, errors.New(code)
	}
	return l, nil
}

// GetExclusions lists executables that are not tracked.
func (a *App) GetExclusions() []string {
	if a.svc == nil {
		return []string{}
	}
	return a.svc.Exclusions()
}

// SetExcluded stops or resumes tracking an executable ("chrome.exe" or a full path).
func (a *App) SetExcluded(exe string, excluded bool) error {
	if a.svc == nil {
		return errNotReady
	}
	if err := a.svc.SetExcluded(exe, excluded); err != nil {
		log.Printf("rewinder: exclusions: %v", err)
		return err
	}
	return nil
}

// DeleteAppHistory removes all snapshots of an app.
func (a *App) DeleteAppHistory(appID string) error {
	if a.svc == nil {
		return errNotReady
	}
	if err := a.svc.DeleteAppHistory(appID); err != nil {
		log.Printf("rewinder: deleting history of %s: %v", appID, err)
		return errors.New("history.delete")
	}
	return nil
}

func (a *App) GetLayouts() []services.LayoutSummary {
	if a.svc == nil {
		return []services.LayoutSummary{}
	}
	return a.svc.Layouts()
}

// SaveLayout saves where every open application's windows are now, under name.
func (a *App) SaveLayout(name string) (services.LayoutSummary, error) {
	if a.svc == nil {
		return services.LayoutSummary{}, errNotReady
	}
	return a.svc.SaveLayout(name)
}

func (a *App) RenameLayout(id, name string) error {
	if a.svc == nil {
		return errNotReady
	}
	return a.svc.RenameLayout(id, name)
}

func (a *App) DeleteLayout(id string) error {
	if a.svc == nil {
		return errNotReady
	}
	return a.svc.DeleteLayout(id)
}

func (a *App) MoveLayout(id string, to int) error {
	if a.svc == nil {
		return errNotReady
	}
	return a.svc.MoveLayout(id, to)
}

func (a *App) RestoreLayout(id string) (services.LayoutResult, error) {
	if a.svc == nil {
		return services.LayoutResult{}, errNotReady
	}
	return a.svc.RestoreLayout(id)
}

func (a *App) GetLayoutPreview(id string) (*ipcapi.SnapshotLayout, error) {
	if a.svc == nil {
		return nil, errNotReady
	}
	return a.svc.LayoutPreview(id)
}
