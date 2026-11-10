//go:build windows

package services

import (
	"errors"
	"fmt"
	"log"

	"Rewinder/internal/ipcapi"
)

func (s *Services) GetApps() []ipcapi.AppSummary {
	apps := s.ss.GetApps()
	for i := range apps {
		apps[i].Excluded = s.excl.Excluded(apps[i].ExecutablePath)
	}
	return apps
}

func (s *Services) GetTimeline(appID string) []ipcapi.SnapshotMeta { return s.ss.GetTimeline(appID) }

// Restore puts an app back into one of its snapshots. Only one restore runs at a time.
func (s *Services) Restore(appID string, snapshotID string) error {
	if !s.restoreMu.TryLock() {
		return errors.New("restore.busy")
	}
	defer s.restoreMu.Unlock()

	progress := func(stage string, percent int, msg string) {
		s.emit("onRestoreProgress", ipcapi.RestoreProgressEvent{AppID: appID, SnapshotID: snapshotID, Stage: stage, Percent: percent, Message: msg, Code: "restore." + stage})
	}
	fail := func(err error) error {
		s.ss.ClearRestore(appID)
		log.Printf("rewinder: restore %s/%s failed: %v", appID, snapshotID, err)
		code, params := ipcapi.ErrorCode(err)
		s.emit("onRestoreError", ipcapi.RestoreErrorEvent{AppID: appID, SnapshotID: snapshotID, Error: code, Code: code, Params: params})
		return errors.New(code)
	}
	log.Printf("rewinder: restore requested %s/%s", appID, snapshotID)

	// Marked before moving anything: the capture of the result can come quickly.
	s.ss.MarkRestore(appID, snapshotID)
	progress("resolve", 5, "Resolving snapshot")
	snap, full, err := s.ss.ResolveSnapshot(appID, snapshotID)
	if err != nil {
		return fail(&ipcapi.Fault{Code: "restore.snapshot", Cause: err})
	}
	res, err := s.rs.RestoreSnapshot(progress, snap, full)
	if err != nil {
		return fail(err)
	}
	msg := fmt.Sprintf("Restored %d of %d window(s)", res.Matched, res.Saved)
	if res.Relaunched {
		msg += ", app relaunched"
	}
	log.Printf("rewinder: restore %s done: %s", appID, msg)
	s.emit("onRestoreProgress", ipcapi.RestoreProgressEvent{AppID: appID, SnapshotID: snapshotID, Stage: "done", Percent: 100, Message: msg, Code: "restore.done", Params: map[string]any{"matched": res.Matched, "saved": res.Saved, "relaunched": res.Relaunched}})
	return nil
}

// Busy reports whether a restore (of a snapshot or a layout) is running.
func (s *Services) Busy() bool {
	if s.restoreMu.TryLock() {
		s.restoreMu.Unlock()
		return false
	}
	return true
}
