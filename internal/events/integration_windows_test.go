//go:build windows && integration

package events

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
	"unsafe"
)

var procSetWindowPos = user32.NewProc("SetWindowPos")
var procFindWindowW = user32.NewProc("FindWindowW")

// The old code installed the WinEvent hook on one goroutine and pumped messages
// on another, so out-of-context events were never delivered. This proves events
// from another process's window actually arrive.
func TestHookDeliversWindowEvents(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "winhelper.exe")
	build := exec.Command("go", "build", "-o", exe, "../restore/testdata/winhelper")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatal(err)
	}

	bus := NewBus(4096)
	bus.StartWindowsSources()
	defer bus.Stop()

	cmd := exec.Command(exe, "Hook Test")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()
	pid := cmd.Process.Pid

	var hwnd uintptr
	cls, _ := syscallUTF16("RewinderTestWnd")
	title, _ := syscallUTF16("Hook Test")
	deadline := time.Now().Add(5 * time.Second)
	for hwnd == 0 && time.Now().Before(deadline) {
		hwnd, _, _ = procFindWindowW.Call(uintptr(unsafe.Pointer(cls)), uintptr(unsafe.Pointer(title)))
		time.Sleep(50 * time.Millisecond)
	}
	if hwnd == 0 {
		t.Fatal("helper window did not appear")
	}
	_, _, _ = procSetWindowPos.Call(hwnd, 0, 321, 234, 0, 0, 0x0001|0x0004|0x0010) // NOSIZE|NOZORDER|NOACTIVATE

	got := map[EventType]bool{}
	timeout := time.After(5 * time.Second)
	for !got[EventWindowMoved] || !got[EventWindowShown] {
		select {
		case ev := <-bus.Events():
			if ev.PID == pid {
				got[ev.Type] = true
			}
		case <-timeout:
			t.Fatalf("events received for helper: %v; want window_shown and window_moved", got)
		}
	}
}

func syscallUTF16(s string) (*uint16, error) {
	b := make([]uint16, 0, len(s)+1)
	for _, r := range s {
		b = append(b, uint16(r))
	}
	return &append(b, 0)[0], nil
}

func TestBusStopReturns(t *testing.T) {
	bus := NewBus(16)
	bus.StartWindowsSources()
	done := make(chan struct{})
	go func() { bus.Stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Stop hung: the hook thread did not quit")
	}
}
