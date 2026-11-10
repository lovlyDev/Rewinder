package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"Rewinder/internal/update"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// UpdateInfo is what the settings page shows about updates.
type UpdateInfo struct {
	Current     string `json:"current"`
	Version     string `json:"version"`
	Notes       string `json:"notes"`
	URL         string `json:"url"`
	Available   bool   `json:"available"`
	Installable bool   `json:"installable"`
}

var (
	updateMu       sync.Mutex
	latest         *update.Release
	installing     bool
	autoCheckDue   = startDelay()
	autoCheckEvery = 6 * time.Hour
)

// CheckForUpdate asks GitHub Releases for the newest version.
func (a *App) CheckForUpdate() (UpdateInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	rel, err := update.Latest(ctx)
	if err != nil {
		log.Printf("rewinder: update check: %v", err)
		return UpdateInfo{Current: version}, errors.New("update.check")
	}
	updateMu.Lock()
	latest = rel
	updateMu.Unlock()
	return UpdateInfo{
		Current:     version,
		Version:     rel.Version,
		Notes:       rel.Notes,
		URL:         rel.URL,
		Available:   update.Newer(rel.Version, version),
		Installable: rel.Installable(),
	}, nil
}

// InstallUpdate downloads the release found by CheckForUpdate, verifies its
// checksum and signature, swaps the executable and restarts Rewinder.
func (a *App) InstallUpdate() error {
	updateMu.Lock()
	rel := latest
	if installing {
		updateMu.Unlock()
		return errors.New("update.busy")
	}
	installing = true
	updateMu.Unlock()
	defer func() {
		updateMu.Lock()
		installing = false
		updateMu.Unlock()
	}()
	if rel == nil || !update.Newer(rel.Version, version) {
		return errors.New("update.none")
	}
	exe, err := os.Executable()
	if err != nil {
		return errors.New("update.failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	path, err := update.Download(ctx, rel, exe, func(p int) {
		runtime.EventsEmit(a.ctx, "onUpdateProgress", p)
	})
	if err != nil {
		log.Printf("rewinder: update download: %v", err)
		if err.Error() == "update.checksum" || err.Error() == "update.assets" {
			return err
		}
		if os.IsPermission(err) {
			return errors.New("update.permission")
		}
		return errors.New("update.download")
	}
	if err := update.CheckSignature(exe, path); err != nil {
		_ = os.Remove(path)
		return err
	}
	if err := update.Swap(exe, path); err != nil {
		_ = os.Remove(path)
		log.Printf("rewinder: update swap: %v", err)
		return errors.New("update.permission") // e.g. installed under Program Files
	}
	log.Printf("rewinder: updated to %s, restarting", rel.Version)
	args := []string{fmt.Sprintf("%s%d", waitPidFlag, os.Getpid())}
	if !a.GetWindowVisible() {
		args = append(args, "--autostart") // it was in the tray: come back in the tray
	}
	cmd := exec.Command(exe, args...)
	if err := cmd.Start(); err != nil {
		return errors.New("update.restart")
	}
	go func() {
		time.Sleep(300 * time.Millisecond)
		a.ExitApp()
	}()
	return nil
}

// autoCheckUpdate checks a little after start and then every few hours, as the
// "updates" setting says: offer the new version, or install it right away.
func (a *App) autoCheckUpdate() {
	if version == "dev" {
		return // local builds have nothing to compare with
	}
	time.Sleep(autoCheckDue)
	for {
		a.autoCheckOnce()
		time.Sleep(autoCheckEvery)
	}
}

func (a *App) autoCheckOnce() {
	a.settingsMu.Lock()
	mode := a.preferences.Updates
	a.settingsMu.Unlock()
	if mode == "" {
		mode = "install" // settings saved before this option existed
	}
	if mode == "off" {
		return
	}
	info, err := a.CheckForUpdate()
	if err != nil || !info.Available {
		return
	}
	runtime.EventsEmit(a.ctx, "onUpdateAvailable", info)
	if mode != "install" || !info.Installable {
		return
	}
	// Never in the middle of something: not while the quick-restore overlay is
	// open or a restore is running.
	for a.overlayOpen() || a.svc.Busy() {
		time.Sleep(30 * time.Second)
	}
	log.Printf("rewinder: installing update %s automatically", info.Version)
	if err := a.InstallUpdate(); err != nil {
		log.Printf("rewinder: automatic update: %v", err)
		runtime.EventsEmit(a.ctx, "onUpdateFailed", err.Error())
	}
}

func (a *App) overlayOpen() bool {
	a.windowMu.Lock()
	defer a.windowMu.Unlock()
	return a.overlay
}

// startDelay: 15 s normally; REWINDER_UPDATE_DELAY (seconds) shortens it in tests.
func startDelay() time.Duration {
	if v, err := time.ParseDuration(os.Getenv("REWINDER_UPDATE_DELAY") + "s"); err == nil && v > 0 {
		return v
	}
	return 15 * time.Second
}
