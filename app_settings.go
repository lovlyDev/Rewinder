package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"Rewinder/internal/hotkey"
	"Rewinder/internal/services"
	"Rewinder/internal/settings"
	"Rewinder/internal/winapi"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func (a *App) GetSettings() settings.Settings {
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	return a.preferences
}
func (a *App) SaveSettings(s settings.Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	return a.saveSettingsLocked(s)
}
func (a *App) saveSettingsLocked(s settings.Settings) error {
	old := a.preferences
	if a.svc == nil {
		return errNotReady
	}
	if s.Hotkey != old.Hotkey {
		if err := a.svc.SetHotkey(s.Hotkey); err != nil {
			return err
		}
	}
	if err := settings.Save(settings.Directory(), s); err != nil {
		log.Printf("rewinder: saving settings: %v", err)
		if s.Hotkey != old.Hotkey {
			if rollback := a.svc.SetHotkey(old.Hotkey); rollback != nil {
				log.Printf("hotkey rollback: %v", rollback)
			}
		}
		return errors.New("settings.write")
	}
	a.preferences = s
	language := s.Language
	if language == "system" {
		language = winapi.SystemLanguage()
	}
	a.svc.UpdatePreferences(s.Hotkey, language)
	return nil
}
func (a *App) SetHotkey(mods, key uint32) error {
	combo := hotkey.Combination{Mods: mods, Key: key}
	if err := combo.Validate(); err != nil {
		return err
	}
	a.settingsMu.Lock()
	defer a.settingsMu.Unlock()
	s := a.preferences
	s.Hotkey = hotkey.Combination{Mods: mods, Key: key}
	return a.saveSettingsLocked(s)
}

type AppInfo struct {
	Version           string `json:"version"`
	Elevated          bool   `json:"elevated"`          // running as administrator now
	ElevationDeclined bool   `json:"elevationDeclined"` // "run as administrator" is on but UAC was declined
	LogPath           string `json:"logPath"`
	HistoryPath       string `json:"historyPath"`
	SystemLanguage    string `json:"systemLanguage"`
}

func (a *App) GetAppInfo() AppInfo {
	dir := settings.Directory()
	return AppInfo{
		Version:           version,
		Elevated:          winapi.SelfElevated(),
		ElevationDeclined: a.elevationDeclined,
		LogPath:           filepath.Join(dir, "rewinder.log"),
		HistoryPath:       filepath.Join(dir, "history"),
		SystemLanguage:    winapi.SystemLanguage(),
	}
}
func (a *App) GetBackgroundImage() (string, error) {
	b, err := os.ReadFile(filepath.Join(settings.Directory(), "background.png"))
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("background.read")
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(b), nil
}
func (a *App) PickBackgroundImage() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{Filters: []runtime.FileFilter{{DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg;*.gif"}}})
	if err != nil {
		return "", errors.New("background.read")
	}
	if path == "" {
		return "", nil
	}
	f, err := os.Open(path)
	if err != nil {
		return "", errors.New("background.read")
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, 20<<20+1))
	if err != nil || len(raw) > 20<<20 {
		return "", errors.New("background.size")
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width <= 0 || config.Height <= 0 || int64(config.Width)*int64(config.Height) > 24000000 {
		return "", errors.New("background.size")
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return "", errors.New("background.read")
	}
	var data bytes.Buffer
	if err = png.Encode(&data, img); err != nil {
		return "", errors.New("background.read")
	}
	dir := settings.Directory()
	if err = os.MkdirAll(dir, 0700); err != nil {
		return "", errors.New("settings.write")
	}
	if err = os.WriteFile(filepath.Join(dir, "background.png"), data.Bytes(), 0600); err != nil {
		return "", errors.New("settings.write")
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(data.Bytes()), nil
}

// SetRunAsAdmin saves the "run as administrator" choice. Switching it on while
// not elevated restarts Rewinder through the UAC prompt; declining keeps the
// setting off. Switching it off applies at the next start (an elevated process
// cannot drop its rights).
func (a *App) SetRunAsAdmin(on bool) error {
	a.settingsMu.Lock()
	s := a.preferences
	s.RunAsAdmin = on
	err := a.saveSettingsLocked(s)
	a.settingsMu.Unlock()
	if err != nil {
		return err
	}
	if on && !winapi.SelfElevated() {
		if err := restartElevated(); err != nil {
			a.settingsMu.Lock()
			s.RunAsAdmin = false
			_ = a.saveSettingsLocked(s)
			a.settingsMu.Unlock()
			if errors.Is(err, winapi.ErrCancelled) {
				return errors.New("admin.cancelled")
			}
			log.Printf("rewinder: restart as administrator: %v", err)
			return errors.New("admin.failed")
		}
		go func() {
			time.Sleep(300 * time.Millisecond) // let the UI get its answer first
			a.ExitApp()
		}()
		return nil
	}
	if err := services.SyncAutostart(on && winapi.SelfElevated()); err != nil {
		log.Printf("rewinder: autostart: %v", err)
	}
	return nil
}
