//go:build windows

package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"Rewinder/internal/events"
	"Rewinder/internal/ipcapi"
	"Rewinder/internal/plugins"
	"Rewinder/internal/policy"
	"Rewinder/internal/restore"
	"Rewinder/internal/snapshot"
	"Rewinder/internal/state"
	"Rewinder/internal/trayhotkey"
	"Rewinder/internal/winapi"
)

const (
	captureQuiet    = 400 * time.Millisecond // capture once a window has been still this long
	captureMaxDelay = 3 * time.Second        // ...but never later than this after the first event
	saveInterval    = 30 * time.Second
)

type Dependencies struct {
	EmitEvent          func(name string, data any)
	OnOverlayRequested func(appID string)
	ShowTimelineWindow func()
}

type Services struct {
	deps Dependencies
	cfg  *policy.Config

	ev  *events.Bus
	cap *state.CaptureEngine
	ss  *snapshot.Engine
	rs  *restore.Engine
	th  *trayhotkey.Manager
	pl  *plugins.Registry

	selfPID int

	trackingMu sync.RWMutex
	trackingOn bool
	appPaused  map[string]bool

	restoreMu sync.Mutex

	stopOnce sync.Once
	stopCh   chan struct{}
	loopDone chan struct{}
}

func New(deps Dependencies) *Services {
	cfg := policy.DefaultConfig()
	pl := plugins.DefaultRegistry()
	cap := state.NewCaptureEngine()
	cap.NeedsOpenFiles = pl.NeedsOpenFiles

	ss := snapshot.NewEngine(snapshot.EngineConfig{
		MaxSnapshotsPerApp: cfg.MaxSnapshotsPerApp,
		Retention:          cfg.Retention,
		StorageDir:         cfg.StorageDir,
	})
	if err := ss.Load(); err != nil {
		log.Printf("rewinder: loading history: %v", err)
	}

	return &Services{
		deps:       deps,
		cfg:        cfg,
		ev:         events.NewBus(4096),
		cap:        cap,
		ss:         ss,
		rs:         restore.NewEngine(cap, pl),
		pl:         pl,
		selfPID:    os.Getpid(),
		trackingOn: true,
		appPaused:  map[string]bool{},
		stopCh:     make(chan struct{}),
		loopDone:   make(chan struct{}),
	}
}

func (s *Services) emit(name string, data any) {
	if s.deps.EmitEvent != nil {
		s.deps.EmitEvent(name, data)
	}
}

func (s *Services) Start(ctx context.Context) {
	s.th = trayhotkey.NewManager(trayhotkey.Dependencies{
		OnOpenTimeline: func() {
			if s.deps.ShowTimelineWindow != nil {
				s.deps.ShowTimelineWindow()
			}
		},
		OnPauseTracking:  func() { s.PauseTracking(nil) },
		OnResumeTracking: func() { s.ResumeTracking(nil) },
		OnOpenSettings: func() {
			if s.deps.ShowTimelineWindow != nil {
				s.deps.ShowTimelineWindow()
			}
			s.emit("onShowSettingsRequested", nil)
		},
		OnExit: func() { s.emit("onExitRequested", nil) },
		OnOverlayHotkey: func() {
			id := s.foregroundAppID()
			log.Printf("rewinder: hotkey pressed, foreground app %q", id)
			if s.deps.OnOverlayRequested != nil {
				s.deps.OnOverlayRequested(id)
			}
		},
		OnHotkeyError: func(err error) {
			log.Printf("rewinder: %v", err)
			s.emit("onTrackingStateChanged", ipcapi.TrackingStateChangedEvent{State: "error", Reason: err.Error(), AtUTC: ipcapi.NowUTC()})
		},
	})
	s.th.Start()
	s.ev.StartWindowsSources()
	go s.captureLoop()

	// Record what is already open, so there is something to go back to right away.
	go s.captureAllVisible()

	s.emit("onTrackingStateChanged", ipcapi.TrackingStateChangedEvent{State: "active", AtUTC: ipcapi.NowUTC()})
}

func (s *Services) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		if s.th != nil { // Start ran, so the capture loop is running
			<-s.loopDone
			s.th.Stop()
		}
		s.ev.Stop()
		if err := s.ss.Close(); err != nil {
			log.Printf("rewinder: saving history: %v", err)
		}
	})
}

// foregroundAppID is the app the user was in when they pressed the hotkey.
func (s *Services) foregroundAppID() string {
	pid := int(winapi.WindowPID(winapi.ForegroundWindow()))
	if pid == 0 || pid == s.selfPID {
		return ""
	}
	return state.StableAppID(winapi.ExePath(uint32(pid)))
}

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

func (s *Services) GetApps() []ipcapi.AppSummary { return s.ss.GetApps() }

func (s *Services) GetTimeline(appID string) []ipcapi.SnapshotMeta { return s.ss.GetTimeline(appID) }

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

func (s *Services) shouldTrack(appID, exePath, windowClass string) bool {
	s.trackingMu.RLock()
	on, paused := s.trackingOn, s.appPaused[appID]
	s.trackingMu.RUnlock()
	return on && !paused && s.cfg.Rules.Allow(exePath, windowClass)
}

type pendingCapture struct {
	first, last time.Time
}

// captureLoop debounces window events per process: a burst of move events
// (dragging a window) results in one capture of the final position.
func (s *Services) captureLoop() {
	defer close(s.loopDone)
	pending := map[int]*pendingCapture{}
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	save := time.NewTicker(saveInterval)
	defer save.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case ev := <-s.ev.Events():
			if ev.PID <= 0 || ev.PID == s.selfPID {
				continue
			}
			now := time.Now()
			if p := pending[ev.PID]; p != nil {
				p.last = now
			} else {
				pending[ev.PID] = &pendingCapture{first: now, last: now}
			}
		case now := <-tick.C:
			for pid, p := range pending {
				if now.Sub(p.last) >= captureQuiet || now.Sub(p.first) >= captureMaxDelay {
					delete(pending, pid)
					s.captureProcess(pid)
				}
			}
		case <-save.C:
			if err := s.ss.Save(); err != nil {
				log.Printf("rewinder: saving history: %v", err)
			}
		}
	}
}

func (s *Services) captureAllVisible() {
	pids := map[int]bool{}
	for _, h := range winapi.TopLevelWindows() {
		if winapi.IsAppWindow(h) {
			pids[int(winapi.WindowPID(h))] = true
		}
	}
	for pid := range pids {
		if pid != s.selfPID {
			s.captureProcess(pid)
		}
	}
}

func (s *Services) captureProcess(pid int) {
	s.trackingMu.RLock()
	on := s.trackingOn
	s.trackingMu.RUnlock()
	if !on {
		return
	}
	app, err := s.cap.CaptureProcess(pid)
	if err != nil {
		return // process gone or no windows: keep the last good snapshot
	}
	if !s.shouldTrack(app.AppID, app.ExecutablePath, app.ForegroundWindowClass) {
		return
	}
	s.pl.Capture(app)
	meta, err := s.ss.Ingest(app)
	if err != nil || meta == nil {
		return
	}
	s.emit("onSnapshotCreated", ipcapi.SnapshotCreatedEvent{AppID: app.AppID, Snapshot: *meta, OccurredAt: ipcapi.NowUTC()})
}
