//go:build windows

package restore

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"Rewinder/internal/plugins"
	"Rewinder/internal/snapshot"
	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

type ProgressFn func(stage string, percent int, msg string)

type Engine struct {
	cap     *state.CaptureEngine
	plugins *plugins.Registry

	LaunchTimeout time.Duration // how long to wait for a relaunched app's windows
	ReopenTimeout time.Duration // how long to wait for reopened documents/folders
	SettleDelay   time.Duration // apps often move their own window right after start
}

func NewEngine(cap *state.CaptureEngine, reg *plugins.Registry) *Engine {
	return &Engine{
		cap:           cap,
		plugins:       reg,
		LaunchTimeout: 15 * time.Second,
		ReopenTimeout: 8 * time.Second,
		SettleDelay:   800 * time.Millisecond,
	}
}

// Result says what a restore actually did.
type Result struct {
	Relaunched bool
	Reopened   int
	Matched    int
	Saved      int
}

func (e *Engine) RestoreSnapshot(progress ProgressFn, _ *snapshot.Snapshot, full *snapshot.FullSnapshot) (Result, error) {
	var res Result
	if full == nil {
		return res, errors.New("no snapshot state")
	}
	saved := full.App.Clone()
	res.Saved = len(saved.Windows)
	if saved.ExecutablePath == "" {
		return res, errors.New("snapshot has no executable path")
	}

	progress("find_instance", 10, "Looking for running instance")
	live := e.liveState(&saved)

	if len(live.Windows) == 0 {
		progress("relaunch", 25, "Starting "+saved.ExecutablePath)
		if err := e.relaunch(&saved); err != nil {
			return res, fmt.Errorf("could not start the app: %w", err)
		}
		res.Relaunched = true
		progress("wait_windows", 40, "Waiting for windows")
		live = e.waitForWindows(&saved, e.LaunchTimeout)
		if len(live.Windows) == 0 {
			return res, errors.New("the app started but no window appeared")
		}
	} else if n := e.plugins.Reopen(&saved, live); n > 0 {
		res.Reopened = n
		progress("reopen", 30, fmt.Sprintf("Reopening %d item(s)", n))
		live = e.waitForWindows(&saved, e.ReopenTimeout)
	}

	if len(saved.Windows) == 0 {
		return res, nil
	}

	progress("restore_windows", 60, "Restoring window positions")
	moved, blocked := e.apply(&saved, live)
	if blocked {
		return res, errors.New("Windows refused to move the app's windows; if the app runs as administrator, run Rewinder as administrator too")
	}
	if moved == 0 {
		return res, errors.New("none of the saved windows could be found")
	}
	res.Matched = moved
	if res.Relaunched || res.Reopened > 0 {
		time.Sleep(e.SettleDelay)
		if again, _ := e.apply(&saved, e.liveState(&saved)); again > 0 {
			res.Matched = again
		}
	}

	progress("restore_focus", 90, "Restoring focus")
	return res, nil
}

// liveState captures every current app window of the saved executable,
// whatever process owns it.
func (e *Engine) liveState(saved *state.AppState) *state.AppState {
	hwnds := winapi.AppWindowsByExe(saved.ExecutablePath)
	if len(hwnds) == 0 {
		return &state.AppState{ExecutablePath: saved.ExecutablePath, PluginData: map[string]state.PluginState{}}
	}
	byPID := map[uint32]int{}
	for _, h := range hwnds {
		byPID[winapi.WindowPID(h)]++
	}
	var main uint32
	for pid, n := range byPID {
		if n > byPID[main] {
			main = pid
		}
	}
	live, err := e.cap.CaptureProcess(int(main))
	if err != nil {
		live = &state.AppState{ExecutablePath: saved.ExecutablePath, PID: int(main), PluginData: map[string]state.PluginState{}}
	}
	fg := winapi.ForegroundWindow()
	live.Windows = live.Windows[:0]
	for i, h := range hwnds {
		live.Windows = append(live.Windows, state.DescribeWindow(h, i, h == fg))
	}
	e.plugins.Capture(live)
	return live
}

func (e *Engine) relaunch(saved *state.AppState) error {
	if handled, err := e.plugins.Relaunch(saved); handled {
		return err
	}
	if saved.AppUserModelID != "" {
		return winapi.LaunchPackaged(saved.AppUserModelID)
	}
	if _, err := winapi.Launch(saved.ExecutablePath, saved.CommandLine, saved.WorkingDir); err != nil {
		// e.g. ERROR_ELEVATION_REQUIRED: let the shell handle it
		return winapi.ShellOpen(saved.ExecutablePath)
	}
	return nil
}

// waitForWindows polls until the app shows as many windows as were saved and the
// count holds steady, or the timeout passes. It returns the last capture.
func (e *Engine) waitForWindows(saved *state.AppState, timeout time.Duration) *state.AppState {
	want := len(saved.Windows)
	if want == 0 {
		want = 1
	}
	deadline := time.Now().Add(timeout)
	live := e.liveState(saved)
	stable := 0
	for time.Now().Before(deadline) {
		time.Sleep(250 * time.Millisecond)
		next := e.liveState(saved)
		if len(next.Windows) >= want && len(next.Windows) == len(live.Windows) {
			stable++
			if stable >= 2 {
				return next
			}
		} else {
			stable = 0
		}
		live = next
	}
	return live
}

// apply puts matched windows back. It returns how many windows were actually
// moved, and blocked=true when windows matched but Windows refused every change
// (typically because the target app runs as administrator and we do not).
func (e *Engine) apply(saved *state.AppState, live *state.AppState) (moved int, blocked bool) {
	m := MatchWindows(saved.Windows, live.Windows)
	if len(m) == 0 {
		return 0, false
	}
	type pair struct {
		saved state.WindowState
		hwnd  uintptr
	}
	var pairs []pair
	for si, li := range m {
		pairs = append(pairs, pair{saved: saved.Windows[si], hwnd: live.Windows[li].HWND})
	}
	for _, p := range pairs {
		if err := winapi.SetPlacement(p.hwnd, p.saved.Placement()); err == nil {
			moved++
		}
	}
	if moved == 0 {
		return 0, true
	}

	// Raise bottom-most first so the saved topmost window ends up on top.
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].saved.ZOrder > pairs[j].saved.ZOrder })
	for _, p := range pairs {
		if !p.saved.IsMinimized {
			winapi.BringToTop(p.hwnd)
		}
	}
	focus := uintptr(0)
	for _, p := range pairs {
		if p.saved.IsForeground && !p.saved.IsMinimized {
			focus = p.hwnd
		}
	}
	if focus == 0 {
		for i := len(pairs) - 1; i >= 0; i-- {
			if !pairs[i].saved.IsMinimized {
				focus = pairs[i].hwnd
				break
			}
		}
	}
	if focus != 0 {
		winapi.SetForeground(focus)
	}
	return moved, false
}
