//go:build windows

package plugins

import "Rewinder/internal/winapi"

// DefaultRegistry wires the plugins to the real system.
func DefaultRegistry() *Registry {
	env := Env{
		Open: winapi.ShellOpen,
		Launch: func(exePath, cmdLine, workingDir string) error {
			_, err := winapi.Launch(exePath, cmdLine, workingDir)
			return err
		},
		ExplorerFolder: winapi.ExplorerFolders,
	}
	return NewRegistry(env, NewExplorerPlugin(), NewChromePlugin(), NewOfficePlugin())
}
