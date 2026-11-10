//go:build windows

package winapi

import (
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	procNtQuerySystemInformation  = ntdll.NewProc("NtQuerySystemInformation")
	procGetFinalPathNameByHandleW = kernel32.NewProc("GetFinalPathNameByHandleW")
	procGetFileType               = kernel32.NewProc("GetFileType")
)

const (
	systemExtendedHandleInformation = 0x40
	statusInfoLengthMismatch        = 0xC0000004
	fileTypeDisk                    = 0x0001
)

// systemHandleEntryEx is SYSTEM_HANDLE_TABLE_ENTRY_INFO_EX.
type systemHandleEntryEx struct {
	Object                uintptr
	UniqueProcessID       uintptr
	HandleValue           uintptr
	GrantedAccess         uint32
	CreatorBackTraceIndex uint16
	ObjectTypeIndex       uint16
	HandleAttributes      uint32
	Reserved              uint32
}

// systemHandleInfoEx is the SYSTEM_HANDLE_INFORMATION_EX header; entries follow it.
type systemHandleInfoEx struct {
	NumberOfHandles uintptr
	Reserved        uintptr
}

func processHandles(pid uint32) ([]uintptr, error) {
	size := uint32(4 << 20)
	for i := 0; i < 8; i++ {
		buf := make([]byte, size)
		var ret uint32
		st, _, _ := procNtQuerySystemInformation.Call(systemExtendedHandleInformation,
			uintptr(unsafe.Pointer(&buf[0])), uintptr(size), uintptr(unsafe.Pointer(&ret)))
		if uint32(st) == statusInfoLengthMismatch {
			if ret > size {
				size = ret + 1<<20
			} else {
				size *= 2
			}
			continue
		}
		if st != 0 {
			return nil, windows.NTStatus(st)
		}
		hdr := (*systemHandleInfoEx)(unsafe.Pointer(&buf[0]))
		n := int(hdr.NumberOfHandles)
		maxN := (len(buf) - int(unsafe.Sizeof(*hdr))) / int(unsafe.Sizeof(systemHandleEntryEx{}))
		if n > maxN {
			n = maxN
		}
		entries := unsafe.Slice((*systemHandleEntryEx)(unsafe.Add(unsafe.Pointer(&buf[0]), unsafe.Sizeof(*hdr))), n)
		var out []uintptr
		for _, e := range entries {
			if uint32(e.UniqueProcessID) == pid {
				out = append(out, e.HandleValue)
			}
		}
		return out, nil
	}
	return nil, windows.ERROR_INSUFFICIENT_BUFFER
}

// OpenFiles lists the disk files a process has open. It is best effort and
// bounded in time. Only FILE_TYPE_DISK handles are resolved, so it cannot hang
// on synchronous pipes the way NtQueryObject(ObjectNameInformation) does.
func OpenFiles(pid uint32, budget time.Duration) []string {
	handles, err := processHandles(pid)
	if err != nil || len(handles) == 0 {
		return nil
	}
	hProc, err := windows.OpenProcess(windows.PROCESS_DUP_HANDLE, false, pid)
	if err != nil {
		return nil
	}
	defer windows.CloseHandle(hProc)

	deadline := time.Now().Add(budget)
	seen := map[string]bool{}
	var out []string
	buf := make([]uint16, windows.MAX_LONG_PATH)
	for _, hv := range handles {
		if time.Now().After(deadline) {
			break
		}
		var dup windows.Handle
		if err := windows.DuplicateHandle(hProc, windows.Handle(hv), windows.CurrentProcess(), &dup, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
			continue
		}
		ft, _, _ := procGetFileType.Call(uintptr(dup))
		if ft != fileTypeDisk {
			_ = windows.CloseHandle(dup)
			continue
		}
		n, _, _ := procGetFinalPathNameByHandleW.Call(uintptr(dup), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), 0)
		_ = windows.CloseHandle(dup)
		if n == 0 || int(n) >= len(buf) {
			continue
		}
		p := NormalizeFinalPath(windows.UTF16ToString(buf[:n]))
		if p == "" {
			continue
		}
		key := strings.ToLower(p)
		if !seen[key] {
			seen[key] = true
			out = append(out, p)
		}
	}
	return out
}

// NormalizeFinalPath turns a GetFinalPathNameByHandle result into a normal DOS path.
func NormalizeFinalPath(p string) string {
	switch {
	case strings.HasPrefix(p, `\\?\UNC\`):
		return `\\` + p[len(`\\?\UNC\`):]
	case strings.HasPrefix(p, `\\?\`):
		rest := p[len(`\\?\`):]
		if len(rest) >= 2 && rest[1] == ':' {
			return rest
		}
		return "" // volume GUID path without a drive letter
	default:
		return p
	}
}
