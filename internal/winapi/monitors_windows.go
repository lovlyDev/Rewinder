//go:build windows

package winapi

import (
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var procEnumDisplayMonitors = user32.NewProc("EnumDisplayMonitors")

const monitorInfoPrimary = 1

// Monitor is one display, in screen coordinates.
type Monitor struct {
	Bounds  Rect `json:"bounds"`
	Work    Rect `json:"work"` // without the taskbar
	Primary bool `json:"primary"`
}

// One shared callback, like EnumWindows: windows.NewCallback slots are limited.
var (
	monMu  sync.Mutex
	monOut []uintptr
	monCB  = windows.NewCallback(func(hmon, _, _, _ uintptr) uintptr {
		monOut = append(monOut, hmon)
		return 1
	})
)

// Monitors lists the displays connected now, the primary one first.
func Monitors() []Monitor {
	monMu.Lock()
	monOut = monOut[:0]
	_, _, _ = procEnumDisplayMonitors.Call(0, 0, monCB, 0)
	handles := append([]uintptr(nil), monOut...)
	monMu.Unlock()

	out := make([]Monitor, 0, len(handles))
	for _, h := range handles {
		var mi monitorInfoEx
		mi.CbSize = uint32(unsafe.Sizeof(mi))
		if r, _, _ := procGetMonitorInfoW.Call(h, uintptr(unsafe.Pointer(&mi))); r == 0 {
			continue
		}
		m := Monitor{Bounds: mi.Monitor, Work: mi.Work, Primary: mi.Flags&monitorInfoPrimary != 0}
		if m.Primary {
			out = append([]Monitor{m}, out...)
		} else {
			out = append(out, m)
		}
	}
	return out
}

// WorkspaceWorkAreas returns the monitors' work areas in the "workspace"
// coordinates WINDOWPLACEMENT uses: screen coordinates shifted by the primary
// monitor's taskbar when it sits on the top or left edge. Primary first.
func WorkspaceWorkAreas() []Rect {
	mons := Monitors()
	if len(mons) == 0 {
		return nil
	}
	dx := mons[0].Work.Left - mons[0].Bounds.Left
	dy := mons[0].Work.Top - mons[0].Bounds.Top
	out := make([]Rect, len(mons))
	for i, m := range mons {
		out[i] = Rect{Left: m.Work.Left - dx, Top: m.Work.Top - dy, Right: m.Work.Right - dx, Bottom: m.Work.Bottom - dy}
	}
	return out
}
