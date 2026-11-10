//go:build windows

package winapi

import (
	"runtime"
	"time"

	"golang.org/x/sys/windows"
)

var (
	procAttachThreadInput = user32.NewProc("AttachThreadInput")
	procBringWindowToTop  = user32.NewProc("BringWindowToTop")
)

// VisibleWindowOf returns the topmost visible, unowned top-level window of pid, or 0.
func VisibleWindowOf(pid uint32) uintptr {
	for _, h := range TopLevelWindows() {
		if WindowPID(h) == pid && isVisible(h) && owner(h) == 0 {
			return h
		}
	}
	return 0
}

// ForceForeground makes hwnd the foreground window. A plain SetForegroundWindow is
// silently refused by the focus-stealing guard when another app holds the
// foreground; attaching to the foreground thread's input queue is the documented
// way around it for a window the user asked for (here: by pressing the hotkey).
func ForceForeground(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	if ForegroundWindow() == hwnd {
		return true
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	self := windows.GetCurrentThreadId()
	fg := ForegroundWindow()
	var fgThread uint32
	if fg != 0 {
		r, _, _ := procGetWindowThreadProcessId.Call(fg, 0)
		fgThread = uint32(r)
	}
	if fgThread != 0 && fgThread != self {
		_, _, _ = procAttachThreadInput.Call(uintptr(self), uintptr(fgThread), 1)
		defer procAttachThreadInput.Call(uintptr(self), uintptr(fgThread), 0)
	}
	_, _, _ = procBringWindowToTop.Call(hwnd)
	SetForeground(hwnd)
	return ForegroundWindow() == hwnd
}

// WaitAndForceForeground waits up to timeout for pid to show a window and makes it
// the foreground window. It reports whether that worked.
func WaitAndForceForeground(pid uint32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for {
		if h := VisibleWindowOf(pid); h != 0 && ForceForeground(h) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		time.Sleep(30 * time.Millisecond)
	}
}
