package plugins

import (
	"strings"

	"Rewinder/internal/state"
)

// ExplorerPlugin remembers which folders File Explorer windows show.
type ExplorerPlugin struct{}

func NewExplorerPlugin() *ExplorerPlugin { return &ExplorerPlugin{} }

func (p *ExplorerPlugin) ID() string { return "explorer" }

func (p *ExplorerPlugin) CanHandle(exePath string) bool { return exeBase(exePath) == "explorer.exe" }

func (p *ExplorerPlugin) NeedsOpenFiles() bool { return false }

func (p *ExplorerPlugin) Capture(env Env, app *state.AppState) {
	if env.ExplorerFolder == nil {
		return
	}
	folders := env.ExplorerFolder()
	seen := map[string]bool{}
	var paths []string
	for i := range app.Windows {
		f := folders[app.Windows[i].HWND]
		if f == "" {
			continue
		}
		app.Windows[i].Path = f
		if k := strings.ToLower(f); !seen[k] {
			seen[k] = true
			paths = append(paths, f)
		}
	}
	app.PluginData[p.ID()] = state.PluginState{Paths: paths}
}

// Explorer is the shell and is always running, so "relaunching" means opening the folders.
func (p *ExplorerPlugin) Relaunch(env Env, saved *state.AppState) (bool, error) {
	openAll(env, pathsOf(saved, p.ID()))
	return true, nil
}

func (p *ExplorerPlugin) Reopen(env Env, saved, current *state.AppState) int {
	return openAll(env, missing(pathsOf(saved, p.ID()), pathsOf(current, p.ID())))
}
