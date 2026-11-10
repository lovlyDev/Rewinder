//go:build windows

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"Rewinder/internal/winapi"
)

// version is set at release builds: -ldflags "-X main.version=1.2.3".
var version = "dev"

const waitPidFlag = "--wait-pid="

// restartElevated starts an elevated copy of Rewinder that waits for this one
// to exit (so the single-instance lock is free) and keeps the window hidden
// when this instance was started by autostart.
func restartElevated() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := []string{fmt.Sprintf("%s%d", waitPidFlag, os.Getpid())}
	for _, a := range os.Args[1:] {
		if !strings.HasPrefix(a, waitPidFlag) {
			args = append(args, a)
		}
	}
	return winapi.RunElevated(exe, strings.Join(args, " "))
}

// waitForPreviousInstance handles --wait-pid=N from restartElevated.
func waitForPreviousInstance(args []string) {
	for _, a := range args {
		if pid, err := strconv.ParseUint(strings.TrimPrefix(a, waitPidFlag), 10, 32); strings.HasPrefix(a, waitPidFlag) && err == nil {
			winapi.WaitForExit(uint32(pid), 10_000)
		}
	}
}
