//go:build windows

package plugins

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32           = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

func shellOpen(path string) error {
	verb, _ := windows.UTF16PtrFromString("open")
	p, _ := windows.UTF16PtrFromString(path)
	r1, _, e1 := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(p)),
		0,
		0,
		1,
	)
	if r1 <= 32 {
		return e1
	}
	return nil
}
