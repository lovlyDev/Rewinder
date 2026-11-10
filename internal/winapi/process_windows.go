//go:build windows

package winapi

import (
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetApplicationUserModelId  = kernel32.NewProc("GetApplicationUserModelId")
	procNtQueryInformationProcess  = ntdll.NewProc("NtQueryInformationProcess")
)

const (
	processCmdLineCls = 60 // ProcessCommandLineInformation
	insufficientBuf   = 122
)

func openLimited(pid uint32) (windows.Handle, error) {
	return windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
}

// ExePath returns the full image path of a process, or "" if it cannot be opened.
func ExePath(pid uint32) string {
	if pid == 0 {
		return ""
	}
	h, err := openLimited(pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	// Almost every path fits in MAX_PATH; only retry with a long-path buffer when it does not.
	for _, n := range [...]int{windows.MAX_PATH, windows.MAX_LONG_PATH} {
		buf := make([]uint16, n)
		size := uint32(n)
		r, _, _ := procQueryFullProcessImageNameW.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
		if r != 0 {
			return windows.UTF16ToString(buf[:size])
		}
	}
	return ""
}

type unicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

// CommandLine reads a process's command line. It needs only
// PROCESS_QUERY_LIMITED_INFORMATION, so no admin rights and no WMI.
func CommandLine(pid uint32) string {
	h, err := openLimited(pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	size := uint32(4096)
	for i := 0; i < 4; i++ {
		buf := make([]byte, size)
		var ret uint32
		st, _, _ := procNtQueryInformationProcess.Call(uintptr(h), processCmdLineCls,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(size), uintptr(unsafe.Pointer(&ret)))
		if st == 0 {
			us := (*unicodeString)(unsafe.Pointer(&buf[0]))
			if us.Buffer == nil || us.Length == 0 {
				return ""
			}
			return windows.UTF16ToString(unsafe.Slice(us.Buffer, us.Length/2))
		}
		if ret <= size {
			return ""
		}
		size = ret
	}
	return ""
}

// AppUserModelID returns the AUMID of a packaged (Store/MSIX) app, or "".
// Packaged apps cannot be started through CreateProcess on their WindowsApps exe.
func AppUserModelID(pid uint32) string {
	h, err := openLimited(pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	n := uint32(256)
	for i := 0; i < 2; i++ {
		buf := make([]uint16, n)
		r, _, _ := procGetApplicationUserModelId.Call(uintptr(h), uintptr(unsafe.Pointer(&n)), uintptr(unsafe.Pointer(&buf[0])))
		switch r {
		case 0:
			return windows.UTF16ToString(buf)
		case insufficientBuf:
			continue
		default:
			return ""
		}
	}
	return ""
}

// ProcessAlive reports whether pid is running the given executable (guards against PID reuse).
func ProcessAlive(pid uint32, exePath string) bool {
	p := ExePath(pid)
	return p != "" && (exePath == "" || strings.EqualFold(p, exePath))
}

// IsElevated reports whether the process runs as administrator. Windows does not
// let a normal process move the windows of an elevated one.
func IsElevated(pid uint32) bool {
	h, err := openLimited(pid)
	if err != nil {
		return false
	}
	defer windows.CloseHandle(h)
	var tok windows.Token
	if err := windows.OpenProcessToken(h, windows.TOKEN_QUERY, &tok); err != nil {
		return false
	}
	defer tok.Close()
	return tok.IsElevated()
}
