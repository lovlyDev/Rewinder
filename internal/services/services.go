//go:build windows

// Package services wires capture, history, restore, tray and hotkey together.
package services

import (
	"context"
	"log"
	"os"
	"sync"

	"Rewinder/internal/events"
	"Rewinder/internal/hotkey"
	"Rewinder/internal/ipcapi"
	"Rewinder/internal/plugins"
	"Rewinder/internal/policy"
	"Rewinder/internal/restore"
	"Rewinder/internal/snapshot"
	"Rewinder/internal/state"
	"Rewinder/internal/tray"
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
	pl  *plugins.Registry

	tray   *tray.Menu
	hotkey *hotkey.Listener

	selfPID int

	trackingMu sync.RWMutex
	trackingOn bool
	appPaused  map[string]bool

	restoreMu sync.Mutex

	started  bool
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

func (s *Services) showTimeline() {
	if s.deps.ShowTimelineWindow != nil {
		s.deps.ShowTimelineWindow()
	}
}

func (s *Services) Start(ctx context.Context) {
	s.started = true
	s.tray = &tray.Menu{
		OnOpenTimeline:   s.showTimeline,
		OnPauseTracking:  func() { s.PauseTracking(nil) },
		OnResumeTracking: func() { s.ResumeTracking(nil) },
		OnOpenSettings: func() {
			s.showTimeline()
			s.emit("onShowSettingsRequested", nil)
		},
		OnExit: func() { s.emit("onExitRequested", nil) },
	}
	s.hotkey = &hotkey.Listener{
		OnPress: s.onHotkey,
		OnError: func(err error) {
			log.Printf("rewinder: %v", err)
			s.emit("onTrackingStateChanged", ipcapi.TrackingStateChangedEvent{State: "error", Reason: err.Error(), AtUTC: ipcapi.NowUTC()})
		},
	}
	s.tray.Start()
	s.hotkey.Start()
	s.ev.StartWindowsSources()
	go s.captureLoop()

	// Record what is already open, so there is something to go back to right away.
	go s.captureAllVisible()

	s.emit("onTrackingStateChanged", ipcapi.TrackingStateChangedEvent{State: "active", AtUTC: ipcapi.NowUTC()})
}

func (s *Services) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopCh)
		if s.started {
			<-s.loopDone
			s.hotkey.Stop()
			s.tray.Stop()
		}
		s.ev.Stop()
		if err := s.ss.Close(); err != nil {
			log.Printf("rewinder: saving history: %v", err)
		}
	})
}

func (s *Services) onHotkey() {
	id := s.foregroundAppID()
	log.Printf("rewinder: hotkey pressed, foreground app %q", id)
	if s.deps.OnOverlayRequested != nil {
		s.deps.OnOverlayRequested(id)
	}
}
