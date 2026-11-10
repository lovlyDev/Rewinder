// Package settings owns validated, atomically persisted user preferences.
package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"Rewinder/internal/hotkey"
)

type Background struct {
	Kind     string  `json:"kind"`
	Color    string  `json:"color"`
	EndColor string  `json:"endColor"`
	Image    string  `json:"image"`
	Dim      float64 `json:"dim"`
	Blur     int     `json:"blur"`
}
type Theme struct {
	Accent     string     `json:"accent"`
	Radius     int        `json:"radius"`
	Density    string     `json:"density"`
	Background Background `json:"background"`
	Motion     string     `json:"motion"`
}
type Settings struct {
	Hotkey   hotkey.Combination `json:"hotkey"`
	Language string             `json:"language"`
	Theme    Theme              `json:"theme"`
	// RunAsAdmin restarts Rewinder elevated, so it can move the windows of
	// applications that run as administrator.
	RunAsAdmin bool `json:"runAsAdmin"`
	// Updates: "auto" checks GitHub for a new version and offers it, "install"
	// also installs it on its own (restarting quietly into the tray), "off" never checks.
	Updates string `json:"updates"`
}

func Default() Settings {
	return Settings{Hotkey: hotkey.Default(), Language: "system", Updates: "install", Theme: Theme{Accent: "#a99bff", Radius: 14, Density: "normal", Motion: "normal", Background: Background{Kind: "color", Color: "#0a0a0c", EndColor: "#171126", Dim: 0.65}}}
}

var color = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (s Settings) Validate() error {
	if err := s.Hotkey.Validate(); err != nil {
		return err
	}
	b := s.Theme.Background
	if s.Language != "system" && s.Language != "ru" && s.Language != "en" {
		return errors.New("settings.language")
	}
	if !color.MatchString(s.Theme.Accent) || !color.MatchString(b.Color) || !color.MatchString(b.EndColor) {
		return errors.New("settings.color")
	}
	if s.Theme.Radius < 0 || s.Theme.Radius > 24 || b.Blur < 0 || b.Blur > 32 || !(b.Dim >= 0 && b.Dim <= 1) {
		return errors.New("settings.range")
	}
	if s.Theme.Density != "normal" && s.Theme.Density != "compact" {
		return errors.New("settings.density")
	}
	if s.Theme.Motion != "normal" && s.Theme.Motion != "fast" && s.Theme.Motion != "off" {
		return errors.New("settings.motion")
	}
	if b.Kind != "color" && b.Kind != "gradient" && b.Kind != "image" {
		return errors.New("settings.background")
	}
	if s.Updates != "" && s.Updates != "auto" && s.Updates != "install" && s.Updates != "off" {
		return errors.New("settings.updates")
	}
	if b.Image != "" && b.Image != "background.png" {
		return errors.New("settings.image")
	}
	return nil
}
func Directory() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base, _ = os.UserConfigDir()
	}
	return filepath.Join(base, "Rewinder")
}
func Load(dir string) (Settings, error) {
	s := Default()
	data, err := os.ReadFile(filepath.Join(dir, "settings.json"))
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(data, &s); err != nil {
		return Default(), err
	}
	if err = s.Validate(); err != nil {
		return Default(), err
	}
	return s, nil
}
func Save(dir string, s Settings) error {
	if err := s.Validate(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, "settings-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	target := filepath.Join(dir, "settings.json")
	if err = replace(f.Name(), target); err != nil {
		// Some file system filters (sandboxes, sync clients, antivirus) refuse the
		// rename. A direct write is still fine: Load falls back to defaults on a torn file.
		if werr := os.WriteFile(target, data, 0o600); werr != nil {
			return fmt.Errorf("replace: %v; direct write: %w", err, werr)
		}
	}
	return nil
}
