package layouts

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"Rewinder/internal/state"
)

func apps(n int) []state.AppState {
	out := make([]state.AppState, n)
	for i := range out {
		out[i] = state.AppState{AppID: "app", ExecutablePath: `C:\a.exe`, Windows: []state.WindowState{{HWND: uintptr(i + 1)}}}
	}
	return out
}

func TestAddRenameMoveDeletePersist(t *testing.T) {
	dir := t.TempDir()
	s := Load(dir)
	now := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	work, err := s.Add("  Работа  ", apps(2), now)
	if err != nil || work.Name != "Работа" {
		t.Fatalf("%+v %v", work, err)
	}
	stream, _ := s.Add("Stream", apps(1), now)
	if err := s.Rename(stream.ID, "Стрим"); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(stream.ID, 0); err != nil {
		t.Fatal(err)
	}

	again := Load(dir).List()
	if len(again) != 2 || again[0].Name != "Стрим" || again[1].Name != "Работа" || len(again[1].Apps) != 2 {
		t.Fatalf("reloaded %+v", again)
	}
	if err := s.Delete(work.ID); err != nil {
		t.Fatal(err)
	}
	if got := Load(dir).List(); len(got) != 1 || got[0].ID != stream.ID {
		t.Fatalf("after delete %+v", got)
	}
}

func TestSameNameReplaces(t *testing.T) {
	s := Load(t.TempDir())
	a, _ := s.Add("Work", apps(1), time.Now())
	b, _ := s.Add("work", apps(3), time.Now())
	if a.ID != b.ID || len(s.List()) != 1 || len(s.List()[0].Apps) != 3 {
		t.Fatalf("%+v", s.List())
	}
}

func TestRejectsEmptyAndKeepsCorruptFile(t *testing.T) {
	dir := t.TempDir()
	s := Load(dir)
	if _, err := s.Add("   ", apps(1), time.Now()); err == nil {
		t.Fatal("empty name accepted")
	}
	if _, err := s.Add("x", nil, time.Now()); err == nil {
		t.Fatal("layout without apps accepted")
	}
	if err := os.WriteFile(filepath.Join(dir, fileName), []byte("{nope"), 0o600); err != nil {
		t.Fatal(err)
	}
	if len(Load(dir).List()) != 0 {
		t.Fatal("corrupt file should load as empty")
	}
	if _, err := os.Stat(filepath.Join(dir, fileName+".bad")); err != nil {
		t.Fatal("corrupt file was not kept aside")
	}
}
