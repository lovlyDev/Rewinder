package policy

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const maxExclusions = 500

// Exclusions are executables the user chose not to track, by file name
// ("chrome.exe"), stored in exclusions.json next to the settings.
type Exclusions struct {
	path string
	mu   sync.RWMutex
	set  map[string]bool
}

func normalize(exe string) string { return strings.ToLower(strings.TrimSpace(filepath.Base(exe))) }

// LoadExclusions reads dir/exclusions.json; a missing or unreadable file means none.
func LoadExclusions(dir string) *Exclusions {
	e := &Exclusions{path: filepath.Join(dir, "exclusions.json"), set: map[string]bool{}}
	data, err := os.ReadFile(e.path)
	if err != nil {
		return e
	}
	var names []string
	if json.Unmarshal(data, &names) == nil {
		for _, n := range names {
			if n = normalize(n); n != "" && n != "." {
				e.set[n] = true
			}
		}
	}
	return e
}

// Excluded reports whether the executable at exePath is excluded.
func (e *Exclusions) Excluded(exePath string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.set[normalize(exePath)]
}

// List returns the excluded file names, sorted.
func (e *Exclusions) List() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]string, 0, len(e.set))
	for n := range e.set {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Set adds or removes one executable and saves the list.
func (e *Exclusions) Set(exe string, excluded bool) error {
	n := normalize(exe)
	if n == "" || n == "." || !strings.HasSuffix(n, ".exe") {
		return errors.New("exclusions.name")
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if excluded == e.set[n] {
		return nil
	}
	if excluded && len(e.set) >= maxExclusions {
		return errors.New("exclusions.full")
	}
	if excluded {
		e.set[n] = true
	} else {
		delete(e.set, n)
	}
	if err := e.saveLocked(); err != nil {
		// keep memory and disk in agreement
		if excluded {
			delete(e.set, n)
		} else {
			e.set[n] = true
		}
		return err
	}
	return nil
}

func (e *Exclusions) saveLocked() error {
	names := make([]string, 0, len(e.set))
	for n := range e.set {
		names = append(names, n)
	}
	sort.Strings(names)
	data, _ := json.MarshalIndent(names, "", "  ")
	if err := os.MkdirAll(filepath.Dir(e.path), 0o700); err != nil {
		return errors.New("settings.write")
	}
	tmp := e.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return errors.New("settings.write")
	}
	if err := os.Rename(tmp, e.path); err != nil {
		// same fallback as the settings file: some filters refuse renames
		_ = os.Remove(tmp)
		if err := os.WriteFile(e.path, data, 0o600); err != nil {
			return errors.New("settings.write")
		}
	}
	return nil
}
