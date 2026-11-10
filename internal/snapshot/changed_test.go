package snapshot

import (
	"testing"

	"Rewinder/internal/state"
)

func TestDiffFileCounts(t *testing.T) {
	f := func(paths ...string) []state.FileRef {
		out := make([]state.FileRef, len(paths))
		for i, p := range paths {
			out[i] = state.FileRef{Path: p}
		}
		return out
	}
	cases := []struct {
		prev, next     []state.FileRef
		added, removed int
	}{
		{nil, nil, 0, 0},
		{nil, f("a"), 1, 0},
		{f("a"), nil, 0, 1},
		{f(`C:\A.docx`, "b"), f(`c:\a.docx`, "c"), 1, 1},
		{f("a", "a"), f("a"), 0, 0},
	}
	for i, c := range cases {
		a, r := diffFileCounts(c.prev, c.next)
		if a != c.added || r != c.removed {
			t.Errorf("case %d: got +%d -%d, want +%d -%d", i, a, r, c.added, c.removed)
		}
	}
}

func TestPluginEqTreatsNilAndEmptyAlike(t *testing.T) {
	a := map[string]state.PluginState{"explorer": {Paths: nil, Values: nil}}
	b := map[string]state.PluginState{"explorer": {Paths: []string{}, Values: map[string]string{}}}
	if !pluginEq(a, b) || !pluginEq(nil, map[string]state.PluginState{}) {
		t.Fatal("nil and empty plugin data must compare equal")
	}
	c := map[string]state.PluginState{"explorer": {Paths: []string{`C:\x`}}}
	if pluginEq(a, c) {
		t.Fatal("different paths must not compare equal")
	}
	d := map[string]state.PluginState{"chrome": {Values: map[string]string{"profile": "1"}}}
	e := map[string]state.PluginState{"chrome": {Values: map[string]string{"profile": "2"}}}
	if pluginEq(d, e) {
		t.Fatal("different values must not compare equal")
	}
}
