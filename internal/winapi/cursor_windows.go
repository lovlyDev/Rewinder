package winapi

import "unsafe"

// CursorMonitorWork returns physical desktop coordinates, including negative
// origins on secondary monitors. Wails WindowSetPosition uses the same space.
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

func SystemLanguage() string {
	id, _, _ := kernel32.NewProc("GetUserDefaultUILanguage").Call()
	if id&0x3ff == 0x19 {
		return "ru"
	}
	return "en"
}
