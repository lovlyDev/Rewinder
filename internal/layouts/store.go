// Package layouts keeps named window layouts: a saved arrangement of several
// applications ("Work", "Stream") that can be put back in one go.
package layouts

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"Rewinder/internal/state"

	"github.com/google/uuid"
)

const (
	MaxLayouts = 30
	maxNameLen = 60
	fileName   = "layouts.json"
)

type Layout struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	CreatedAt time.Time        `json:"createdAt"`
	Apps      []state.AppState `json:"apps"`
}

// Store is safe for concurrent use. Layouts keep their order: the first nine
// get the Ctrl+Alt+1…9 shortcuts.
type Store struct {
	path string
	mu   sync.RWMutex
	list []Layout
}

// Load reads dir/layouts.json. A missing file is an empty store; a corrupt one
// is kept aside as layouts.json.bad so nothing is silently overwritten.
func Load(dir string) *Store {
	s := &Store{path: filepath.Join(dir, fileName)}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return s
	}
	if json.Unmarshal(data, &s.list) != nil {
		_ = os.Rename(s.path, s.path+".bad")
		s.list = nil
	}
	return s
}

func CleanName(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" {
		return "", errors.New("layout.name")
	}
	if utf8.RuneCountInString(name) > maxNameLen {
		name = string([]rune(name)[:maxNameLen])
	}
	return name, nil
}

func (s *Store) List() []Layout {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]Layout(nil), s.list...)
}

func (s *Store) Get(id string) (Layout, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, l := range s.list {
		if l.ID == id {
			return l, true
		}
	}
	return Layout{}, false
}

// Add stores a new layout; a layout with the same name is replaced in place.
func (s *Store) Add(name string, apps []state.AppState, now time.Time) (Layout, error) {
	name, err := CleanName(name)
	if err != nil {
		return Layout{}, err
	}
	if len(apps) == 0 {
		return Layout{}, errors.New("layout.empty")
	}
	l := Layout{ID: uuid.NewString(), Name: name, CreatedAt: now, Apps: apps}
	s.mu.Lock()
	defer s.mu.Unlock()
	next := append([]Layout(nil), s.list...)
	replaced := false
	for i := range next {
		if strings.EqualFold(next[i].Name, name) {
			l.ID = next[i].ID
			next[i] = l
			replaced = true
			break
		}
	}
	if !replaced {
		if len(next) >= MaxLayouts {
			return Layout{}, errors.New("layout.full")
		}
		next = append(next, l)
	}
	if err := s.writeLocked(next); err != nil {
		return Layout{}, err
	}
	return l, nil
}

func (s *Store) Rename(id, name string) error {
	name, err := CleanName(name)
	if err != nil {
		return err
	}
	return s.change(id, func(list []Layout, i int) []Layout {
		list[i].Name = name
		return list
	})
}

func (s *Store) Delete(id string) error {
	return s.change(id, func(list []Layout, i int) []Layout {
		return append(list[:i], list[i+1:]...)
	})
}

// Move puts a layout at a new position (which decides its Ctrl+Alt+N shortcut).
func (s *Store) Move(id string, to int) error {
	return s.change(id, func(list []Layout, i int) []Layout {
		l := list[i]
		list = append(list[:i], list[i+1:]...)
		if to < 0 {
			to = 0
		}
		if to > len(list) {
			to = len(list)
		}
		return append(list[:to], append([]Layout{l}, list[to:]...)...)
	})
}

func (s *Store) change(id string, f func([]Layout, int) []Layout) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := append([]Layout(nil), s.list...)
	for i := range next {
		if next[i].ID == id {
			return s.writeLocked(f(next, i))
		}
	}
	return errors.New("layout.missing")
}

func (s *Store) writeLocked(next []Layout) error {
	data, err := json.Marshal(next)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return errors.New("settings.write")
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return errors.New("settings.write")
	}
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp) // some filters refuse renames: write in place
		if err := os.WriteFile(s.path, data, 0o600); err != nil {
			return errors.New("settings.write")
		}
	}
	s.list = next
	return nil
}
