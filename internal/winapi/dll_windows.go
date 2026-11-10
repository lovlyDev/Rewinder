//go:build windows

// Package winapi holds the thin Win32 wrappers shared by capture and restore.
package winapi

import "golang.org/x/sys/windows"

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	dwmapi   = windows.NewLazySystemDLL("dwmapi.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	ntdll    = windows.NewLazySystemDLL("ntdll.dll")
)

type Rect struct {
	Left, Top, Right, Bottom int32
}

type Point struct {
	X, Y int32
}

// ShowWindow commands used in WINDOWPLACEMENT.ShowCmd.
const (
	SwHide          = 0
	SwShowNormal    = 1
	SwShowMinimized = 2
	SwShowMaximized = 3
)
