//go:build windows

package trayhotkey

import (
	_ "embed"
	"errors"
	"runtime"
	"sync"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

//go:embed icon.ico
var trayIcon []byte

type Dependencies struct {
	OnOpenTimeline   func()
	OnPauseTracking  func()
	OnResumeTracking func()
	OnOpenSettings   func()
	OnExit           func()
	OnOverlayHotkey  func()
	OnHotkeyError    func(error)
}

type Manager struct {
	deps     Dependencies
	once     sync.Once
	stopOnce sync.Once
	stop     chan struct{}

	hotkeyThread uint32
	hotkeyReady  chan struct{}
}

func NewManager(deps Dependencies) *Manager {
	return &Manager{deps: deps, stop: make(chan struct{}), hotkeyReady: make(chan struct{})}
}

func (m *Manager) Start() {
	m.once.Do(func() {
		go func() {
			// systray owns a window and a message loop; keep it on one OS thread.
			runtime.LockOSThread()
			systray.Run(m.onReady, func() {})
		}()
		go m.hotkeyLoop()
	})
}

func (m *Manager) Stop() {
	m.stopOnce.Do(func() {
		close(m.stop)
		<-m.hotkeyReady
		_, _, _ = procPostThreadMessageW.Call(uintptr(m.hotkeyThread), wmQuit, 0, 0)
		systray.Quit()
	})
}

func (m *Manager) onReady() {
	systray.SetIcon(trayIcon)
	systray.SetTitle("Rewinder")
	systray.SetTooltip("Rewinder: Ctrl+Alt+Z to restore")

	itemOpen := systray.AddMenuItem("Open Timeline", "Open timeline window")
	itemPause := systray.AddMenuItem("Pause Tracking", "Pause tracking")
	itemResume := systray.AddMenuItem("Resume Tracking", "Resume tracking")
	itemSettings := systray.AddMenuItem("Settings", "Open settings")
	systray.AddSeparator()
	itemExit := systray.AddMenuItem("Exit", "Exit")

	go func() {
		for {
			select {
			case <-m.stop:
				return
			case <-itemOpen.ClickedCh:
				call(m.deps.OnOpenTimeline)
			case <-itemPause.ClickedCh:
				call(m.deps.OnPauseTracking)
			case <-itemResume.ClickedCh:
				call(m.deps.OnResumeTracking)
			case <-itemSettings.ClickedCh:
				call(m.deps.OnOpenSettings)
			case <-itemExit.ClickedCh:
				call(m.deps.OnExit)
				return
			}
		}
	}()
}

func call(f func()) {
	if f != nil {
		f()
	}
}

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

// ErrHotkeyTaken is reported when another program already owns Ctrl+Alt+Z.
var ErrHotkeyTaken = errors.New("Ctrl+Alt+Z is already used by another program")

// hotkeyLoop registers the hotkey and pumps messages on the same locked thread:
// RegisterHotKey(NULL, ...) posts WM_HOTKEY to the registering thread's queue.
func (m *Manager) hotkeyLoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	m.hotkeyThread = windows.GetCurrentThreadId()
	r, _, _ := procRegisterHotKey.Call(0, hotkeyID, modControl|modAlt|modNoRepeat, vkZ)
	close(m.hotkeyReady)
	if r == 0 {
		if m.deps.OnHotkeyError != nil {
			m.deps.OnHotkeyError(ErrHotkeyTaken)
		}
	} else {
		defer procUnregisterHotKey.Call(0, hotkeyID)
	}

	var mm msg
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&mm)), 0, 0, 0)
		if int32(r) <= 0 {
			return
		}
		if mm.Message == wmHotkey && mm.WParam == hotkeyID {
			call(m.deps.OnOverlayHotkey)
		}
	}
}
