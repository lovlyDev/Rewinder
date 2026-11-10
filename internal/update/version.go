// Package update checks GitHub Releases for a newer Rewinder, downloads it,
// verifies it and swaps the running executable.
package update

import (
	"strconv"
	"strings"
)

// Newer reports whether version a is newer than b ("v1.2.10" > "1.2.9").
// Pre-release suffixes ("-beta.1") sort before the release itself. A version
// that does not parse (e.g. "dev") is never newer and everything is newer than it.
func Newer(a, b string) bool {
	pa, oka := parse(a)
	pb, okb := parse(b)
	if !oka {
		return false
	}
	if !okb {
		return true
	}
	for i := 0; i < 3; i++ {
		if pa.nums[i] != pb.nums[i] {
			return pa.nums[i] > pb.nums[i]
		}
	}
	switch {
	case pa.pre == pb.pre:
		return false
	case pa.pre == "":
		return true // 1.2.0 > 1.2.0-beta
	case pb.pre == "":
		return false
	default:
		return pa.pre > pb.pre
	}
}

type semver struct {
	nums [3]int
	pre  string
}

func parse(v string) (semver, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	var out semver
	if i := strings.IndexByte(v, '-'); i >= 0 {
		v, out.pre = v[:i], v[i+1:]
	}
	if i := strings.IndexByte(v, '+'); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out.nums[i] = n
	}
	return out, true
}
