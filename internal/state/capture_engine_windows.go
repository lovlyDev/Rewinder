//go:build windows

package state

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"Rewinder/internal/winapi"
)

// ErrNoWindows means the process has no application windows worth snapshotting.
var ErrNoWindows = errors.New("process has no app windows")

type CaptureEngine struct {
	// NeedsOpenFiles decides whether open file handles are collected for an executable.
	// Enumerating the system handle table is expensive, so only plugins that use it ask.
	NeedsOpenFiles func(exePath string) bool
}

func NewCaptureEngine() *CaptureEngine { return &CaptureEngine{} }

// CaptureForeground captures the app that owns the current foreground window.
func (c *CaptureEngine) CaptureForeground() (*AppState, error) {
	hwnd := winapi.ForegroundWindow()
	if hwnd == 0 {
		return nil, errors.New("no foreground window")
	}
	return c.CaptureProcess(int(winapi.WindowPID(hwnd)))
}

// CaptureProcess captures all app windows of one process.
func (c *CaptureEngine) CaptureProcess(pid int) (*AppState, error) {
	if pid <= 0 {
		return nil, errors.New("invalid pid")
	}
	upid := uint32(pid)
	exe := winapi.ExePath(upid)
	if exe == "" {
		return nil, errors.New("cannot open process")
	}

	wins := captureWindows(map[uint32]bool{upid: true})
	if len(wins) == 0 {
		return nil, ErrNoWindows
	}
	fgClass := ""
	for _, w := range wins {
		if w.IsForeground {
			fgClass = w.ClassName
		}
	}

	st := &AppState{
		AppID:                 StableAppID(exe),
		PID:                   pid,
		ExecutablePath:        exe,
		CommandLine:           winapi.CommandLine(upid),
		WorkingDir:            filepath.Dir(exe),
		AppUserModelID:        winapi.AppUserModelID(upid),
		ForegroundWindowClass: fgClass,
		Windows:               wins,
		PluginData:            map[string]PluginState{},
		Timestamp:             time.Now(),
	}
	if c.NeedsOpenFiles != nil && c.NeedsOpenFiles(exe) {
		for _, p := range winapi.OpenFiles(upid, 300*time.Millisecond) {
			st.OpenFiles = append(st.OpenFiles, FileRef{Path: p})
		}
	}
	return st, nil
}

// CaptureWindowsOf returns the current app windows of the given processes.
func CaptureWindowsOf(pids map[uint32]bool) []WindowState { return captureWindows(pids) }

func captureWindows(pids map[uint32]bool) []WindowState {
	fg := winapi.ForegroundWindow()
	var out []WindowState
	for i, h := range winapi.AppWindowsOf(pids) {
		out = append(out, DescribeWindow(h, i, h == fg))
	}
	return out
}

// DescribeWindow reads the current state of one window.
func DescribeWindow(h uintptr, z int, foreground bool) WindowState {
	r := winapi.WindowRect(h)
	ws := WindowState{
		HWND:         h,
		ClassName:    winapi.ClassName(h),
		Title:        winapi.Title(h),
		ZOrder:       z,
		MonitorID:    winapi.MonitorName(h),
		Rect:         Rect{Left: r.Left, Top: r.Top, Right: r.Right, Bottom: r.Bottom},
		IsForeground: foreground,
	}
	if p, ok := winapi.GetPlacement(h); ok {
		ws.NormalRect = Rect{Left: p.NormalRect.Left, Top: p.NormalRect.Top, Right: p.NormalRect.Right, Bottom: p.NormalRect.Bottom}
		ws.ShowCmd = p.ShowCmd
		ws.Flags = p.Flags
		ws.MinPos = Point{X: p.MinPos.X, Y: p.MinPos.Y}
		ws.MaxPos = Point{X: p.MaxPos.X, Y: p.MaxPos.Y}
		ws.IsMinimized = p.ShowCmd == winapi.SwShowMinimized
		ws.IsMaximized = p.ShowCmd == winapi.SwShowMaximized
	}
	return ws
}

// Placement converts the saved geometry back to a WINDOWPLACEMENT.
func (w WindowState) Placement() winapi.Placement {
	show := w.ShowCmd
	if show == winapi.SwHide {
		show = winapi.SwShowNormal
	}
	return winapi.Placement{
		Flags:      w.Flags,
		ShowCmd:    show,
		MinPos:     winapi.Point{X: w.MinPos.X, Y: w.MinPos.Y},
		MaxPos:     winapi.Point{X: w.MaxPos.X, Y: w.MaxPos.Y},
		NormalRect: winapi.Rect{Left: w.NormalRect.Left, Top: w.NormalRect.Top, Right: w.NormalRect.Right, Bottom: w.NormalRect.Bottom},
	}
}

func StableAppID(exePath string) string {
	if exePath == "" {
		return "unknown"
	}
	base := strings.ToLower(filepath.Base(exePath))
	sum := sha256.Sum256([]byte(strings.ToLower(exePath)))
	return fmt.Sprintf("%s:%s", base, hex.EncodeToString(sum[:8]))
}
