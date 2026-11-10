package update

import "testing"

func TestNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"v0.3.0", "0.2.0", true},
		{"0.2.0", "v0.3.0", false},
		{"1.2.10", "1.2.9", true},
		{"1.2", "1.2.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.0-beta.2", true},
		{"1.0.0-beta.2", "1.0.0", false},
		{"1.0.0-beta.2", "1.0.0-beta.1", true},
		{"v0.3.0", "dev", true},
		{"dev", "0.1.0", false},
		{"garbage", "dev", false},
	}
	for _, c := range cases {
		if got := Newer(c.a, c.b); got != c.want {
			t.Errorf("Newer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestChecksumFor(t *testing.T) {
	sums := "abc123  Rewinder.exe\nfff  *Rewinder-installer.exe\n\n"
	if got := checksumFor(sums, "Rewinder.exe"); got != "abc123" {
		t.Fatalf("got %q", got)
	}
	if got := checksumFor(sums, "Rewinder-installer.exe"); got != "fff" {
		t.Fatalf("binary-mode marker: got %q", got)
	}
	if got := checksumFor(sums, "missing.exe"); got != "" {
		t.Fatalf("got %q", got)
	}
}
