//go:build windows

// Package hotkey registers the global Ctrl+Alt+Z hotkey.
package hotkey

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	modNoRepeat = 0x4000
	wmHotkey    = 0x0312
	wmQuit      = 0x0012
	wmRebind    = 0x8001
	hotkeyID    = 0xA11
)

var (
	user32                 = windows.NewLazySystemDLL("user32.dll")
	procRegisterHotKey     = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey   = user32.NewProc("UnregisterHotKey")
	procPeekMessageW       = user32.NewProc("PeekMessageW")
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

// ErrTaken is reported when another program already owns the combination.
var ErrTaken = errors.New("hotkey.taken")

// Listener calls OnPress each time the hotkey is pressed.
type Listener struct {
	Initial Combination
	OnPress func()
	OnError func(error)

	once     sync.Once
	stopOnce sync.Once
	thread   uint32
	ready    chan struct{}
	done     chan struct{}
	requests chan rebindRequest
	mu       sync.Mutex
}

func (l *Listener) Start() {
	l.once.Do(func() {
		l.ready = make(chan struct{})
		l.done = make(chan struct{})
		l.requests = make(chan rebindRequest, 1)
		go l.run()
	})
}

func (l *Listener) Stop() {
	if l.ready == nil {
		return
	}
	l.stopOnce.Do(func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		<-l.ready
		_, _, _ = procPostThreadMessageW.Call(uintptr(l.thread), wmQuit, 0, 0)
		<-l.done
	})
}

// run registers the hotkey and pumps messages on the same locked thread:
// RegisterHotKey(NULL, ...) posts WM_HOTKEY to the registering thread's queue.
func (l *Listener) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer close(l.done)
	l.thread = windows.GetCurrentThreadId()
	var initialMessage msg
	procPeekMessageW.Call(uintptr(unsafe.Pointer(&initialMessage)), 0, 0, 0, 0)
	current := l.Initial
	if current.Key == 0 {
		current = Default()
	}
	r, _, _ := procRegisterHotKey.Call(0, hotkeyID, uintptr(current.Mods)|modNoRepeat, uintptr(current.Key))
	close(l.ready)
	if r == 0 {
		if l.OnError != nil {
			l.OnError(ErrTaken)
		}
	}
	defer procUnregisterHotKey.Call(0, hotkeyID)

	var m msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		if m.Message == wmRebind {
			req := <-l.requests
			procUnregisterHotKey.Call(0, hotkeyID)
			ok, _, _ := procRegisterHotKey.Call(0, hotkeyID, uintptr(req.combo.Mods)|modNoRepeat, uintptr(req.combo.Key))
			var err error
			if ok == 0 {
				err = ErrTaken
				if restored, _, _ := procRegisterHotKey.Call(0, hotkeyID, uintptr(current.Mods)|modNoRepeat, uintptr(current.Key)); restored == 0 {
					err = fmt.Errorf("hotkey.rollback: %w", err)
				}
			} else {
				current = req.combo
			}
			req.result <- err
		}
		if m.Message == wmHotkey && m.WParam == hotkeyID && l.OnPress != nil {
			go l.OnPress()
		}
	}
}

type rebindRequest struct {
	combo  Combination
	result chan error
}

// Rebind performs registration and rollback on the message-pump thread.
func (l *Listener) Rebind(c Combination) error {
	if err := c.Validate(); err != nil {
		return err
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.ready == nil {
		return errors.New("hotkey.unavailable")
	}
	<-l.ready
	select {
	case <-l.done:
		return errors.New("hotkey.unavailable")
	default:
	}
	req := rebindRequest{c, make(chan error, 1)}
	l.requests <- req
	if ok, _, err := procPostThreadMessageW.Call(uintptr(l.thread), wmRebind, 0, 0); ok == 0 {
		<-l.requests
		return fmt.Errorf("hotkey.unavailable: %w", err)
	}
	select {
	case err := <-req.result:
		return err
	case <-l.done:
		return errors.New("hotkey.unavailable")
	}
}
