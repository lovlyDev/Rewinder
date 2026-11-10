//go:build windows

// Package winapi holds the thin Win32 wrappers shared by capture and restore.
package winapi

import (
	"errors"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")
	dwmapi   = windows.NewLazySystemDLL("dwmapi.dll")
	shell32  = windows.NewLazySystemDLL("shell32.dll")
	ntdll    = windows.NewLazySystemDLL("ntdll.dll")

	procEnumWindows              = user32.NewProc("EnumWindows")
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
	procGetWindowPlacement       = user32.NewProc("GetWindowPlacement")
	procSetWindowPlacement       = user32.NewProc("SetWindowPlacement")
	procSetWindowPos             = user32.NewProc("SetWindowPos")
	procGetForegroundWindow      = user32.NewProc("GetForegroundWindow")
	procSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	procMonitorFromWindow        = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfoW          = user32.NewProc("GetMonitorInfoW")

	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
	procGetApplicationUserModelId  = kernel32.NewProc("GetApplicationUserModelId")

	procDwmGetWindowAttribute = dwmapi.NewProc("DwmGetWindowAttribute")
	procShellExecuteW         = shell32.NewProc("ShellExecuteW")

	procNtQueryInformationProcess = ntdll.NewProc("NtQueryInformationProcess")
)

const (
	gwOwner           = 4
	gwlExStyle        = ^uintptr(20 - 1) // -20
	gaRoot            = 2
	wsExToolWindow    = 0x00000080
	wsExNoActivate    = 0x08000000
	dwmwaCloaked      = 14
	monitorNearest    = 2
	swpNoSize         = 0x0001
	swpNoMove         = 0x0002
	swpNoActivate     = 0x0010
	swpNoOwnerZOrder  = 0x0200
	hwndTop           = 0
	processCmdLineCls = 60 // ProcessCommandLineInformation
	appModelNoPackage = 15703
	insufficientBuf   = 122
)

// Shell windows that belong to explorer.exe but must never be captured or moved.
var shellClasses = map[string]bool{
	"progman":                    true,
	"workerw":                    true,
	"shell_traywnd":              true,
	"shell_secondarytraywnd":     true,
	"windows.ui.core.corewindow": true,
}

type Rect struct {
	Left, Top, Right, Bottom int32
}

type Point struct {
	X, Y int32
}

// Placement mirrors WINDOWPLACEMENT.
type Placement struct {
	Length     uint32
	Flags      uint32
	ShowCmd    uint32
	MinPos     Point
	MaxPos     Point
	NormalRect Rect
}

const (
	SwHide          = 0
	SwShowNormal    = 1
	SwShowMinimized = 2
	SwShowMaximized = 3
)

// EnumWindows uses one shared callback: windows.NewCallback has a hard limit on
// the number of callbacks per process, so creating one per call eventually panics.
var (
	enumMu  sync.Mutex
	enumOut []uintptr
	enumCB  = windows.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		enumOut = append(enumOut, hwnd)
		return 1
	})
)

// TopLevelWindows returns all top-level windows, topmost first (EnumWindows order is Z-order).
func TopLevelWindows() []uintptr {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumOut = enumOut[:0]
	_, _, _ = procEnumWindows.Call(enumCB, 0)
	out := make([]uintptr, len(enumOut))
	copy(out, enumOut)
	return out
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
	buf := make([]uint16, 256)
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
func IsAppWindow(hwnd uintptr) bool {
	if !isVisible(hwnd) || owner(hwnd) != 0 || isCloaked(hwnd) {
		return false
	}
	ex := exStyle(hwnd)
	if ex&wsExToolWindow != 0 || ex&wsExNoActivate != 0 {
		return false
	}
	return !shellClasses[strings.ToLower(ClassName(hwnd))]
}

// AppWindowsOf returns app windows of the given PIDs, topmost first.
func AppWindowsOf(pids map[uint32]bool) []uintptr {
	var out []uintptr
	for _, h := range TopLevelWindows() {
		if pids[WindowPID(h)] && IsAppWindow(h) {
			out = append(out, h)
		}
	}
	return out
}

// AppWindowsByExe returns app windows whose process runs the given executable, topmost first.
// Many apps (browsers, VS Code, Office, Explorer) hand a new launch over to an existing
// instance, so windows have to be found by executable rather than by the launched PID.
func AppWindowsByExe(exePath string) []uintptr {
	cache := map[uint32]bool{}
	var out []uintptr
	for _, h := range TopLevelWindows() {
		pid := WindowPID(h)
		match, ok := cache[pid]
		if !ok {
			match = strings.EqualFold(ExePath(pid), exePath)
			cache[pid] = match
		}
		if match && IsAppWindow(h) {
			out = append(out, h)
		}
	}
	return out
}

type monitorInfoEx struct {
	CbSize  uint32
	Monitor Rect
	Work    Rect
	Flags   uint32
	Device  [32]uint16
}

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
	buf := make([]uint16, windows.MAX_LONG_PATH)
	size := uint32(len(buf))
	r, _, _ := procQueryFullProcessImageNameW.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)))
	if r == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:size])
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

// ShellOpen opens a file, folder or shell: URI with its default handler.
func ShellOpen(target string) error {
	return shellExecute(target, "")
}

func shellExecute(file, params string) error {
	verb, _ := windows.UTF16PtrFromString("open")
	f, _ := windows.UTF16PtrFromString(file)
	var p *uint16
	if params != "" {
		p, _ = windows.UTF16PtrFromString(params)
	}
	r, _, e := procShellExecuteW.Call(0, uintptr(unsafe.Pointer(verb)), uintptr(unsafe.Pointer(f)),
		uintptr(unsafe.Pointer(p)), 0, SwShowNormal)
	if r <= 32 {
		if e != nil && !errors.Is(e, windows.ERROR_SUCCESS) {
			return e
		}
		return errors.New("ShellExecute failed")
	}
	return nil
}

// LaunchPackaged starts a Store/MSIX app by its AUMID.
func LaunchPackaged(aumid string) error {
	return shellExecute("explorer.exe", `shell:AppsFolder\`+aumid)
}

// Launch starts a classic desktop process.
func Launch(exePath, cmdLine, workingDir string) (uint32, error) {
	if exePath == "" {
		return 0, errors.New("missing executable path")
	}
	if cmdLine == "" {
		cmdLine = `"` + exePath + `"`
	}
	app, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return 0, err
	}
	cl, err := windows.UTF16PtrFromString(cmdLine)
	if err != nil {
		return 0, err
	}
	var wd *uint16
	if workingDir != "" {
		wd, _ = windows.UTF16PtrFromString(workingDir)
	}
	var si windows.StartupInfo
	var pi windows.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))
	if err := windows.CreateProcess(app, cl, nil, nil, false, 0, nil, wd, &si, &pi); err != nil {
		return 0, err
	}
	_ = windows.CloseHandle(pi.Thread)
	_ = windows.CloseHandle(pi.Process)
	return pi.ProcessId, nil
}
