package winapi

import (
	"os"
	"unsafe"
)

// CursorMonitorWork returns physical desktop coordinates, including negative
// origins on secondary monitors.
func CursorMonitorWork() (Rect, bool) {
	var p Point
	if ok, _, _ := user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&p))); ok == 0 {
		return Rect{}, false
	}
	h, _, _ := user32.NewProc("MonitorFromPoint").Call(uintptr(uint64(uint32(p.X))|uint64(uint32(p.Y))<<32), monitorNearest)
	var mi monitorInfoEx
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	ok, _, _ := procGetMonitorInfoW.Call(h, uintptr(unsafe.Pointer(&mi)))
	return mi.Work, ok != 0
}

// OwnWindow finds the Wails window even when it is hidden in the tray.
// systray owns another window in this process, so select Wails' native class.
func OwnWindow() uintptr {
	for _, h := range TopLevelWindows() {
		if WindowPID(h) == uint32(os.Getpid()) && ClassName(h) == "wailsWindow" {
			return h
		}
	}
	return 0
}

func MonitorWork(hwnd uintptr) Rect {
	h, _, _ := procMonitorFromWindow.Call(hwnd, monitorNearest)
	var mi monitorInfoEx
	mi.CbSize = uint32(unsafe.Sizeof(mi))
	procGetMonitorInfoW.Call(h, uintptr(unsafe.Pointer(&mi)))
	return mi.Work
}

func SystemLanguage() string {
	id, _, _ := kernel32.NewProc("GetUserDefaultUILanguage").Call()
	if id&0x3ff == 0x19 {
		return "ru"
	}
	return "en"
}

func OwnWindowVisible() bool { return isVisible(OwnWindow()) }
