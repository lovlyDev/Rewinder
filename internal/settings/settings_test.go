package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsAndRoundtrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Load(dir)
	if err != nil || s != Default() {
		t.Fatal(s, err)
	}
	if s.Theme.Radius != 14 || s.Hotkey.Key != 'Z' {
		t.Fatal(s)
	}
	for _, radius := range []int{0, 24, 14} {
		s.Theme.Radius = radius
		if err = Save(dir, s); err != nil {
			t.Fatal(err)
		}
		got, err := Load(dir)
		if err != nil || got != s {
			t.Fatal(got, err)
		}
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatal(entries)
	}
}
func TestValidationPreservesFile(t *testing.T) {
	dir := t.TempDir()
	original := Default()
	if err := Save(dir, original); err != nil {
		t.Fatal(err)
	}
	for _, modify := range []func(*Settings){func(s *Settings) { s.Theme.Radius = 25 }, func(s *Settings) { s.Language = "xx" }, func(s *Settings) { s.Theme.Accent = "red" }, func(s *Settings) { s.Hotkey.Mods = 0 }, func(s *Settings) { s.Theme.Background.Image = "../bad" }, func(s *Settings) { s.Theme.Background.Dim = 2 }, func(s *Settings) { s.Theme.Density = "bad" }, func(s *Settings) { s.Theme.Motion = "bad" }} {
		s := original
		modify(&s)
		if err := Save(dir, s); err == nil {
			t.Fatal("accepted", s)
		}
		got, err := Load(dir)
		if err != nil || got != original {
			t.Fatal(got, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"theme":{"radius":99}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if got, err := Load(dir); err == nil || got != Default() {
		t.Fatal(got, err)
	}
}
