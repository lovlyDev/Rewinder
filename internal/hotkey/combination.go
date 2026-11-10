package hotkey

import (
	"errors"
	"fmt"
	"strings"
)

type Combination struct {
	Mods uint32 `json:"mods"`
	Key  uint32 `json:"key"`
}

func Default() Combination { return Combination{Mods: 3, Key: 'Z'} }

func (c Combination) Validate() error {
	if c.Mods == 0 || c.Mods & ^uint32(15) != 0 {
		return errors.New("hotkey.modifiers")
	}
	if !(c.Key >= 'A' && c.Key <= 'Z' || c.Key >= '0' && c.Key <= '9' || c.Key >= 0x70 && c.Key <= 0x87) {
		return errors.New("hotkey.key")
	}
	return nil
}

func (c Combination) String() string {
	parts := []string{}
	for _, m := range []struct {
		bit  uint32
		name string
	}{{2, "Ctrl"}, {1, "Alt"}, {4, "Shift"}, {8, "Win"}} {
		if c.Mods&m.bit != 0 {
			parts = append(parts, m.name)
		}
	}
	key := string(rune(c.Key))
	if c.Key >= 0x70 && c.Key <= 0x87 {
		key = fmt.Sprintf("F%d", c.Key-0x6f)
	}
	return strings.Join(append(parts, key), " + ")
}

func Parse(s string) (Combination, error) {
	var c Combination
	for _, part := range strings.Split(strings.ToUpper(s), "+") {
		p := strings.TrimSpace(part)
		var bit uint32
		switch p {
		case "CTRL", "CONTROL":
			bit = 2
		case "ALT":
			bit = 1
		case "SHIFT":
			bit = 4
		case "WIN", "META":
			bit = 8
		}
		if bit != 0 {
			if c.Mods&bit != 0 {
				return c, errors.New("hotkey.duplicate")
			}
			c.Mods |= bit
			continue
		}
		if c.Key != 0 {
			return c, errors.New("hotkey.key")
		}
		if len(p) == 1 {
			c.Key = uint32(p[0])
		} else {
			var n int
			if _, err := fmt.Sscanf(p, "F%d", &n); err != nil || n < 1 || n > 24 || p != fmt.Sprintf("F%d", n) {
				return c, errors.New("hotkey.key")
			}
			c.Key = uint32(0x6f + n)
		}
	}
	return c, c.Validate()
}
