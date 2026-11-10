//go:build windows

package services

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

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

// Autostart has two forms. Normally a value under HKCU\...\Run. With "run as
// administrator" a Run value cannot elevate, so a Task Scheduler task with the
// highest privileges starts Rewinder at sign-in instead (creating or deleting it
// needs administrator rights, which Rewinder has in that mode).

func IsAutostartEnabled() bool { return runValueExists() || taskExists() }

// SetAutostart turns autostart on or off; elevated chooses the task form.
func SetAutostart(enabled, elevated bool) error {
	if !enabled {
		err := deleteRunValue()
		if taskExists() {
			if terr := schtasks("/Delete", "/TN", appName, "/F"); terr != nil {
				return errors.New("autostart.admin")
			}
		}
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe = filepath.Clean(exe)
	if elevated {
		if err := schtasks("/Create", "/TN", appName, "/TR", `"`+exe+`" `+autostartFlag, "/SC", "ONLOGON", "/RL", "HIGHEST", "/F"); err != nil {
			return errors.New("autostart.admin")
		}
		return deleteRunValue()
	}
	if taskExists() {
		_ = schtasks("/Delete", "/TN", appName, "/F") // best effort: needs admin
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(appName, `"`+exe+`" `+autostartFlag)
}

// SyncAutostart moves an enabled autostart to the form that matches the
// current mode, e.g. after "run as administrator" was switched on.
func SyncAutostart(elevated bool) error {
	if !IsAutostartEnabled() || (elevated && taskExists() && !runValueExists()) || (!elevated && !taskExists()) {
		return nil
	}
	return SetAutostart(true, elevated)
}

func runValueExists() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	_, _, err = key.GetStringValue(appName)
	return err == nil
}

func deleteRunValue() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	if err := key.DeleteValue(appName); err != nil && !errors.Is(err, registry.ErrNotExist) {
		return err
	}
	return nil
}

func taskExists() bool { return schtasks("/Query", "/TN", appName) == nil }

func schtasks(args ...string) error {
	cmd := exec.Command("schtasks.exe", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000} // CREATE_NO_WINDOW
	return cmd.Run()
}
