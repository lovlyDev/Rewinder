//go:build windows

package plugins

import (
	"os"
	"path/filepath"
	"strings"

	"Rewinder/internal/state"
)

type ExplorerPlugin struct{}

func NewExplorerPlugin() *ExplorerPlugin { return &ExplorerPlugin{} }

func (p *ExplorerPlugin) ID() string { return "explorer" }

func (p *ExplorerPlugin) CanHandle(app *state.AppState) bool {
	base := strings.ToLower(filepath.Base(app.ExecutablePath))
	return base == "explorer.exe"
}

func (p *ExplorerPlugin) Capture(app *state.AppState) error {
	var folders []string
	seen := map[string]struct{}{}
	for _, f := range app.OpenFiles {
		if len(folders) >= 3 {
			break
		}
		if f.Path == "" {
			continue
		}
		folder := f.Path
		if info, err := os.Stat(folder); err == nil && !info.IsDir() {
			folder = filepath.Dir(folder)
		}
		if folder == "" {
			continue
		}
		low := strings.ToLower(folder)
		if _, ok := seen[low]; ok {
			continue
		}
		seen[low] = struct{}{}
		folders = append(folders, folder)
	}
	if len(folders) > 0 {
		app.PluginData[p.ID()] = map[string]any{"folders": folders}
	}
	return nil
}

func (p *ExplorerPlugin) Restore(app *state.AppState) error {
	v, ok := app.PluginData[p.ID()].(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := v["folders"].([]string)
	if !ok || len(raw) == 0 {
		return nil
	}
	for _, folder := range raw {
		_ = shellOpen(folder)
	}
	return nil
}
