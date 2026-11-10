// Package snapshot keeps a per-application timeline of captured states and
// persists it to disk, so history survives a restart of Rewinder or of the PC.
//
// Each snapshot stores the full AppState. A window list is a few KB, so this costs
// little and avoids the failure modes of delta chains (lost bases, stale PIDs).
package snapshot

import (
	"errors"
	"sort"
	"sync"
	"time"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/state"
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

func NewEngine(cfg EngineConfig) *Engine {
	if cfg.MaxSnapshotsPerApp <= 0 {
		cfg.MaxSnapshotsPerApp = 500
	}
	if cfg.Retention <= 0 {
		cfg.Retention = 7 * 24 * time.Hour
	}
	return &Engine{cfg: cfg, now: time.Now, apps: map[string]*appTimeline{}}
}

// GetApps lists tracked apps, most recently active first.
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

// ResolveSnapshot returns a private copy of one snapshot.
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

func meta(s *Snapshot, prev *state.AppState) ipcapi.SnapshotMeta {
	m := ipcapi.SnapshotMeta{
		SnapshotID:   s.SnapshotID,
		AppID:        s.AppID,
		Timestamp:    s.Timestamp.UTC().UnixMilli(),
		WindowsCount: len(s.State.Windows),
	}
	if prev != nil {
		m.FilesAdded, m.FilesRemoved = diffFileCounts(prev.OpenFiles, s.State.OpenFiles)
	} else {
		m.FilesAdded = len(s.State.OpenFiles)
	}
	return m
}
