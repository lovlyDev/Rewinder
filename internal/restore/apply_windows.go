//go:build windows

package restore

import (
	"fmt"
	"log"
	"path/filepath"
	"sort"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

type placedWindow struct {
	saved state.WindowState
	hwnd  uintptr
}

// apply puts matched windows back and returns how many were actually moved.
// It returns an error when windows matched but Windows refused every change.
func (e *Engine) apply(saved *state.AppState, live *state.AppState) (int, error) {
	m := MatchWindows(saved.Windows, live.Windows)
	if len(m) == 0 {
		return 0, nil
	}
	pairs := make([]placedWindow, 0, len(m))
	for si, li := range m {
		pairs = append(pairs, placedWindow{saved: saved.Windows[si], hwnd: live.Windows[li].HWND})
	}
	works := workAreas()
	moved := 0
	var firstErr error
	for _, p := range pairs {
		pl := p.saved.Placement()
		if fitted, changed := FitToMonitors(fromWinRect(pl.NormalRect), works); changed {
			// Its monitor is gone (or the layout changed): bring it onto the primary one.
			log.Printf("rewinder: window %x was off-screen at %+v, moved to %+v", p.hwnd, pl.NormalRect, fitted)
			pl.NormalRect = toWinRect(fitted)
		}
		if err := winapi.SetPlacement(p.hwnd, pl); err == nil {
			moved++
		} else if firstErr == nil {
			firstErr = err
		}
	}
	if moved == 0 {
		return 0, refusedError(saved, pairs[0].hwnd, firstErr)
	}
	raiseAndFocus(pairs)
	return moved, nil
}

func workAreas() []state.Rect {
	ws := winapi.WorkspaceWorkAreas()
	out := make([]state.Rect, len(ws))
	for i, w := range ws {
		out[i] = fromWinRect(w)
	}
	return out
}

func fromWinRect(r winapi.Rect) state.Rect {
	return state.Rect{Left: r.Left, Top: r.Top, Right: r.Right, Bottom: r.Bottom}
}

func toWinRect(r state.Rect) winapi.Rect {
	return winapi.Rect{Left: r.Left, Top: r.Top, Right: r.Right, Bottom: r.Bottom}
}

// refusedError explains why Windows would not move the windows. The usual reason
// is UIPI: a normal process may not touch the windows of an elevated one.
func refusedError(saved *state.AppState, hwnd uintptr, err error) error {
	name := filepath.Base(saved.ExecutablePath)
	if winapi.IsElevated(winapi.WindowPID(hwnd)) {
		return &ipcapi.Fault{Code: "restore.elevated", Params: map[string]any{"app": name}, Cause: fmt.Errorf("%s is elevated: %w", name, err)}
	}
	return &ipcapi.Fault{Code: "restore.refused", Params: map[string]any{"app": name}, Cause: err}
}

// raiseAndFocus restores Z-order and gives focus to the window that had it.
func raiseAndFocus(pairs []placedWindow) {
	// Raise bottom-most first so the saved topmost window ends up on top.
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].saved.ZOrder > pairs[j].saved.ZOrder })
	focus := uintptr(0)
	for _, p := range pairs {
		if p.saved.IsMinimized {
			continue
		}
		winapi.BringToTop(p.hwnd)
		if p.saved.IsForeground {
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
}
