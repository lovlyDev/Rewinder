package restore

import (
	"sort"

	"Rewinder/internal/state"
)

// MatchWindows pairs saved windows with windows that exist now. It returns
// saved index -> live index. Each live window is used at most once.
//
// Passes, strongest first:
//  1. the same HWND (the app never closed)
//  2. the same document/folder path
//  3. the same class and title
//  4. the same class (titles change: browser tabs, unsaved markers)
//
// Within a pass saved windows are taken topmost first, so when several windows
// look alike the stacking order is preserved.
func MatchWindows(saved, live []state.WindowState) map[int]int {
	out := map[int]int{}
	used := map[int]bool{}

	order := make([]int, len(saved))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return saved[order[a]].ZOrder < saved[order[b]].ZOrder })

	liveOrder := make([]int, len(live))
	for i := range liveOrder {
		liveOrder[i] = i
	}
	sort.SliceStable(liveOrder, func(a, b int) bool { return live[liveOrder[a]].ZOrder < live[liveOrder[b]].ZOrder })

	pass := func(eq func(s, l *state.WindowState) bool) {
		for _, si := range order {
			if _, done := out[si]; done {
				continue
			}
			for _, li := range liveOrder {
				if used[li] || !eq(&saved[si], &live[li]) {
					continue
				}
				out[si] = li
				used[li] = true
				break
			}
		}
	}

	pass(func(s, l *state.WindowState) bool { return s.HWND != 0 && s.HWND == l.HWND })
	pass(func(s, l *state.WindowState) bool { return s.Path != "" && equalFold(s.Path, l.Path) })
	pass(func(s, l *state.WindowState) bool { return s.ClassName == l.ClassName && s.Title == l.Title })
	pass(func(s, l *state.WindowState) bool { return s.ClassName == l.ClassName })
	return out
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
