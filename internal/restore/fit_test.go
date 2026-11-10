package restore

import (
	"testing"

	"Rewinder/internal/state"
)

func rect(l, t, r, b int32) state.Rect { return state.Rect{Left: l, Top: t, Right: r, Bottom: b} }

func TestFitToMonitors(t *testing.T) {
	primary := rect(0, 0, 1920, 1040)
	right := rect(1920, 0, 4480, 1400)

	cases := []struct {
		name  string
		in    state.Rect
		works []state.Rect
		want  state.Rect
		moved bool
	}{
		{"on primary", rect(100, 100, 900, 700), []state.Rect{primary}, rect(100, 100, 900, 700), false},
		{"on second monitor", rect(2000, 50, 3000, 800), []state.Rect{primary, right}, rect(2000, 50, 3000, 800), false},
		{"second monitor unplugged", rect(2000, 50, 3000, 800), []state.Rect{primary}, rect(460, 145, 1460, 895), true},
		{"partly off the edge but grabbable", rect(1500, 200, 2300, 700), []state.Rect{primary}, rect(1500, 200, 2300, 700), false},
		{"only a sliver left", rect(1880, 200, 2680, 700), []state.Rect{primary}, rect(560, 270, 1360, 770), true},
		{"title bar above the screen", rect(200, -600, 1000, 30), []state.Rect{primary}, rect(560, 205, 1360, 835), true},
		{"bigger than the monitor", rect(-5000, 0, 0, 3000), []state.Rect{primary}, rect(0, 0, 1920, 1040), true},
		{"no monitors known", rect(9000, 9000, 9100, 9100), nil, rect(9000, 9000, 9100, 9100), false},
	}
	for _, c := range cases {
		got, moved := FitToMonitors(c.in, c.works)
		if got != c.want || moved != c.moved {
			t.Errorf("%s: got %+v moved=%v, want %+v moved=%v", c.name, got, moved, c.want, c.moved)
		}
	}
}
