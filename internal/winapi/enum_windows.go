//go:build windows

package winapi

import (
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

var procEnumWindows = user32.NewProc("EnumWindows")

// EnumWindows uses one shared callback: windows.NewCallback has a hard limit on
// the number of callbacks per process, so creating one per call eventually panics.
var (
	enumMu  sync.Mutex
	enumOut []uintptr
	enumCB  = windows.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		enumOut = append(enumOut, hwnd)
		return 1
	})
)

// TopLevelWindows returns all top-level windows, topmost first (EnumWindows order is Z-order).
func TopLevelWindows() []uintptr {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumOut = enumOut[:0]
	_, _, _ = procEnumWindows.Call(enumCB, 0)
	out := make([]uintptr, len(enumOut))
	copy(out, enumOut)
	return out
}

// AppWindowsOf returns app windows of the given PIDs, topmost first.
func AppWindowsOf(pids map[uint32]bool) []uintptr {
	var out []uintptr
	for _, h := range TopLevelWindows() {
		if pids[WindowPID(h)] && IsAppWindow(h) {
			out = append(out, h)
		}
	}
	return out
}

// AppWindowsByExe returns app windows whose process runs the given executable, topmost first.
// Many apps (browsers, Office, Explorer) hand a new launch over to an existing
// instance, so windows have to be found by executable rather than by the launched PID.
func AppWindowsByExe(exePath string) []uintptr {
	cache := map[uint32]bool{}
	var out []uintptr
	for _, h := range TopLevelWindows() {
		if !isVisible(h) { // cheap check first: most top-level windows are hidden
			continue
		}
		pid := WindowPID(h)
		match, ok := cache[pid]
		if !ok {
			match = strings.EqualFold(ExePath(pid), exePath)
			cache[pid] = match
		}
		if match && IsAppWindow(h) {
			out = append(out, h)
		}
	}
	return out
}
