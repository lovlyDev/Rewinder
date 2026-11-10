package snapshot

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"Rewinder/internal/state"
)

func win(hwnd uintptr, l, t, r, b int32) state.WindowState {
	return state.WindowState{
		HWND: hwnd, ClassName: "Cls", Title: "T",
		Rect:       state.Rect{Left: l, Top: t, Right: r, Bottom: b},
		NormalRect: state.Rect{Left: l, Top: t, Right: r, Bottom: b},
		ShowCmd:    1,
	}
}

func app(ts time.Time, wins ...state.WindowState) *state.AppState {
	return &state.AppState{
		AppID: "notepad.exe:0011", PID: 42, ExecutablePath: `C:\Windows\notepad.exe`,
		CommandLine: `notepad.exe C:\a.txt`, Windows: wins, Timestamp: ts,
		PluginData: map[string]state.PluginState{"x": {Paths: []string{`C:\a`}}},
	}
}

func newTestEngine(t *testing.T, dir string, now time.Time) *Engine {
	t.Helper()
	e := NewEngine(EngineConfig{MaxSnapshotsPerApp: 5, Retention: time.Hour, StorageDir: dir})
	e.now = func() time.Time { return now }
	return e
}

func TestIngestResolveRoundTrip(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	e := newTestEngine(t, "", now)

	var states []*state.AppState
	for i := 0; i < 4; i++ {
		s := app(now.Add(time.Duration(i)*time.Second), win(1, int32(i*10), 0, 500, 400))
		states = append(states, s)
		m, err := e.Ingest(s)
		if err != nil || m == nil {
			t.Fatalf("ingest %d: meta=%v err=%v", i, m, err)
		}
	}
	tl := e.GetTimeline("notepad.exe:0011")
	if len(tl) != 4 {
		t.Fatalf("timeline len = %d, want 4", len(tl))
	}
	// newest first; every snapshot resolves to exactly the state that was ingested,
	// including fields a delta chain used to lose (PID, command line, plugin data)
	for i, m := range tl {
		want := states[len(states)-1-i]
		_, full, err := e.ResolveSnapshot(m.AppID, m.SnapshotID)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(full.App, *want) {
			t.Errorf("snapshot %d resolved to %+v, want %+v", i, full.App, *want)
		}
	}
}

func TestIngestSkipsUnchangedState(t *testing.T) {
	now := time.Now()
	e := newTestEngine(t, "", now)
	if m, _ := e.Ingest(app(now, win(1, 0, 0, 100, 100))); m == nil {
		t.Fatal("first state must create a snapshot")
	}
	same := app(now.Add(time.Second), win(1, 0, 0, 100, 100))
	same.Windows[0].Title = "changed title"
	same.Windows[0].IsForeground = true
	same.Windows[0].ZOrder = 3
	if m, _ := e.Ingest(same); m != nil {
		t.Fatal("title/focus/z-order changes alone must not create a snapshot")
	}
	if m, _ := e.Ingest(app(now.Add(2*time.Second), win(1, 50, 0, 150, 100))); m == nil {
		t.Fatal("a moved window must create a snapshot")
	}
	min := app(now.Add(3*time.Second), win(1, 50, 0, 150, 100))
	min.Windows[0].ShowCmd = 2
	if m, _ := e.Ingest(min); m == nil {
		t.Fatal("minimizing must create a snapshot")
	}
}

func TestStoredSnapshotsDoNotAliasInput(t *testing.T) {
	now := time.Now()
	e := newTestEngine(t, "", now)
	s := app(now, win(1, 0, 0, 100, 100))
	e.Ingest(s)
	s.Windows[0].Rect.Left = 999
	s.PluginData["x"].Paths[0] = "mutated"

	m := e.GetTimeline(s.AppID)[0]
	_, full, _ := e.ResolveSnapshot(s.AppID, m.SnapshotID)
	if full.App.Windows[0].Rect.Left != 0 || full.App.PluginData["x"].Paths[0] != `C:\a` {
		t.Fatalf("stored snapshot was mutated through the caller's state: %+v", full.App)
	}
}

func TestTrimByCountAndRetention(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	e := newTestEngine(t, "", now)
	// two snapshots older than retention, then eight fresh ones
	for i := 0; i < 10; i++ {
		ts := now.Add(-30 * time.Minute).Add(time.Duration(i) * time.Second)
		if i < 2 {
			ts = now.Add(-2 * time.Hour)
		}
		e.Ingest(app(ts, win(1, int32(i), 0, 100, 100)))
	}
	tl := e.GetTimeline("notepad.exe:0011")
	if len(tl) != 5 {
		t.Fatalf("len = %d, want max 5", len(tl))
	}
	for _, m := range tl {
		if time.UnixMilli(m.Timestamp).Before(now.Add(-time.Hour)) {
			t.Fatalf("snapshot older than retention survived: %v", m)
		}
		if _, _, err := e.ResolveSnapshot(m.AppID, m.SnapshotID); err != nil {
			t.Fatalf("trimmed timeline left an unresolvable snapshot: %v", err)
		}
	}
}

func TestSaveLoadPersistsHistory(t *testing.T) {
	dir := t.TempDir()
	now := time.Now().Truncate(time.Millisecond)
	e := newTestEngine(t, dir, now)
	e.Ingest(app(now, win(1, 0, 0, 100, 100)))
	e.Ingest(app(now.Add(time.Second), win(1, 10, 10, 100, 100)))
	if err := e.Save(); err != nil {
		t.Fatal(err)
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.json.gz"))
	if len(files) != 1 {
		t.Fatalf("files = %v", files)
	}

	e2 := newTestEngine(t, dir, now)
	if err := e2.Load(); err != nil {
		t.Fatal(err)
	}
	a, b := e.GetTimeline("notepad.exe:0011"), e2.GetTimeline("notepad.exe:0011")
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("timeline after reload differs:\n%v\n%v", a, b)
	}
	_, full, err := e2.ResolveSnapshot(b[0].AppID, b[0].SnapshotID)
	if err != nil || full.App.PluginData["x"].Paths[0] != `C:\a` || full.App.CommandLine == "" {
		t.Fatalf("plugin data or command line lost on reload: %+v err=%v", full.App, err)
	}
	if apps := e2.GetApps(); len(apps) != 1 || apps[0].SnapshotCount != 2 {
		t.Fatalf("apps after reload = %+v", apps)
	}
}

func TestLoadSkipsCorruptFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "broken.json.gz"), []byte("not gzip"), 0o644); err != nil {
		t.Fatal(err)
	}
	e := newTestEngine(t, dir, time.Now())
	if err := e.Load(); err != nil {
		t.Fatal(err)
	}
	if len(e.GetApps()) != 0 {
		t.Fatal("corrupt file must be ignored")
	}
}

func TestSaveOnlyWritesWhenChanged(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	e := newTestEngine(t, dir, now)
	e.Ingest(app(now, win(1, 0, 0, 100, 100)))
	if err := e.Save(); err != nil {
		t.Fatal(err)
	}
	f, _ := filepath.Glob(filepath.Join(dir, "*.json.gz"))
	st1, _ := os.Stat(f[0])
	time.Sleep(20 * time.Millisecond)
	if err := e.Save(); err != nil {
		t.Fatal(err)
	}
	st2, _ := os.Stat(f[0])
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Fatal("Save rewrote an unchanged timeline")
	}
}
