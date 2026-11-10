package restore

import (
	"testing"

	"Rewinder/internal/state"
)

func w(hwnd uintptr, class, title string, z int) state.WindowState {
	return state.WindowState{HWND: hwnd, ClassName: class, Title: title, ZOrder: z}
}

func TestMatchSameHWND(t *testing.T) {
	saved := []state.WindowState{w(10, "A", "x", 0), w(11, "A", "y", 1)}
	live := []state.WindowState{w(11, "A", "y", 0), w(10, "A", "x", 1)}
	m := MatchWindows(saved, live)
	if m[0] != 1 || m[1] != 0 {
		t.Fatalf("got %v", m)
	}
}

func TestMatchAfterRelaunchByTitleThenClass(t *testing.T) {
	// relaunched: all HWNDs are new
	saved := []state.WindowState{
		w(1, "Chrome_WidgetWin_1", "Docs - Google Chrome", 0),
		w(2, "Chrome_WidgetWin_1", "Mail - Google Chrome", 1),
	}
	live := []state.WindowState{
		w(100, "Chrome_WidgetWin_1", "New Tab - Google Chrome", 0),
		w(101, "Chrome_WidgetWin_1", "Mail - Google Chrome", 1),
	}
	m := MatchWindows(saved, live)
	if m[1] != 1 {
		t.Fatalf("title match must win: %v", m)
	}
	if m[0] != 0 {
		t.Fatalf("remaining window must fall back to class match: %v", m)
	}
}

func TestMatchByPath(t *testing.T) {
	a := w(1, "CabinetWClass", "Docs", 0)
	a.Path = `C:\Users\me\Docs`
	b := w(2, "CabinetWClass", "Docs", 1)
	b.Path = `D:\Backup\Docs`
	la := w(50, "CabinetWClass", "Docs", 0)
	la.Path = `d:\backup\docs`
	lb := w(51, "CabinetWClass", "Docs", 1)
	lb.Path = `c:\users\me\docs`
	m := MatchWindows([]state.WindowState{a, b}, []state.WindowState{la, lb})
	if m[0] != 1 || m[1] != 0 {
		t.Fatalf("folders with the same name must be told apart by path: %v", m)
	}
}

func TestMatchNeverReusesOrInvents(t *testing.T) {
	saved := []state.WindowState{w(1, "A", "x", 0), w(2, "A", "y", 1), w(3, "B", "z", 2)}
	live := []state.WindowState{w(9, "A", "q", 0)}
	m := MatchWindows(saved, live)
	if len(m) != 1 || m[0] != 0 {
		t.Fatalf("one live window must match exactly one saved window (topmost): %v", m)
	}
	if got := MatchWindows(saved, []state.WindowState{w(9, "C", "x", 0)}); len(got) != 0 {
		t.Fatalf("a window of another class must not be matched: %v", got)
	}
}
