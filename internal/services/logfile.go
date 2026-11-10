package services

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

const maxLogBytes = 1 << 20

// SetupLogFile sends the standard logger to %LOCALAPPDATA%\Rewinder\rewinder.log
// (a GUI app has no console). The file is started over once it passes 1 MB.
func SetupLogFile() string {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return ""
	}
	dir := filepath.Join(base, "Rewinder")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	path := filepath.Join(dir, "rewinder.log")
	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if st, err := os.Stat(path); err == nil && st.Size() > maxLogBytes {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	f, err := os.OpenFile(path, flags, 0o644)
	if err != nil {
		return ""
	}
	log.SetOutput(io.MultiWriter(f, os.Stderr))
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	return path
}
