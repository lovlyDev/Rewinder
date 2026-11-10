// Package snapshot keeps a per-application timeline of captured states and
// persists it to disk, so history survives a restart of Rewinder or of the PC.
//
// Each snapshot stores the full AppState. A window list is a few KB, so this costs
// little and avoids the failure modes of delta chains (lost bases, stale PIDs).
package snapshot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/state"

	"github.com/google/uuid"
)

type EngineConfig struct {
	MaxSnapshotsPerApp int
	Retention          time.Duration
	StorageDir         string // empty disables persistence
}

type Engine struct {
	cfg EngineConfig
	now func() time.Time

	mu   sync.RWMutex
	apps map[string]*appTimeline
}

type appTimeline struct {
	AppID        string     `json:"appID"`
	Exe          string     `json:"exe"`
	Name         string     `json:"name"`
	LastActivity time.Time  `json:"lastActivity"`
	Snapshots    []Snapshot `json:"snapshots"`

	dirty bool
}

type Snapshot struct {
	SnapshotID string         `json:"snapshotID"`
	AppID      string         `json:"appID"`
	Timestamp  time.Time      `json:"timestamp"`
	State      state.AppState `json:"state"`
}

// FullSnapshot is the resolved state handed to the restore engine.
type FullSnapshot struct {
	App state.AppState `json:"app"`
}

const fileFormatVersion = 2

type timelineFile struct {
	Version  int          `json:"version"`
	Timeline *appTimeline `json:"timeline"`
}

func NewEngine(cfg EngineConfig) *Engine {
	if cfg.MaxSnapshotsPerApp <= 0 {
		cfg.MaxSnapshotsPerApp = 500
	}
	if cfg.Retention <= 0 {
		cfg.Retention = 7 * 24 * time.Hour
	}
	return &Engine{cfg: cfg, now: time.Now, apps: map[string]*appTimeline{}}
}

func (e *Engine) GetApps() []ipcapi.AppSummary {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]ipcapi.AppSummary, 0, len(e.apps))
	for _, tl := range e.apps {
		if len(tl.Snapshots) == 0 {
			continue
		}
		out = append(out, ipcapi.AppSummary{
			AppID:           tl.AppID,
			Name:            tl.Name,
			ExecutablePath:  tl.Exe,
			LastActivityUTC: tl.LastActivity.UTC().UnixMilli(),
			SnapshotCount:   len(tl.Snapshots),
			TrackingState:   "active",
			RetentionStatus: "ok",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].LastActivityUTC > out[j].LastActivityUTC })
	return out
}

// GetTimeline returns snapshot metadata, newest first.
func (e *Engine) GetTimeline(appID string) []ipcapi.SnapshotMeta {
	e.mu.RLock()
	defer e.mu.RUnlock()
	tl := e.apps[appID]
	if tl == nil {
		return []ipcapi.SnapshotMeta{}
	}
	out := make([]ipcapi.SnapshotMeta, 0, len(tl.Snapshots))
	for i := len(tl.Snapshots) - 1; i >= 0; i-- {
		var prev *state.AppState
		if i > 0 {
			prev = &tl.Snapshots[i-1].State
		}
		out = append(out, meta(&tl.Snapshots[i], prev))
	}
	return out
}

func meta(s *Snapshot, prev *state.AppState) ipcapi.SnapshotMeta {
	m := ipcapi.SnapshotMeta{
		SnapshotID:   s.SnapshotID,
		AppID:        s.AppID,
		Timestamp:    s.Timestamp.UTC().UnixMilli(),
		WindowsCount: len(s.State.Windows),
	}
	if prev != nil {
		added, removed := diffFiles(prev.OpenFiles, s.State.OpenFiles)
		m.FilesAdded, m.FilesRemoved = len(added), len(removed)
	} else {
		m.FilesAdded = len(s.State.OpenFiles)
	}
	return m
}

