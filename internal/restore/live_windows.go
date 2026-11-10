//go:build windows

package restore

import (
	"time"

	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

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
