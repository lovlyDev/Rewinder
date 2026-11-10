//go:build windows

package winapi

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procMonitorFromWindow = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW   = user32.NewProc("GetMonitorInfoW")
)

const monitorNearest = 2

type monitorInfoEx struct {
	CbSize  uint32
	Monitor Rect
	Work    Rect
	Flags   uint32
	Device  [32]uint16
}

// MonitorName returns the device name of the monitor the window is on.
func MonitorName(hwnd uintptr) string {
	hmon, _, _ := procMonitorFromWindow.Call(hwnd, monitorNearest)
	if hmon == 0 {
		return ""
	}
	var mi monitorInfoEx
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	r, _, _ := procGetMonitorInfoW.Call(hmon, uintptr(unsafe.Pointer(&mi)))
	if r == 0 {
		return ""
	}
	return windows.UTF16ToString(mi.Device[:])
}
