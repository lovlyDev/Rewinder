package hotkey

import "testing"

func TestParse(t *testing.T) {
	for _, s := range []string{"Ctrl + Alt + Z", "Shift+Ctrl+F12", "Win + 1", "Alt+F24"} {
		c, err := Parse(s)
		if err != nil {
			t.Fatal(s, err)
		}
		again, err := Parse(c.String())
		if err != nil || again != c {
			t.Fatal(s, c, again, err)
		}
	}
	for _, s := range []string{"Z", "Ctrl+", "Ctrl+Alt", "Ctrl+Ctrl+Z", "Ctrl+A+B", "Alt+F25", "Ctrl+F1junk"} {
		if _, err := Parse(s); err == nil {
			t.Fatal("accepted", s)
		}
	}
	if c, err := Parse("ctrl+alt+z"); err != nil || c != Default() {
		t.Fatal(c, err)
	}
}
