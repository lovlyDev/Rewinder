//go:build windows

package winapi

import "unsafe"

var (
	procGetWindowPlacement  = user32.NewProc("GetWindowPlacement")
	procSetWindowPlacement  = user32.NewProc("SetWindowPlacement")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

const (
	swpNoSize        = 0x0001
	swpNoMove        = 0x0002
	swpNoActivate    = 0x0010
	swpNoOwnerZOrder = 0x0200
	hwndTop          = 0
)

// Placement mirrors WINDOWPLACEMENT.
type Placement struct {
	Length     uint32
	Flags      uint32
	ShowCmd    uint32
	MinPos     Point
	MaxPos     Point
	NormalRect Rect
}

func GetPlacement(hwnd uintptr) (Placement, bool) {
	var p Placement
	p.Length = uint32(unsafe.Sizeof(p))
	r, _, _ := procGetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&p)))
	return p, r != 0
}

func SetPlacement(hwnd uintptr, p Placement) error {
	p.Length = uint32(unsafe.Sizeof(p))
	r, _, e := procSetWindowPlacement.Call(hwnd, uintptr(unsafe.Pointer(&p)))
	if r == 0 {
		return e
	}
	return nil
}

// BringToTop raises the window in Z-order without activating it.
func BringToTop(hwnd uintptr) {
	_, _, _ = procSetWindowPos.Call(hwnd, hwndTop, 0, 0, 0, 0, swpNoMove|swpNoSize|swpNoActivate|swpNoOwnerZOrder)
}

func ForegroundWindow() uintptr {
	r, _, _ := procGetForegroundWindow.Call()
	return r
}

func SetForeground(hwnd uintptr) bool {
	r, _, _ := procSetForegroundWindow.Call(hwnd)
	return r != 0
}
