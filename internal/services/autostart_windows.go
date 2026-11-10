//go:build windows

package services

import (
	"errors"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

const registryKey = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`
const appName = "Rewinder"
const autostartFlag = "--autostart"

// StartedByAutostart reports whether Windows launched us at sign-in; then the
// window stays hidden and Rewinder lives in the tray.
func StartedByAutostart(args []string) bool {
	for _, a := range args {
		if a == autostartFlag {
			return true
		}
	}
	return false
}

func IsAutostartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	_, _, err = key.GetStringValue(appName)
	return err == nil
}

func SetAutostart(enabled bool) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	if !enabled {
		if err := key.DeleteValue(appName); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return err
		}
		return nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	return key.SetStringValue(appName, `"`+filepath.Clean(exePath)+`" `+autostartFlag)
}
