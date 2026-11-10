package main

import (
	"encoding/base64"
	"strings"
	"sync"

	"Rewinder/internal/winapi"
)

const appIconSize = 64

var (
	iconMu    sync.Mutex
	iconCache = map[string]string{}
)

// GetAppIcon returns the icon of a tracked application as a PNG data URL, or ""
// when it has none. Only executables Rewinder already tracks are read.
func (a *App) GetAppIcon(appID string) string {
	if a.svc == nil {
		return ""
	}
	exe := ""
	for _, app := range a.svc.GetApps() {
		if app.AppID == appID {
			exe = app.ExecutablePath
			break
		}
	}
	if exe == "" {
		return ""
	}
	key := strings.ToLower(exe)
	iconMu.Lock()
	cached, ok := iconCache[key]
	iconMu.Unlock()
	if ok {
		return cached
	}
	url := ""
	if data, err := winapi.AppIconPNG(exe, appIconSize); err == nil {
		url = "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)
	}
	iconMu.Lock()
	iconCache[key] = url
	iconMu.Unlock()
	return url
}
