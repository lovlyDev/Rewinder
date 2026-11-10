// Package plugins adds per-application knowledge on top of generic window capture:
// what to remember (open folders, documents, browser profile) and how to bring an
// app back when it is no longer running.
package plugins

import (
	"path/filepath"
	"strings"

	"Rewinder/internal/state"
)

// Env carries the side effects plugins may perform, so they can be faked in tests.
type Env struct {
	Open           func(target string) error                       // open a file/folder with its default handler
	Launch         func(exePath, cmdLine, workingDir string) error // start a process
	ExplorerFolder func() map[uintptr]string                       // HWND -> folder of open Explorer windows
}

type AppPlugin interface {
	ID() string
	CanHandle(exePath string) bool
	NeedsOpenFiles() bool
	// Capture records plugin data into app.PluginData[ID()] and may set WindowState.Path.
	Capture(env Env, app *state.AppState)
	// Relaunch starts the app when no instance is running. handled=false means the
	// restore engine should start it the generic way.
	Relaunch(env Env, saved *state.AppState) (handled bool, err error)
	// Reopen opens what saved had and the running instance (current) lacks.
	// It returns how many items it asked the system to open.
	Reopen(env Env, saved, current *state.AppState) int
}

type Registry struct {
	env  Env
	list []AppPlugin
}

func NewRegistry(env Env, list ...AppPlugin) *Registry {
	return &Registry{env: env, list: list}
}

func (r *Registry) find(exePath string) AppPlugin {
	for _, p := range r.list {
		if p.CanHandle(exePath) {
			return p
		}
	}
	return nil
}

func (r *Registry) NeedsOpenFiles(exePath string) bool {
	p := r.find(exePath)
	return p != nil && p.NeedsOpenFiles()
}

func (r *Registry) Capture(app *state.AppState) {
	if app.PluginData == nil {
		app.PluginData = map[string]state.PluginState{}
	}
	if p := r.find(app.ExecutablePath); p != nil {
		p.Capture(r.env, app)
	}
}

func (r *Registry) Relaunch(saved *state.AppState) (bool, error) {
	if p := r.find(saved.ExecutablePath); p != nil {
		return p.Relaunch(r.env, saved)
	}
	return false, nil
}

func (r *Registry) Reopen(saved, current *state.AppState) int {
	if p := r.find(saved.ExecutablePath); p != nil {
		return p.Reopen(r.env, saved, current)
	}
	return 0
}

func exeBase(exePath string) string {
	return strings.ToLower(filepath.Base(exePath))
}

// missing returns the items of want that are not in have (case-insensitive).
func missing(want, have []string) []string {
	h := map[string]bool{}
	for _, p := range have {
		h[strings.ToLower(filepath.Clean(p))] = true
	}
	var out []string
	for _, p := range want {
		if !h[strings.ToLower(filepath.Clean(p))] {
			out = append(out, p)
		}
	}
	return out
}

func openAll(env Env, paths []string) int {
	n := 0
	for _, p := range paths {
		if env.Open != nil && env.Open(p) == nil {
			n++
		}
	}
	return n
}

func pathsOf(app *state.AppState, id string) []string {
	if app == nil || app.PluginData == nil {
		return nil
	}
	return app.PluginData[id].Paths
}
