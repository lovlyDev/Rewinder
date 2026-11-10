//go:build windows

// Package restore puts an application back into a saved state: it relaunches
// the app or reopens its documents when needed, then restores window placement.
package restore

import (
	"errors"
	"fmt"
	"time"

	"Rewinder/internal/plugins"
	"Rewinder/internal/snapshot"
	"Rewinder/internal/state"
)

type ProgressFn func(stage string, percent int, msg string)

type Engine struct {
	cap     *state.CaptureEngine
	plugins *plugins.Registry

	LaunchTimeout time.Duration // how long to wait for a relaunched app's windows
	ReopenTimeout time.Duration // how long to wait for reopened documents/folders
	SettleDelay   time.Duration // apps often move their own window right after start
}

func NewEngine(cap *state.CaptureEngine, reg *plugins.Registry) *Engine {
	return &Engine{
		cap:           cap,
		plugins:       reg,
		LaunchTimeout: 15 * time.Second,
		ReopenTimeout: 8 * time.Second,
		SettleDelay:   800 * time.Millisecond,
	}
}

// Result says what a restore actually did.
type Result struct {
	Relaunched bool
	Reopened   int
	Matched    int
	Saved      int
}

func (e *Engine) RestoreSnapshot(progress ProgressFn, _ *snapshot.Snapshot, full *snapshot.FullSnapshot) (Result, error) {
	var res Result
	if full == nil {
		return res, errors.New("no snapshot state")
	}
	saved := full.App.Clone()
	res.Saved = len(saved.Windows)
	if saved.ExecutablePath == "" {
		return res, errors.New("snapshot has no executable path")
	}

	progress("find_instance", 10, "Looking for running instance")
	live := e.liveState(&saved)

	if len(live.Windows) == 0 {
		progress("relaunch", 25, "Starting "+saved.ExecutablePath)
		if err := e.relaunch(&saved); err != nil {
			return res, fmt.Errorf("could not start the app: %w", err)
		}
		res.Relaunched = true
		progress("wait_windows", 40, "Waiting for windows")
		live = e.waitForWindows(&saved, e.LaunchTimeout)
		if len(live.Windows) == 0 {
			return res, errors.New("the app started but no window appeared")
		}
	} else if n := e.plugins.Reopen(&saved, live); n > 0 {
		res.Reopened = n
		progress("reopen", 30, fmt.Sprintf("Reopening %d item(s)", n))
		live = e.waitForWindows(&saved, e.ReopenTimeout)
	}

	if len(saved.Windows) == 0 {
		return res, nil
	}

	progress("restore_windows", 60, "Restoring window positions")
	moved, err := e.apply(&saved, live)
	if err != nil {
		return res, err
	}
	if moved == 0 {
		return res, errors.New("none of the saved windows could be found")
	}
	res.Matched = moved
	if res.Relaunched || res.Reopened > 0 {
		time.Sleep(e.SettleDelay)
		if again, _ := e.apply(&saved, e.liveState(&saved)); again > 0 {
			res.Matched = again
		}
	}

	progress("restore_focus", 90, "Restoring focus")
	return res, nil
}
