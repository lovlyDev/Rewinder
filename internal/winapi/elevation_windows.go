//go:build windows

package winapi

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ErrCancelled: the user said no to the UAC prompt.
var ErrCancelled = errors.New("elevation cancelled")

// SelfElevated reports whether this process runs as administrator.
func SelfElevated() bool { return windows.GetCurrentProcessToken().IsElevated() }

// RunElevated starts exe with args through the "runas" verb, which shows the
// UAC prompt. It returns ErrCancelled when the user declines.
func RunElevated(exe, args string) error {
	verb, _ := windows.UTF16PtrFromString("runas")
	file, err := windows.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}
	var params *uint16
	if args != "" {
		params, _ = windows.UTF16PtrFromString(args)
	}
	r, _, e := procShellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(file)),
		uintptr(unsafe.Pointer(params)), 0, SwShowNormal)
	if r > 32 {
		return nil
	}
	if errors.Is(e, windows.ERROR_CANCELLED) {
		return ErrCancelled
	}
	if e != nil && !errors.Is(e, windows.ERROR_SUCCESS) {
		return e
	}
	return errors.New("ShellExecute runas failed")
}

// WaitForExit waits up to timeoutMs for a process to end; used by a restarted
// Rewinder so the single-instance lock is free before it starts.
func WaitForExit(pid uint32, timeoutMs uint32) {
	h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, pid)
	if err != nil {
		return // already gone
	}
	defer windows.CloseHandle(h)
	_, _ = windows.WaitForSingleObject(h, timeoutMs)
}
