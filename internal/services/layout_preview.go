//go:build windows

package services

import (
	"Rewinder/internal/ipcapi"
	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

// SnapshotLayout returns the monitors and saved window rectangles of one snapshot.
func (s *Services) SnapshotLayout(appID, snapshotID string) (*ipcapi.SnapshotLayout, error) {
	_, full, err := s.ss.ResolveSnapshot(appID, snapshotID)
	if err != nil {
		return nil, &ipcapi.Fault{Code: "restore.snapshot", Cause: err}
	}
	return layoutOf(&full.App), nil
}

func layoutOf(app *state.AppState) *ipcapi.SnapshotLayout {
	mons := winapi.Monitors()
	out := &ipcapi.SnapshotLayout{Monitors: []ipcapi.LayoutMonitor{}, Windows: []ipcapi.LayoutWindow{}}
	var dx, dy int32 // workspace -> screen: the primary taskbar offset
	if len(mons) > 0 {
		dx, dy = mons[0].Work.Left-mons[0].Bounds.Left, mons[0].Work.Top-mons[0].Bounds.Top
	}
	for _, m := range mons {
		out.Monitors = append(out.Monitors, ipcapi.LayoutMonitor{Rect: fromWin(m.Bounds), Primary: m.Primary})
	}
	// Bottom-most first, so the frontend can draw in order and the topmost ends on top.
	for i := len(app.Windows) - 1; i >= 0; i-- {
		w := app.Windows[i]
		r := w.Rect
		if w.IsMinimized || r.Left <= -30000 {
			// A minimized window's on-screen rect is parked at -32000; show where it goes back to.
			n := w.NormalRect
			r = state.Rect{Left: n.Left + dx, Top: n.Top + dy, Right: n.Right + dx, Bottom: n.Bottom + dy}
		}
		out.Windows = append(out.Windows, ipcapi.LayoutWindow{
			Rect:       ipcapi.LayoutRect{X: r.Left, Y: r.Top, W: r.Right - r.Left, H: r.Bottom - r.Top},
			Title:      w.Title,
			Minimized:  w.IsMinimized,
			Maximized:  w.IsMaximized,
			Foreground: w.IsForeground,
		})
	}
	return out
}

func fromWin(r winapi.Rect) ipcapi.LayoutRect {
	return ipcapi.LayoutRect{X: r.Left, Y: r.Top, W: r.Right - r.Left, H: r.Bottom - r.Top}
}
