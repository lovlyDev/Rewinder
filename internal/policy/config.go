package policy

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Retention          time.Duration
	StorageDir         string
	MaxSnapshotsPerApp int
	Rules              Rules
}

type Rules struct {
	ExcludeExeNames      []string
	ExcludePathSubstr    []string
	ExcludeWindowClasses []string
}

func (r Rules) Allow(exePath string, windowClass string) bool {
	if exePath == "" {
		return false
	}
	exeLower := strings.ToLower(filepath.Base(exePath))
	pathLower := strings.ToLower(exePath)
	classLower := strings.ToLower(windowClass)

	for _, n := range r.ExcludeExeNames {
		if exeLower == strings.ToLower(n) {
			return false
		}
	}
	for _, s := range r.ExcludePathSubstr {
		if s != "" && strings.Contains(pathLower, strings.ToLower(s)) {
			return false
		}
	}
	for _, c := range r.ExcludeWindowClasses {
		if c != "" && classLower == strings.ToLower(c) {
			return false
		}
	}
	return true
}

// Shell and system processes whose windows are not user apps, or that host other
// apps' windows (ApplicationFrameHost owns UWP frames, so relaunching it is useless).
var systemExes = []string{
	"applicationframehost.exe",
	"searchhost.exe",
	"searchapp.exe",
	"startmenuexperiencehost.exe",
	"shellexperiencehost.exe",
	"shellhost.exe",
	"textinputhost.exe",
	"lockapp.exe",
	"systemsettings.exe",
	"taskmgr.exe",
	"rewinder.exe",
	// password managers: never keep a record of their windows
	"keepass.exe",
	"keepassxc.exe",
	"1password.exe",
	"bitwarden.exe",
}

func DefaultConfig() *Config {
	return &Config{
		Retention:          7 * 24 * time.Hour,
		StorageDir:         defaultStorageDir(),
		MaxSnapshotsPerApp: 500,
		Rules: Rules{
			ExcludeExeNames:   append([]string(nil), systemExes...),
			ExcludePathSubstr: []string{`\AppData\Local\Temp\`},
		},
	}
}

func defaultStorageDir() string {
	if v := os.Getenv("LOCALAPPDATA"); v != "" {
		return filepath.Join(v, "Rewinder", "history")
	}
	if v := os.Getenv("APPDATA"); v != "" {
		return filepath.Join(v, "Rewinder", "history")
	}
	return filepath.Join(".", "history")
}
