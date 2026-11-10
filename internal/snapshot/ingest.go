package snapshot

import (
	"errors"
	"path/filepath"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/state"

	"github.com/google/uuid"
)

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

// trimLocked drops snapshots past retention and over the per-app limit.
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
