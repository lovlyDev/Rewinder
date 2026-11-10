//go:build windows

package services

import (
	"errors"
	"fmt"
	"log"

	"Rewinder/internal/ipcapi"
)

func (s *Services) GetApps() []ipcapi.AppSummary { return s.ss.GetApps() }

func (s *Services) GetTimeline(appID string) []ipcapi.SnapshotMeta { return s.ss.GetTimeline(appID) }

// Restore puts an app back into one of its snapshots. Only one restore runs at a time.
func (s *Services) Restore(appID string, snapshotID string) error {
	if !s.restoreMu.TryLock() {
		return errors.New("another restore is in progress")
	}
	defer s.restoreMu.Unlock()

	progress := func(stage string, percent int, msg string) {
		s.emit("onRestoreProgress", ipcapi.RestoreProgressEvent{AppID: appID, SnapshotID: snapshotID, Stage: stage, Percent: percent, Message: msg})
	}
	fail := func(err error) error {
		log.Printf("rewinder: restore %s/%s failed: %v", appID, snapshotID, err)
		s.emit("onRestoreError", ipcapi.RestoreErrorEvent{AppID: appID, SnapshotID: snapshotID, Error: err.Error()})
		return err
	}
	log.Printf("rewinder: restore requested %s/%s", appID, snapshotID)

	progress("resolve", 5, "Resolving snapshot")
	snap, full, err := s.ss.ResolveSnapshot(appID, snapshotID)
	if err != nil {
		return fail(err)
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
	progress("done", 100, msg)
	return nil
}
