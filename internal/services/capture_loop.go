//go:build windows

package services

import (
	"log"
	"time"

	"Rewinder/internal/ipcapi"
	"Rewinder/internal/winapi"
)

const (
	captureQuiet    = 400 * time.Millisecond // capture once a window has been still this long
	captureMaxDelay = 3 * time.Second        // ...but never later than this after the first event
	capturePoll     = 100 * time.Millisecond // how often pending captures are checked
	saveInterval    = 30 * time.Second
)

type pendingCapture struct {
	first, last time.Time
}

// captureLoop debounces window events per process: a burst of move events
// (dragging a window) results in one capture of the final position.
// The poll ticker runs only while captures are pending, so an idle desktop
// costs no wakeups beyond the periodic save.
func (s *Services) captureLoop() {
	defer close(s.loopDone)
	pending := map[int]*pendingCapture{}
	poll := time.NewTicker(capturePoll)
	poll.Stop()
	polling := false
	save := time.NewTicker(saveInterval)
	defer save.Stop()
	defer poll.Stop()

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
			if !polling {
				poll.Reset(capturePoll)
				polling = true
			}
		case now := <-poll.C:
			for pid, p := range pending {
				if now.Sub(p.last) >= captureQuiet || now.Sub(p.first) >= captureMaxDelay {
					delete(pending, pid)
					s.captureProcess(pid)
				}
			}
			if len(pending) == 0 {
				poll.Stop()
				polling = false
			}
		case <-save.C:
			if err := s.ss.Save(); err != nil {
				log.Printf("rewinder: saving history: %v", err)
			}
		}
	}
}

// captureAllVisible records every app that already has a window open.
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
	if !s.trackingEnabled() {
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
