//go:build windows

// Package hotkey registers the global Ctrl+Alt+Z hotkey.
package hotkey

import (
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	modAlt      = 0x0001
	modControl  = 0x0002
	modNoRepeat = 0x4000
	vkZ         = 0x5A
	wmHotkey    = 0x0312
	wmQuit      = 0x0012
	hotkeyID    = 0xA11
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procGetMessageW        = user32.NewProc("GetMessageW")
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

// ErrTaken is reported when another program already owns Ctrl+Alt+Z.
var ErrTaken = errors.New("Ctrl+Alt+Z is already used by another program")

// Listener calls OnPress each time the hotkey is pressed.
type Listener struct {
	OnPress func()
	OnError func(error)

	once     sync.Once
	stopOnce sync.Once
	thread   uint32
	ready    chan struct{}
}

func (l *Listener) Start() {
	l.once.Do(func() {
		l.ready = make(chan struct{})
		go l.run()
	})
}

func (l *Listener) Stop() {
	if l.ready == nil {
		return
	}
	l.stopOnce.Do(func() {
		<-l.ready
		_, _, _ = procPostThreadMessageW.Call(uintptr(l.thread), wmQuit, 0, 0)
	})
}

// run registers the hotkey and pumps messages on the same locked thread:
// RegisterHotKey(NULL, ...) posts WM_HOTKEY to the registering thread's queue.
func (l *Listener) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	l.thread = windows.GetCurrentThreadId()
	r, _, _ := procRegisterHotKey.Call(0, hotkeyID, modControl|modAlt|modNoRepeat, vkZ)
	close(l.ready)
	if r == 0 {
		if l.OnError != nil {
			l.OnError(ErrTaken)
		}
	} else {
		defer procUnregisterHotKey.Call(0, hotkeyID)
	}

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		if m.Message == wmHotkey && m.WParam == hotkeyID && l.OnPress != nil {
			l.OnPress()
		}
	}
}
