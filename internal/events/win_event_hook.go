//go:build windows

package events

import (
	"runtime"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	eventSystemForeground     = 0x0003
	eventSystemMoveSizeEnd    = 0x000B
	eventSystemMinimizeStart  = 0x0016
	eventSystemMinimizeEnd    = 0x0017
	eventObjectShow           = 0x8002
	eventObjectHide           = 0x8003
	eventObjectLocationChange = 0x800B
	eventObjectCloaked        = 0x8017
	eventObjectUncloaked      = 0x8018 // windows become visible this way after their open animation

	objidWindow            = 0
	childidSelf            = 0
	wineventOutOfContext   = 0x0000
	wineventSkipOwnProcess = 0x0002
	wmQuit                 = 0x0012
	gaRoot                 = 2
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procSetWinEventHook    = user32.NewProc("SetWinEventHook")
	procUnhookWinEvent     = user32.NewProc("UnhookWinEvent")
	procGetWindowThreadPID = user32.NewProc("GetWindowThreadProcessId")
	procGetAncestor        = user32.NewProc("GetAncestor")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procPostThreadMessageW = user32.NewProc("PostThreadMessageW")
)

type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	PtX     int32
	PtY     int32
}

// winEventHook delivers top-level window events. Out-of-context WinEvent hooks
// are delivered through the message queue of the thread that installed them,
// so install and pump must happen on one locked OS thread.
type winEventHook struct {
	emit     func(SystemEvent)
	threadID uint32
	ready    chan struct{}
}

var hookCallback uintptr

func (w *winEventHook) callback(_ windows.Handle, event uint32, hwnd uintptr, idObject int32, idChild int32, _ uint32, _ uint32) uintptr {
	if idObject != objidWindow || idChild != childidSelf || hwnd == 0 {
		return 0
	}
	if root, _, _ := procGetAncestor.Call(hwnd, gaRoot); root != hwnd {
		return 0
	}
	var t EventType
	switch event {
	case eventSystemForeground:
		t = EventForegroundChanged
	case eventSystemMoveSizeEnd, eventObjectLocationChange, eventSystemMinimizeStart, eventSystemMinimizeEnd:
		t = EventWindowMoved
	case eventObjectShow, eventObjectUncloaked:
		t = EventWindowShown
	case eventObjectHide, eventObjectCloaked:
		t = EventWindowHidden
	default:
		return 0
	}
	var pid uint32
	_, _, _ = procGetWindowThreadPID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	w.emit(SystemEvent{Type: t, Timestamp: time.Now().UTC().UnixMilli(), PID: int(pid), HWND: hwnd})
	return 0
}

// run blocks until stop is called. It must be started in its own goroutine.
func (w *winEventHook) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	w.threadID = windows.GetCurrentThreadId()
	if hookCallback == 0 {
		hookCallback = windows.NewCallback(w.callback)
	}
	ranges := [][2]uint32{
		{eventSystemForeground, eventSystemForeground},
		{eventSystemMoveSizeEnd, eventSystemMoveSizeEnd},
		{eventSystemMinimizeStart, eventSystemMinimizeEnd},
		{eventObjectShow, eventObjectHide},
		{eventObjectLocationChange, eventObjectLocationChange},
		{eventObjectCloaked, eventObjectUncloaked},
	}
	var hooks []uintptr
	for _, r := range ranges {
		h, _, _ := procSetWinEventHook.Call(uintptr(r[0]), uintptr(r[1]), 0, hookCallback, 0, 0,
			wineventOutOfContext|wineventSkipOwnProcess)
		if h != 0 {
			hooks = append(hooks, h)
		}
	}
	close(w.ready)
	defer func() {
		for _, h := range hooks {
			_, _, _ = procUnhookWinEvent.Call(h)
		}
	}()

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 { // WM_QUIT or error
			return
		}
		_, _, _ = procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		_, _, _ = procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func (w *winEventHook) stop() {
	<-w.ready
	_, _, _ = procPostThreadMessageW.Call(uintptr(w.threadID), wmQuit, 0, 0)
}
