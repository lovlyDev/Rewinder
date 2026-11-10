package restore

import "Rewinder/internal/state"

// Minimum part of a window that must lie on some monitor for it to count as
// reachable: enough of the title bar to grab it with the mouse.
const (
	minVisibleW = 120
	minVisibleH = 40
)

// FitToMonitors keeps a saved window rectangle reachable on the monitors that
// exist now. works are the monitors' work areas in the same coordinate space as
// r, the primary monitor first. When too little of r lies on any monitor (its
// monitor was unplugged, or the layout changed), r is moved onto the primary
// monitor: same size, shrunk to fit if needed, centred. moved reports whether r
// changed.
func FitToMonitors(r state.Rect, works []state.Rect) (state.Rect, bool) {
	if len(works) == 0 || r.Right <= r.Left || r.Bottom <= r.Top {
		return r, false
	}
	for _, w := range works {
		ix := min32(r.Right, w.Right) - max32(r.Left, w.Left)
		iy := min32(r.Bottom, w.Bottom) - max32(r.Top, w.Top)
		// The top edge (title bar) must be on screen, not only some corner.
		if ix >= minVisibleW && iy >= minVisibleH && r.Top >= w.Top-8 && r.Top < w.Bottom-minVisibleH/2 {
			return r, false
		}
	}
	w := works[0]
	width := min32(r.Right-r.Left, w.Right-w.Left)
	height := min32(r.Bottom-r.Top, w.Bottom-w.Top)
	left := w.Left + (w.Right-w.Left-width)/2
	top := w.Top + (w.Bottom-w.Top-height)/2
	return state.Rect{Left: left, Top: top, Right: left + width, Bottom: top + height}, true
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
