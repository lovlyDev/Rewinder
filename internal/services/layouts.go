//go:build windows

package services

import (
	"errors"
	"log"
	"path/filepath"
	"time"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/layouts"
	"Rewinder/internal/snapshot"
	"Rewinder/internal/state"
	"Rewinder/internal/tray"
	"Rewinder/internal/winapi"
)

// LayoutSummary is a layout as the UI lists it.
type LayoutSummary struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	CreatedAt int64              `json:"createdAtUTC"`
	Shortcut  int                `json:"shortcut"` // 1…9 for Ctrl+Alt+N, 0 = none
	Apps      []LayoutAppSummary `json:"apps"`
}

type LayoutAppSummary struct {
	AppID   string `json:"appID"`
	Name    string `json:"name"`
	Windows int    `json:"windows"`
}

func loadLayouts(storageDir string) *layouts.Store {
	return layouts.Load(filepath.Dir(storageDir))
}

func summarize(list []layouts.Layout) []LayoutSummary {
	out := make([]LayoutSummary, 0, len(list))
	for i, l := range list {
		s := LayoutSummary{ID: l.ID, Name: l.Name, CreatedAt: l.CreatedAt.UTC().UnixMilli(), Apps: []LayoutAppSummary{}}
		if i < tray.MaxLayoutItems {
			s.Shortcut = i + 1
		}
		for _, a := range l.Apps {
			s.Apps = append(s.Apps, LayoutAppSummary{AppID: a.AppID, Name: filepath.Base(a.ExecutablePath), Windows: len(a.Windows)})
		}
		out = append(out, s)
	}
	return out
}

func (s *Services) Layouts() []LayoutSummary { return summarize(s.lay.List()) }

// SaveLayout records every tracked application that has a window open now.
func (s *Services) SaveLayout(name string) (LayoutSummary, error) {
	var apps []state.AppState
	seen := map[int]bool{}
	for _, h := range winapi.TopLevelWindows() {
		if !winapi.IsAppWindow(h) {
			continue
		}
		pid := int(winapi.WindowPID(h))
		if pid == s.selfPID || seen[pid] {
			continue
		}
		seen[pid] = true
		app, err := s.cap.CaptureProcess(pid)
		if err != nil || s.excl.Excluded(app.ExecutablePath) || !s.cfg.Rules.Allow(app.ExecutablePath, app.ForegroundWindowClass) {
			continue
		}
		s.pl.Capture(app)
		apps = append(apps, app.Clone())
	}
	l, err := s.lay.Add(name, apps, time.Now())
	if err != nil {
		return LayoutSummary{}, err
	}
	s.layoutsChanged()
	for _, sum := range s.Layouts() {
		if sum.ID == l.ID {
			return sum, nil
		}
	}
	return LayoutSummary{}, errors.New("layout.missing")
}

func (s *Services) RenameLayout(id, name string) error {
	if err := s.lay.Rename(id, name); err != nil {
		return err
	}
	s.layoutsChanged()
	return nil
}

func (s *Services) DeleteLayout(id string) error {
	if err := s.lay.Delete(id); err != nil {
		return err
	}
	s.layoutsChanged()
	return nil
}

func (s *Services) MoveLayout(id string, to int) error {
	if err := s.lay.Move(id, to); err != nil {
		return err
	}
	s.layoutsChanged()
	return nil
}

// LayoutResult says how a layout restore went.
type LayoutResult struct {
	Name     string   `json:"name"`
	Restored int      `json:"restored"`
	Total    int      `json:"total"`
	Failed   []string `json:"failed"`
}

// RestoreLayout puts every application of a layout back, one after another
// (relaunching the closed ones). It fails only when nothing could be restored.
func (s *Services) RestoreLayout(id string) (LayoutResult, error) {
	l, ok := s.lay.Get(id)
	if !ok {
		return LayoutResult{}, errors.New("layout.missing")
	}
	if !s.restoreMu.TryLock() {
		return LayoutResult{}, errors.New("restore.busy")
	}
	defer s.restoreMu.Unlock()

	res := LayoutResult{Name: l.Name, Total: len(l.Apps), Failed: []string{}}
	log.Printf("rewinder: restoring layout %q (%d apps)", l.Name, len(l.Apps))
	for i, app := range l.Apps {
		s.emit("onLayoutProgress", map[string]any{"id": id, "done": i, "total": len(l.Apps), "app": filepath.Base(app.ExecutablePath)})
		full := &snapshot.FullSnapshot{App: app.Clone()}
		if _, err := s.rs.RestoreSnapshot(func(string, int, string) {}, nil, full); err != nil {
			log.Printf("rewinder: layout %q: %s: %v", l.Name, app.ExecutablePath, err)
			res.Failed = append(res.Failed, filepath.Base(app.ExecutablePath))
			continue
		}
		res.Restored++
	}
	s.emit("onLayoutRestored", res)
	if res.Restored == 0 {
		return res, errors.New("layout.failed")
	}
	return res, nil
}

// RestoreLayoutAt restores the layout in that position (tray item, Ctrl+Alt+N).
func (s *Services) RestoreLayoutAt(i int) {
	list := s.lay.List()
	if i < 0 || i >= len(list) {
		return
	}
	if _, err := s.RestoreLayout(list[i].ID); err != nil {
		log.Printf("rewinder: layout %d: %v", i+1, err)
		code, params := ipcapi.ErrorCode(err)
		s.emit("onLayoutError", map[string]any{"code": code, "params": params})
	}
}

// LayoutPreview merges the windows of every app in a layout for the minimap.
func (s *Services) LayoutPreview(id string) (*ipcapi.SnapshotLayout, error) {
	l, ok := s.lay.Get(id)
	if !ok {
		return nil, errors.New("layout.missing")
	}
	out := &ipcapi.SnapshotLayout{Monitors: []ipcapi.LayoutMonitor{}, Windows: []ipcapi.LayoutWindow{}}
	for i := range l.Apps {
		one := layoutOf(&l.Apps[i])
		out.Monitors = one.Monitors
		out.Windows = append(out.Windows, one.Windows...)
	}
	return out, nil
}

func (s *Services) layoutsChanged() {
	list := s.lay.List()
	names := make([]string, len(list))
	for i, l := range list {
		names[i] = l.Name
	}
	if s.tray != nil {
		s.tray.SetLayouts(names)
	}
	s.emit("onLayoutsChanged", nil)
}
