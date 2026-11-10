package plugins

import (
	"reflect"
	"testing"

	"Rewinder/internal/state"
)

func TestExtractFlagValue(t *testing.T) {
	cases := []struct{ cmd, flag, want string }{
		{`chrome.exe --user-data-dir="C:\My Profile" --foo`, "--user-data-dir", `C:\My Profile`},
		{`chrome.exe --user-data-dir=C:\P --foo`, "--user-data-dir", `C:\P`},
		{`chrome.exe --user-data-dir C:\P`, "--user-data-dir", `C:\P`},
		{`chrome.exe --user-data-dir-extra=1 --user-data-dir=C:\Real`, "--user-data-dir", `C:\Real`},
		{`chrome.exe --user-data-dir-extra=1`, "--user-data-dir", ``},
		{`chrome.exe`, "--user-data-dir", ``},
		{`chrome.exe --profile-directory="Profile 2"`, "--profile-directory", `Profile 2`},
	}
	for _, c := range cases {
		if got := extractFlagValue(c.cmd, c.flag); got != c.want {
			t.Errorf("extractFlagValue(%q, %q) = %q, want %q", c.cmd, c.flag, got, c.want)
		}
	}
}

func TestCanHandleIsExact(t *testing.T) {
	r := NewRegistry(Env{}, NewExplorerPlugin(), NewChromePlugin(), NewOfficePlugin())
	if p := r.find(`C:\Tools\barcode.exe`); p != nil {
		t.Fatalf("barcode.exe must not be handled by %s", p.ID())
	}
	if p := r.find(`C:\Program Files\Google\Chrome\Application\chrome.exe`); p == nil || p.ID() != "chrome" {
		t.Fatal("chrome.exe must be handled by the chrome plugin")
	}
	if !r.NeedsOpenFiles(`C:\Office\WINWORD.EXE`) || r.NeedsOpenFiles(`C:\Windows\explorer.exe`) {
		t.Fatal("only Office needs open file handles")
	}
}

func TestChromeRelaunchBuildsCleanCommand(t *testing.T) {
	var gotExe, gotCmd string
	env := Env{Launch: func(exe, cmd, wd string) error { gotExe, gotCmd = exe, cmd; return nil }}
	p := NewChromePlugin()
	app := &state.AppState{
		ExecutablePath: `C:\Program Files\Google\Chrome\Application\chrome.exe`,
		CommandLine:    `"C:\Program Files\Google\Chrome\Application\chrome.exe" --user-data-dir="D:\Chrome Data" https://example.com`,
		PluginData:     map[string]state.PluginState{},
	}
	p.Capture(env, app)
	handled, err := p.Relaunch(env, app)
	if !handled || err != nil {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	want := `"C:\Program Files\Google\Chrome\Application\chrome.exe" --user-data-dir="D:\Chrome Data" --restore-last-session`
	if gotExe != app.ExecutablePath || gotCmd != want {
		t.Fatalf("launch(%q, %q), want cmd %q", gotExe, gotCmd, want)
	}
}

func TestExplorerCaptureAndReopenOnlyMissing(t *testing.T) {
	var opened []string
	env := Env{
		Open:           func(p string) error { opened = append(opened, p); return nil },
		ExplorerFolder: func() map[uintptr]string { return map[uintptr]string{1: `C:\A`, 2: `C:\B`} },
	}
	p := NewExplorerPlugin()
	saved := &state.AppState{
		ExecutablePath: `C:\Windows\explorer.exe`,
		Windows:        []state.WindowState{{HWND: 1}, {HWND: 2}, {HWND: 3}},
		PluginData:     map[string]state.PluginState{},
	}
	p.Capture(env, saved)
	if got := saved.PluginData["explorer"].Paths; !reflect.DeepEqual(got, []string{`C:\A`, `C:\B`}) {
		t.Fatalf("paths = %v", got)
	}
	if saved.Windows[0].Path != `C:\A` || saved.Windows[2].Path != "" {
		t.Fatalf("window paths = %+v", saved.Windows)
	}

	current := &state.AppState{PluginData: map[string]state.PluginState{"explorer": {Paths: []string{`c:\a`}}}}
	if n := p.Reopen(env, saved, current); n != 1 || !reflect.DeepEqual(opened, []string{`C:\B`}) {
		t.Fatalf("reopened %d: %v, want only C:\\B", n, opened)
	}
}

func TestOfficeDocsFiltersNoise(t *testing.T) {
	files := []state.FileRef{
		{Path: `C:\Users\me\Report.docx`},
		{Path: `C:\Users\me\~$Report.docx`},
		{Path: `C:\Users\me\AppData\Local\Temp\x.docx`},
		{Path: `C:\Program Files\Microsoft Office\root\Office16\WINWORD.EXE`},
		{Path: `C:\Users\me\Budget.XLSX`},
		{Path: `C:\Users\me\report.docx`},
	}
	got := OfficeDocs(files)
	want := []string{`C:\Users\me\Report.docx`, `C:\Users\me\Budget.XLSX`}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestOfficeRelaunchOpensDocsOrDefers(t *testing.T) {
	var opened []string
	env := Env{Open: func(p string) error { opened = append(opened, p); return nil }}
	p := NewOfficePlugin()
	noDocs := &state.AppState{ExecutablePath: `WINWORD.EXE`, PluginData: map[string]state.PluginState{}}
	if handled, _ := p.Relaunch(env, noDocs); handled {
		t.Fatal("without documents the engine must start Word itself")
	}
	withDocs := &state.AppState{PluginData: map[string]state.PluginState{"office": {Paths: []string{`C:\a.docx`}}}}
	if handled, _ := p.Relaunch(env, withDocs); !handled || len(opened) != 1 {
		t.Fatalf("handled=%v opened=%v", handled, opened)
	}
}

func TestPluginStateSurvivesJSON(t *testing.T) {
	// The old code stored []string in map[string]any, which came back as []any
	// after a disk round trip, so Restore silently did nothing.
	in := state.AppState{PluginData: map[string]state.PluginState{"explorer": {Paths: []string{`C:\A`}}}}
	b, err := jsonRoundTrip(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := pathsOf(&b, "explorer"); !reflect.DeepEqual(got, []string{`C:\A`}) {
		t.Fatalf("paths after JSON = %v", got)
	}
}
