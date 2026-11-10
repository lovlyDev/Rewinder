//go:build windows && integration

package update

import (
	"os"
	"testing"
)

func TestSignedRecognisesEmbeddedSignatures(t *testing.T) {
	self, _ := os.Executable()
	if Signed(self) {
		t.Fatal("a freshly built test binary is not signed")
	}
	// Any of these carry an embedded Authenticode signature when installed.
	for _, p := range []string{
		os.Getenv("ProgramFiles") + `\Google\Chrome\Application\chrome.exe`,
		os.Getenv("ProgramFiles(x86)") + `\Microsoft\Edge\Application\msedge.exe`,
		os.Getenv("ProgramFiles") + `\Mozilla Firefox\firefox.exe`,
		os.Getenv("ProgramFiles") + `\Git\git-bash.exe`,
	} {
		if _, err := os.Stat(p); err == nil {
			if !Signed(p) {
				t.Fatalf("%s should verify as signed", p)
			}
			return
		}
	}
	t.Skip("no known signed executable installed")
}
