//go:build windows && integration

// Integration tests drive real windows on the current desktop.
// Run with: go test -tags integration ./internal/restore/
package restore

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"Rewinder/internal/plugins"
	"Rewinder/internal/snapshot"
	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

var helperExe string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "rewinder-it-")
	if err != nil {
		panic(err)
	}
	helperExe = filepath.Join(dir, "winhelper.exe")
	build := exec.Command("go", "build", "-o", helperExe, "./testdata/winhelper")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		panic("building winhelper: " + err.Error())
	}
	code := m.Run()
	killHelpers()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

// killHelpers closes every helper process, including relaunched ones.
func killHelpers() {
	for _, h := range winapi.AppWindowsByExe(helperExe) {
		if p, err := os.FindProcess(int(winapi.WindowPID(h))); err == nil {
			_ = p.Kill()
		}
	}
	waitFor(func() bool { return len(winapi.AppWindowsByExe(helperExe)) == 0 }, 5*time.Second)
}

func waitFor(cond func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return cond()
}

func startHelper(t *testing.T, title string, args ...string) (*exec.Cmd, uintptr) {
	t.Helper()
	cmd := exec.Command(helperExe, append([]string{title}, args...)...)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	var hwnd uintptr
	ok := waitFor(func() bool {
		for _, h := range winapi.AppWindowsOf(map[uint32]bool{uint32(cmd.Process.Pid): true}) {
			hwnd = h
			return true
		}
		return false
	}, 5*time.Second)
	if !ok {
		t.Fatal("helper window did not appear")
	}
	return cmd, hwnd
}

func newEngine() *Engine {
	reg := plugins.DefaultRegistry()
	cap := state.NewCaptureEngine()
	cap.NeedsOpenFiles = reg.NeedsOpenFiles
	e := NewEngine(cap, reg)
	e.SettleDelay = 200 * time.Millisecond
	return e
}

func noProgress(string, int, string) {}

func placementOf(t *testing.T, hwnd uintptr) winapi.Placement {
	t.Helper()
	p, ok := winapi.GetPlacement(hwnd)
	if !ok {
		t.Fatal("GetWindowPlacement failed")
	}
	return p
}

func moveTo(t *testing.T, hwnd uintptr, r winapi.Rect, show uint32) {
	t.Helper()
	p := placementOf(t, hwnd)
	p.NormalRect, p.ShowCmd = r, show
	if err := winapi.SetPlacement(hwnd, p); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
}

func capture(t *testing.T, pid int) *snapshot.FullSnapshot {
	t.Helper()
	st, err := state.NewCaptureEngine().CaptureProcess(pid)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Windows) != 1 {
		t.Fatalf("captured %d windows, want 1: %+v", len(st.Windows), st.Windows)
	}
	return &snapshot.FullSnapshot{App: *st}
}

func TestCaptureSeesOnlyAppWindows(t *testing.T) {
	defer killHelpers()
	cmd, _ := startHelper(t, "Capture Test")
	full := capture(t, cmd.Process.Pid)
	w := full.App.Windows[0]
	if w.Title != "Capture Test" || w.ClassName != "RewinderTestWnd" {
		t.Fatalf("captured %+v", w)
	}
	if full.App.CommandLine == "" || full.App.ExecutablePath == "" {
		t.Fatalf("command line or exe missing: %+v", full.App)
	}
}

func TestRestoreMovedWindow(t *testing.T) {
	defer killHelpers()
	cmd, hwnd := startHelper(t, "Move Test")
	saved := winapi.Rect{Left: 150, Top: 120, Right: 650, Bottom: 520}
	moveTo(t, hwnd, saved, winapi.SwShowNormal)
	full := capture(t, cmd.Process.Pid)

	moveTo(t, hwnd, winapi.Rect{Left: 400, Top: 300, Right: 700, Bottom: 500}, winapi.SwShowNormal)

	res, err := newEngine().RestoreSnapshot(noProgress, nil, full)
	if err != nil {
		t.Fatal(err)
	}
	if res.Relaunched || res.Matched != 1 {
		t.Fatalf("result %+v", res)
	}
	if p := placementOf(t, hwnd); p.NormalRect != saved || p.ShowCmd != winapi.SwShowNormal {
		t.Fatalf("placement = %+v, want rect %+v normal", p, saved)
	}
}

