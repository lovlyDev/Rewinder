//go:build windows

package winapi

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

var procShellExecuteW = shell32.NewProc("ShellExecuteW")

// ShellOpen opens a file, folder or shell: URI with its default handler.
func ShellOpen(target string) error {
	return shellExecute(target, "")
}

// LaunchPackaged starts a Store/MSIX app by its AUMID.
func LaunchPackaged(aumid string) error {
	return shellExecute("explorer.exe", `shell:AppsFolder\`+aumid)
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
