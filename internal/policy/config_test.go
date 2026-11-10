package policy

import "testing"

func TestDefaultRules(t *testing.T) {
	r := DefaultConfig().Rules
	cases := []struct {
		exe  string
		want bool
	}{
		{`C:\Windows\notepad.exe`, true},
		{`C:\Users\me\AppData\Local\Temp\setup123\installer.exe`, false},
		{`C:\Windows\System32\ApplicationFrameHost.exe`, false},
		{`C:\Program Files\KeePassXC\KeePassXC.exe`, false},
		{``, false},
	}
	for _, c := range cases {
		if got := r.Allow(c.exe, ""); got != c.want {
			t.Errorf("Allow(%q) = %v, want %v", c.exe, got, c.want)
		}
	}
}

func TestExcludedWindowClass(t *testing.T) {
	r := Rules{ExcludeWindowClasses: []string{"SecretWnd"}}
	if r.Allow(`C:\a.exe`, "secretwnd") {
		t.Fatal("class match must be case-insensitive")
	}
}
