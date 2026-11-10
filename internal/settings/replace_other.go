//go:build !windows

package settings

import "os"

func replace(from, to string) error { return os.Rename(from, to) }
