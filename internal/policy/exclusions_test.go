package policy

import (
	"reflect"
	"testing"
)

func TestExclusionsPersistAndMatchByFileName(t *testing.T) {
	dir := t.TempDir()
	e := LoadExclusions(dir)
	if e.Excluded(`C:\Program Files\Google\Chrome\Application\chrome.exe`) {
		t.Fatal("nothing is excluded at first")
	}
	if err := e.Set(`C:\Program Files\Google\Chrome\Application\Chrome.EXE`, true); err != nil {
		t.Fatal(err)
	}
	if err := e.Set("telegram.exe", true); err != nil {
		t.Fatal(err)
	}
	if !e.Excluded(`D:\other\place\chrome.exe`) {
		t.Fatal("exclusion is by file name, case-insensitive")
	}

	again := LoadExclusions(dir)
	if got := again.List(); !reflect.DeepEqual(got, []string{"chrome.exe", "telegram.exe"}) {
		t.Fatalf("reloaded %v", got)
	}
	if err := again.Set("chrome.exe", false); err != nil {
		t.Fatal(err)
	}
	if LoadExclusions(dir).Excluded("chrome.exe") {
		t.Fatal("removal was not saved")
	}
}

func TestExclusionsRejectBadNames(t *testing.T) {
	e := LoadExclusions(t.TempDir())
	for _, bad := range []string{"", "  ", "notes.txt", `C:\folder\`} {
		if err := e.Set(bad, true); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}
