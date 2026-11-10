//go:build windows

package winapi

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procGetWindowThreadProcessId = user32.NewProc("GetWindowThreadProcessId")
	procIsWindow                 = user32.NewProc("IsWindow")
	procIsWindowVisible          = user32.NewProc("IsWindowVisible")
	procGetWindow                = user32.NewProc("GetWindow")
	procGetWindowLongPtrW        = user32.NewProc("GetWindowLongPtrW")
	procGetAncestor              = user32.NewProc("GetAncestor")
	procGetClassNameW            = user32.NewProc("GetClassNameW")
	procGetWindowTextW           = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW     = user32.NewProc("GetWindowTextLengthW")
	procGetWindowRect            = user32.NewProc("GetWindowRect")
	procDwmGetWindowAttribute    = dwmapi.NewProc("DwmGetWindowAttribute")
)

const (
	gwOwner        = 4
	gwlExStyle     = ^uintptr(20 - 1) // -20
	gaRoot         = 2
	wsExToolWindow = 0x00000080
	wsExNoActivate = 0x08000000
	dwmwaCloaked   = 14
)

// Shell windows that belong to explorer.exe but must never be captured or moved.
var shellClasses = map[string]bool{
	"progman":                    true,
	"workerw":                    true,
	"shell_traywnd":              true,
	"shell_secondarytraywnd":     true,
	"windows.ui.core.corewindow": true,
}

func WindowPID(hwnd uintptr) uint32 {
	var pid uint32
	_, _, _ = procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	return pid
}

func IsWindow(hwnd uintptr) bool {
	r, _, _ := procIsWindow.Call(hwnd)
	return r != 0
}

func RootWindow(hwnd uintptr) uintptr {
	r, _, _ := procGetAncestor.Call(hwnd, gaRoot)
	return r
}

func ClassName(hwnd uintptr) string {
	var buf [256]uint16
	r, _, _ := procGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:r])
}

func Title(hwnd uintptr) string {
	n, _, _ := procGetWindowTextLengthW.Call(hwnd)
	if int(n) <= 0 {
		return ""
	}
	buf := make([]uint16, n+1)
	r, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:r])
}

func WindowRect(hwnd uintptr) Rect {
	var r Rect
	_, _, _ = procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	return r
}

func isVisible(hwnd uintptr) bool {
	r, _, _ := procIsWindowVisible.Call(hwnd)
	return r != 0
}

func owner(hwnd uintptr) uintptr {
	r, _, _ := procGetWindow.Call(hwnd, gwOwner)
	return r
}

func exStyle(hwnd uintptr) uintptr {
	r, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
	return r
}

func isCloaked(hwnd uintptr) bool {
	var cloaked uint32
	r, _, _ := procDwmGetWindowAttribute.Call(hwnd, dwmwaCloaked, uintptr(unsafe.Pointer(&cloaked)), unsafe.Sizeof(cloaked))
	return r == 0 && cloaked != 0
}

// IsAppWindow reports whether hwnd is a real application window: the kind that
// shows on the taskbar and that a user would expect to be put back in place.
// Checks run cheapest first.
func IsAppWindow(hwnd uintptr) bool {
	if !isVisible(hwnd) || owner(hwnd) != 0 {
		return false
	}
	ex := exStyle(hwnd)
	if ex&wsExToolWindow != 0 || ex&wsExNoActivate != 0 || isCloaked(hwnd) {
		return false
	}
	return !shellClasses[strings.ToLower(ClassName(hwnd))]
}