func TestRestoreMaximizedAndMinimized(t *testing.T) {
	defer killHelpers()
	cmd, hwnd := startHelper(t, "State Test")
	normal := winapi.Rect{Left: 200, Top: 150, Right: 700, Bottom: 550}

	moveTo(t, hwnd, normal, winapi.SwShowMaximized)
	maxSnap := capture(t, cmd.Process.Pid)
	moveTo(t, hwnd, normal, winapi.SwShowMinimized)
	minSnap := capture(t, cmd.Process.Pid)
	if !minSnap.App.Windows[0].IsMinimized || !maxSnap.App.Windows[0].IsMaximized {
		t.Fatalf("states not captured: max=%+v min=%+v", maxSnap.App.Windows[0], minSnap.App.Windows[0])
	}

	// back to maximized from minimized: the old MoveWindow(GetWindowRect) code
	// would have put the window at -32000 or un-maximized it
	if _, err := newEngine().RestoreSnapshot(noProgress, nil, maxSnap); err != nil {
		t.Fatal(err)
	}
	if p := placementOf(t, hwnd); p.ShowCmd != winapi.SwShowMaximized || p.NormalRect != normal {
		t.Fatalf("after restoring maximized: %+v", p)
	}
	if _, err := newEngine().RestoreSnapshot(noProgress, nil, minSnap); err != nil {
		t.Fatal(err)
	}
	if p := placementOf(t, hwnd); p.ShowCmd != winapi.SwShowMinimized {
		t.Fatalf("after restoring minimized: %+v", p)
	}
}

func TestRestoreRelaunchesClosedApp(t *testing.T) {
	defer killHelpers()
	cmd, hwnd := startHelper(t, "Relaunch Test", "100", "100", "400", "300")
	saved := winapi.Rect{Left: 300, Top: 200, Right: 900, Bottom: 650}
	moveTo(t, hwnd, saved, winapi.SwShowNormal)
	full := capture(t, cmd.Process.Pid)

	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
	if !waitFor(func() bool { return len(winapi.AppWindowsByExe(helperExe)) == 0 }, 5*time.Second) {
		t.Fatal("helper did not exit")
	}

	res, err := newEngine().RestoreSnapshot(noProgress, nil, full)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Relaunched || res.Matched != 1 {
		t.Fatalf("result %+v", res)
	}
	wins := winapi.AppWindowsByExe(helperExe)
	if len(wins) != 1 {
		t.Fatalf("%d helper windows after relaunch", len(wins))
	}
	if winapi.Title(wins[0]) != "Relaunch Test" {
		t.Fatalf("relaunched with wrong arguments: title %q", winapi.Title(wins[0]))
	}
	if p := placementOf(t, wins[0]); p.NormalRect != saved {
		t.Fatalf("relaunched window at %+v, want %+v", p.NormalRect, saved)
	}
}

func TestRestoreFailsLoudlyWhenNothingMatches(t *testing.T) {
	defer killHelpers()
	cmd, _ := startHelper(t, "Mismatch Test")
	full := capture(t, cmd.Process.Pid)
	full.App.Windows[0].HWND = 1
	full.App.Windows[0].ClassName = "SomeOtherClass"
	if _, err := newEngine().RestoreSnapshot(noProgress, nil, full); err == nil {
		t.Fatal("restore must report an error instead of claiming success")
	}
}

// A window saved on a monitor that is no longer there must come back onto a
// monitor that exists, not to coordinates nobody can see.
func TestRestoreBringsBackWindowFromMissingMonitor(t *testing.T) {
	defer killHelpers()
	cmd, hwnd := startHelper(t, "Monitor Test")
	full := capture(t, cmd.Process.Pid)
	gone := state.Rect{Left: 30000, Top: 200, Right: 30600, Bottom: 600} // far right of any real desktop
	full.App.Windows[0].NormalRect = gone
	full.App.Windows[0].Rect = gone

	if _, err := newEngine().RestoreSnapshot(noProgress, nil, full); err != nil {
		t.Fatal(err)
	}
	works := winapi.WorkspaceWorkAreas()
	p := placementOf(t, hwnd)
	r := p.NormalRect
	if r.Right-r.Left != 600 || r.Bottom-r.Top != 400 {
		t.Fatalf("size changed: %+v", r)
	}
	if r.Left < works[0].Left || r.Right > works[0].Right || r.Top < works[0].Top || r.Bottom > works[0].Bottom {
		t.Fatalf("window at %+v is not on the primary monitor %+v", r, works[0])
	}
}
