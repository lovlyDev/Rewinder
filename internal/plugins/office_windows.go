package plugins

import (
	"path/filepath"
	"strings"

	"Rewinder/internal/state"
)

type OfficePlugin struct{}

func NewOfficePlugin() *OfficePlugin { return &OfficePlugin{} }

func (p *OfficePlugin) ID() string { return "office" }

func (p *OfficePlugin) CanHandle(app *state.AppState) bool {
	base := strings.ToLower(filepath.Base(app.ExecutablePath))
	return base == "winword.exe" || base == "excel.exe" || base == "powerpnt.exe"
}

func (p *OfficePlugin) Capture(app *state.AppState) error {
	var docs []string
	for _, f := range app.OpenFiles {
		low := strings.ToLower(f.Path)
		if strings.HasSuffix(low, ".docx") || strings.HasSuffix(low, ".doc") ||
			strings.HasSuffix(low, ".xlsx") || strings.HasSuffix(low, ".xls") ||
			strings.HasSuffix(low, ".pptx") || strings.HasSuffix(low, ".ppt") {
			docs = append(docs, f.Path)
			if len(docs) >= 5 {
				break
			}
		}
	}
	if len(docs) > 0 {
		app.PluginData[p.ID()] = map[string]any{"docs": docs}
	}
	return nil
}

func (p *OfficePlugin) Restore(app *state.AppState) error {
	v, ok := app.PluginData[p.ID()].(map[string]any)
	if !ok {
		return nil
	}
	docs, ok := v["docs"].([]string)
	if !ok {
		return nil
	}
	for _, d := range docs {
		_ = shellOpen(d)
	}
	return nil
}
