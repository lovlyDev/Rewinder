//go:build windows

// Package tray shows the notification-area icon and its menu.
package tray

import (
	_ "embed"
	"fmt"
	"runtime"
	"sync"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
)

// White symbol for a dark taskbar, the full-colour tile for a light one.
var (
	//go:embed icon-white.ico
	iconWhite []byte
	//go:embed icon-color.ico
	iconColor []byte
)

// icon picks the tray icon that stays visible on the current taskbar theme.
func icon() []byte {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		return iconWhite
	}
	defer k.Close()
	if light, _, err := k.GetIntegerValue("SystemUsesLightTheme"); err == nil && light == 1 {
		return iconColor
	}
	return iconWhite
}

// Menu holds the actions behind the tray menu items. Nil actions are ignored.
type Menu struct {
	Hotkey           string
	Language         string
	mu               sync.Mutex
	items            []*systray.MenuItem
	OnOpenTimeline   func()
	OnPauseTracking  func()
	OnResumeTracking func()
	OnOpenSettings   func()
	OnExit           func()
	// OnLayout restores the layout at that position of the Layouts submenu.
	OnLayout func(int)

	layoutNames []string
	layoutsMenu *systray.MenuItem
	layoutItems []*systray.MenuItem

	once     sync.Once
	stopOnce sync.Once
	stop     chan struct{}
}

func (m *Menu) Start() {
	m.once.Do(func() {
		m.stop = make(chan struct{})
		go func() {
			// systray owns a window and a message loop; keep it on one OS thread.
			runtime.LockOSThread()
			systray.Run(m.onReady, func() {})
		}()
	})
}

func (m *Menu) Stop() {
	if m.stop == nil {
		return
	}
	m.stopOnce.Do(func() {
		close(m.stop)
		systray.Quit()
	})
}

func (m *Menu) onReady() {
	systray.SetIcon(icon())
	systray.SetTitle("Rewinder")

	open := systray.AddMenuItem("Open Timeline", "Open timeline window")
	pause := systray.AddMenuItem("Pause Tracking", "Pause tracking")
	resume := systray.AddMenuItem("Resume Tracking", "Resume tracking")
	settings := systray.AddMenuItem("Settings", "Open settings")
	layouts := systray.AddMenuItem("Layouts", "")
	// systray cannot add items later from another thread safely; create the nine
	// shortcut slots now and show or hide them as layouts come and go.
	slots := make([]*systray.MenuItem, MaxLayoutItems)
	for i := range slots {
		slots[i] = layouts.AddSubMenuItem("", "")
		slots[i].Hide()
		go func(i int, item *systray.MenuItem) {
			for {
				select {
				case <-m.stop:
					return
				case <-item.ClickedCh:
					if m.OnLayout != nil {
						m.OnLayout(i)
					}
				}
			}
		}(i, slots[i])
	}
	systray.AddSeparator()
	exit := systray.AddMenuItem("Exit", "Exit")

	m.mu.Lock()
	m.items = []*systray.MenuItem{open, pause, resume, settings, exit}
	m.layoutsMenu = layouts
	m.layoutItems = slots
	m.updateLocked()
	m.mu.Unlock()
	go func() {
		for {
			select {
			case <-m.stop:
				return
			case <-open.ClickedCh:
				call(m.OnOpenTimeline)
			case <-pause.ClickedCh:
				call(m.OnPauseTracking)
			case <-resume.ClickedCh:
				call(m.OnResumeTracking)
			case <-settings.ClickedCh:
				call(m.OnOpenSettings)
			case <-exit.ClickedCh:
				call(m.OnExit)
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

func (m *Menu) Update(hotkey, language string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Hotkey = hotkey
	m.Language = language
	if len(m.items) > 0 {
		m.updateLocked()
	}
}
func (m *Menu) updateLocked() {
	systray.SetTooltip("Rewinder: " + m.Hotkey)
	labels := []string{"Open Rewinder", "Pause tracking", "Resume tracking", "Settings", "Exit"}
	if m.Language == "ru" {
		labels = []string{"Открыть Rewinder", "Приостановить отслеживание", "Продолжить отслеживание", "Настройки", "Выход"}
	}
	for i, item := range m.items {
		item.SetTitle(labels[i])
		item.SetTooltip(labels[i])
	}
	title := "Layouts"
	if m.Language == "ru" {
		title = "Раскладки"
	}
	m.layoutsMenu.SetTitle(title)
	if len(m.layoutNames) == 0 {
		m.layoutsMenu.Disable()
	} else {
		m.layoutsMenu.Enable()
	}
	for i, item := range m.layoutItems {
		if i < len(m.layoutNames) {
			item.SetTitle(fmt.Sprintf("%s	Ctrl+Alt+%d", m.layoutNames[i], i+1))
			item.Show()
		} else {
			item.Hide()
		}
	}
}

// MaxLayoutItems is how many layouts the tray (and Ctrl+Alt+1…9) can reach.
const MaxLayoutItems = 9

// SetLayouts shows the names of the first nine layouts in the tray submenu.
func (m *Menu) SetLayouts(names []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.layoutNames = append([]string(nil), names...)
	if len(m.layoutNames) > MaxLayoutItems {
		m.layoutNames = m.layoutNames[:MaxLayoutItems]
	}
	if len(m.items) > 0 {
		m.updateLocked()
	}
}