// Ingest records a captured state. It returns nil metadata when the state does
// not differ meaningfully from the latest snapshot of the same app.
func (e *Engine) Ingest(app *state.AppState) (*ipcapi.SnapshotMeta, error) {
	if app == nil || app.AppID == "" {
		return nil, errors.New("empty app state")
	}
	if app.Timestamp.IsZero() {
		app.Timestamp = e.now()
	}
	if app.Timestamp.Before(e.now().Add(-e.cfg.Retention)) {
		return nil, nil // already past retention: it would be dropped right away
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	tl := e.apps[app.AppID]
	if tl == nil {
		tl = &appTimeline{AppID: app.AppID}
		e.apps[app.AppID] = tl
	}
	if app.ExecutablePath != "" {
		tl.Exe = app.ExecutablePath
		tl.Name = filepath.Base(app.ExecutablePath)
	}
	tl.LastActivity = app.Timestamp
	e.trimLocked(tl)

	var prev *state.AppState
	if n := len(tl.Snapshots); n > 0 {
		prev = &tl.Snapshots[n-1].State
		if !Changed(prev, app) {
			return nil, nil
		}
	}

	snap := Snapshot{
		SnapshotID: uuid.NewString(),
		AppID:      app.AppID,
		Timestamp:  app.Timestamp,
		State:      app.Clone(),
	}
	m := meta(&snap, prev) // before trimming, which may drop prev
	tl.Snapshots = append(tl.Snapshots, snap)
	tl.dirty = true
	e.trimLocked(tl)
	return &m, nil
}

func (e *Engine) ResolveSnapshot(appID, snapshotID string) (*Snapshot, *FullSnapshot, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	tl := e.apps[appID]
	if tl == nil {
		return nil, nil, errors.New("unknown app")
	}
	for i := range tl.Snapshots {
		if tl.Snapshots[i].SnapshotID == snapshotID {
			s := tl.Snapshots[i]
			s.State = s.State.Clone()
			return &s, &FullSnapshot{App: s.State.Clone()}, nil
		}
	}
	return nil, nil, errors.New("snapshot not found")
}

func (e *Engine) trimLocked(tl *appTimeline) {
	cut := e.now().Add(-e.cfg.Retention)
	i := 0
	for i < len(tl.Snapshots) && tl.Snapshots[i].Timestamp.Before(cut) {
		i++
	}
	if over := len(tl.Snapshots) - i - e.cfg.MaxSnapshotsPerApp; over > 0 {
		i += over
	}
	if i > 0 {
		tl.Snapshots = append([]Snapshot(nil), tl.Snapshots[i:]...)
		tl.dirty = true
	}
}

// Changed reports whether next differs from prev in a way worth a new snapshot.
// Focus, Z-order and window titles change constantly and alone do not count.
func Changed(prev, next *state.AppState) bool {
	if prev.PID != next.PID || len(prev.Windows) != len(next.Windows) {
		return true
	}
	pw := map[uintptr]state.WindowState{}
	for _, w := range prev.Windows {
		pw[w.HWND] = w
	}
	for _, nw := range next.Windows {
		ow, ok := pw[nw.HWND]
		if !ok || !sameGeometry(&ow, &nw) || ow.Path != nw.Path {
			return true
		}
	}
	added, removed := diffFiles(prev.OpenFiles, next.OpenFiles)
	if len(added) > 0 || len(removed) > 0 {
		return true
	}
	return !pluginEq(prev.PluginData, next.PluginData)
}

func sameGeometry(a, b *state.WindowState) bool {
	return a.NormalRect == b.NormalRect &&
		a.Rect == b.Rect &&
		a.ShowCmd == b.ShowCmd &&
		a.MonitorID == b.MonitorID
}

func diffFiles(prev, next []state.FileRef) (added, removed []string) {
	p := map[string]bool{}
	for _, f := range prev {
		p[strings.ToLower(f.Path)] = true
	}
	n := map[string]bool{}
	for _, f := range next {
		k := strings.ToLower(f.Path)
		n[k] = true
		if !p[k] {
			added = append(added, f.Path)
		}
	}
	for _, f := range prev {
		if !n[strings.ToLower(f.Path)] {
			removed = append(removed, f.Path)
		}
	}
	return added, removed
}

func pluginEq(a, b map[string]state.PluginState) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return bytes.Equal(aj, bj)
}

// ---- persistence ----

func (e *Engine) fileFor(appID string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		}
		return '_'
	}, appID)
	return filepath.Join(e.cfg.StorageDir, safe+".json.gz")
}

// Load reads persisted timelines. Missing or corrupt files are skipped.
func (e *Engine) Load() error {
	if e.cfg.StorageDir == "" {
		return nil
	}
	files, err := filepath.Glob(filepath.Join(e.cfg.StorageDir, "*.json.gz"))
	if err != nil {
		return err
	}
	if leftovers, _ := filepath.Glob(filepath.Join(e.cfg.StorageDir, "*.json.gz.tmp")); len(leftovers) > 0 {
		for _, f := range leftovers {
			_ = os.Remove(f) // from an interrupted save
		}
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, f := range files {
		tl, err := readTimeline(f)
		if err != nil || tl == nil || tl.AppID == "" {
			continue
		}
		e.trimLocked(tl)
		tl.dirty = false
		if len(tl.Snapshots) == 0 {
			_ = os.Remove(f)
			continue
		}
		e.apps[tl.AppID] = tl
	}
	return nil
}

func readTimeline(path string) (*appTimeline, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	raw, err := io.ReadAll(io.LimitReader(zr, 64<<20))
	if err != nil {
		return nil, err
	}
	var tf timelineFile
	if err := json.Unmarshal(raw, &tf); err != nil {
		return nil, err
	}
	if tf.Version != fileFormatVersion {
		return nil, errors.New("unsupported timeline version")
	}
	return tf.Timeline, nil
}

// Save writes every changed timeline to disk atomically.
func (e *Engine) Save() error {
	if e.cfg.StorageDir == "" {
		return nil
	}
	if err := os.MkdirAll(e.cfg.StorageDir, 0o755); err != nil {
		return err
	}
	type job struct {
		path string
		data []byte
		tl   *appTimeline
	}
	var jobs []job
	e.mu.Lock()
	for _, tl := range e.apps {
		if !tl.dirty {
			continue
		}
		raw, err := json.Marshal(timelineFile{Version: fileFormatVersion, Timeline: tl})
		if err != nil {
			continue
		}
		tl.dirty = false
		jobs = append(jobs, job{path: e.fileFor(tl.AppID), data: raw, tl: tl})
	}
	e.mu.Unlock()

	var firstErr error
	for _, j := range jobs {
		if err := writeGzipAtomic(j.path, j.data); err != nil {
			e.mu.Lock()
			j.tl.dirty = true
			e.mu.Unlock()
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func writeGzipAtomic(path string, data []byte) error {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		// Some file system filters (sandboxes, sync clients, AV) refuse the rename.
		// Fall back to a direct write: a torn file is skipped by Load anyway.
		_ = os.Remove(tmp)
		return os.WriteFile(path, buf.Bytes(), 0o644)
	}
	return nil
}

func (e *Engine) Close() error { return e.Save() }
