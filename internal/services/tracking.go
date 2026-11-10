//go:build windows

package services

import (
	"Rewinder/internal/ipcapi"
	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

// PauseTracking stops recording one app, or every app when appID is nil.
func (s *Services) PauseTracking(appID *string) {
	s.trackingMu.Lock()
	if appID == nil {
		s.trackingOn = false
	} else {
		s.appPaused[*appID] = true
	}
	s.trackingMu.Unlock()
	s.emit("onTrackingStateChanged", ipcapi.TrackingStateChangedEvent{AppID: appID, State: "paused", Reason: "paused by user", AtUTC: ipcapi.NowUTC()})
}

// ResumeTracking undoes PauseTracking.
func (s *Services) ResumeTracking(appID *string) {
	s.trackingMu.Lock()
	if appID == nil {
		s.trackingOn = true
	} else {
		delete(s.appPaused, *appID)
	}
	s.trackingMu.Unlock()
	s.emit("onTrackingStateChanged", ipcapi.TrackingStateChangedEvent{AppID: appID, State: "active", Reason: "resumed by user", AtUTC: ipcapi.NowUTC()})
}

func (s *Services) trackingEnabled() bool {
	s.trackingMu.RLock()
	defer s.trackingMu.RUnlock()
	return s.trackingOn
}

func (s *Services) shouldTrack(appID, exePath, windowClass string) bool {
	s.trackingMu.RLock()
	on, paused := s.trackingOn, s.appPaused[appID]
	s.trackingMu.RUnlock()
	return on && !paused && !s.excl.Excluded(exePath) && s.cfg.Rules.Allow(exePath, windowClass)
}

// foregroundAppID is the app the user was in when they pressed the hotkey.
func (s *Services) foregroundAppID() string {
	pid := int(winapi.WindowPID(winapi.ForegroundWindow()))
	if pid == 0 || pid == s.selfPID {
		return ""
	}
	return state.StableAppID(winapi.ExePath(uint32(pid)))
}

// Exclusions lists executables the user chose not to track.
func (s *Services) Exclusions() []string { return s.excl.List() }

// SetExcluded stops (or resumes) tracking an executable, by file name.
func (s *Services) SetExcluded(exe string, excluded bool) error { return s.excl.Set(exe, excluded) }

// DeleteAppHistory forgets every snapshot of one app, on disk too.
func (s *Services) DeleteAppHistory(appID string) error { return s.ss.DeleteApp(appID) }
