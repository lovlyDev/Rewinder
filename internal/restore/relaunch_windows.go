//go:build windows

package restore

import (
	"Rewinder/internal/state"
	"Rewinder/internal/winapi"
)

// relaunch starts a closed app: a plugin first (it knows the app's own restore
// flags), then the Store launcher for packaged apps, then the saved command line.
func (e *Engine) relaunch(saved *state.AppState) error {
	if handled, err := e.plugins.Relaunch(saved); handled {
		return err
	}
	if saved.AppUserModelID != "" {
		return winapi.LaunchPackaged(saved.AppUserModelID)
	}
	if _, err := winapi.Launch(saved.ExecutablePath, saved.CommandLine, saved.WorkingDir); err != nil {
		// e.g. ERROR_ELEVATION_REQUIRED: let the shell handle it
		return winapi.ShellOpen(saved.ExecutablePath)
	}
	return nil
}
