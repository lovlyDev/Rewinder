package snapshot

import (
	"strings"

	"Rewinder/internal/state"
)

// Changed reports whether next differs from prev in a way worth a new snapshot.
// Focus, Z-order and window titles change constantly and alone do not count.
func Changed(prev, next *state.AppState) bool {
	if prev.PID != next.PID || len(prev.Windows) != len(next.Windows) {
		return true
	}
	pw := make(map[uintptr]*state.WindowState, len(prev.Windows))
	for i := range prev.Windows {
		pw[prev.Windows[i].HWND] = &prev.Windows[i]
	}
	for i := range next.Windows {
		nw := &next.Windows[i]
		ow, ok := pw[nw.HWND]
		if !ok || !sameGeometry(ow, nw) || ow.Path != nw.Path {
			return true
		}
	}
	if added, removed := diffFileCounts(prev.OpenFiles, next.OpenFiles); added > 0 || removed > 0 {
		return true
	}
	return !pluginEq(prev.PluginData, next.PluginData)
}

func sameGeometry(a, b *state.WindowState) bool {
	return a.NormalRect == b.NormalRect &&
		a.Rect == b.Rect &&
		a.ShowCmd == b.ShowCmd &&
		a.MonitorID == b.MonitorID
}

// diffFileCounts counts files opened and closed between two states (case-insensitive paths).
func diffFileCounts(prev, next []state.FileRef) (added, removed int) {
	if len(prev) == 0 || len(next) == 0 {
		return len(next), len(prev)
	}
	seen := make(map[string]bool, len(prev)) // path -> still open in next
	for _, f := range prev {
		seen[strings.ToLower(f.Path)] = false
	}
	for _, f := range next {
		k := strings.ToLower(f.Path)
		if _, ok := seen[k]; ok {
			seen[k] = true
		} else {
			added++
		}
	}
	for _, open := range seen {
		if !open {
			removed++
		}
	}
	return added, removed
}

// pluginEq compares plugin data, treating nil and empty collections as equal
// (history loaded from disk and a fresh capture differ in exactly that way).
func pluginEq(a, b map[string]state.PluginState) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok || len(av.Paths) != len(bv.Paths) || len(av.Values) != len(bv.Values) {
			return false
		}
		for i := range av.Paths {
			if av.Paths[i] != bv.Paths[i] {
				return false
			}
		}
		for vk, vv := range av.Values {
			if bv.Values[vk] != vv {
				return false
			}
		}
	}
	return true
}
