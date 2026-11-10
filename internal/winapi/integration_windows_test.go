//go:build windows && integration

package winapi

import (
	"bytes"
	"image/png"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/windows"
)

// The old handle-table struct had the wrong layout, so no file was ever found.
func TestOpenFilesFindsOurOwnFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Rewinder open-files test.docx")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	// A pipe-like handle in the same process must not make enumeration hang.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err == nil {
		defer l.Close()
	}

	start := time.Now()
	files := OpenFiles(uint32(os.Getpid()), 2*time.Second)
	if time.Since(start) > 3*time.Second {
		t.Fatalf("OpenFiles took %v", time.Since(start))
	}
	want, _ := filepath.EvalSymlinks(path)
	for _, p := range files {
		if strings.EqualFold(p, path) || strings.EqualFold(p, want) {
			return
		}
	}
	t.Fatalf("%s not among %d open files: %v", path, len(files), files)
}

func TestProcessInfoOfSelf(t *testing.T) {
	pid := uint32(os.Getpid())
	exe, _ := os.Executable()
	if got := ExePath(pid); !strings.EqualFold(got, exe) {
		t.Fatalf("ExePath = %q, want %q", got, exe)
	}
	if cl := CommandLine(pid); !strings.Contains(strings.ToLower(cl), strings.ToLower(filepath.Base(exe))) {
		t.Fatalf("CommandLine = %q", cl)
	}
	if !ProcessAlive(pid, exe) || ProcessAlive(pid, `C:\not\this.exe`) {
		t.Fatal("ProcessAlive must check the executable, not just the PID")
	}
	if AppUserModelID(pid) != "" {
		t.Fatal("a test binary is not a packaged app")
	}
}

func TestIsElevatedMatchesOwnToken(t *testing.T) {
	want := windows.GetCurrentProcessToken().IsElevated()
	if got := IsElevated(uint32(os.Getpid())); got != want {
		t.Fatalf("IsElevated(self) = %v, token says %v", got, want)
	}
	if IsElevated(0) {
		t.Fatal("PID 0 cannot be opened and must not count as elevated")
	}
}

func TestAppIconPNGOfExplorer(t *testing.T) {
	data, err := AppIconPNG(filepath.Join(os.Getenv("WINDIR"), "explorer.exe"), 64)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 64 || b.Dy() != 64 {
		t.Fatalf("size %v", b)
	}
	visible := 0
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			if _, _, _, a := img.At(x, y).RGBA(); a > 0 {
				visible++
			}
		}
	}
	if visible < 200 || visible == 64*64 {
		t.Fatalf("icon looks wrong: %d of %d pixels visible", visible, 64*64)
	}
}

func TestEnumWindowsManyTimes(t *testing.T) {
	// windows.NewCallback has a fixed limit; the old code created one per call
	// and would panic after a few thousand captures.
	for i := 0; i < 3000; i++ {
		if len(TopLevelWindows()) == 0 {
			t.Fatal("no top-level windows")
		}
	}
}


func waitTrue(cond func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return cond()
}

// Opens one Explorer window on a temp folder, checks it is reported with the
// right path, then closes only that window.
func TestExplorerFoldersSeesOpenFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "RewinderExplorerTest")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ShellOpen(dir); err != nil {
		t.Fatal(err)
	}
	var found uintptr
	deadline := time.Now().Add(10 * time.Second)
	for found == 0 && time.Now().Before(deadline) {
		for hwnd, path := range ExplorerFolders() {
			if hwnd == 0 || path == "" {
				t.Fatalf("bad entry %v %q", hwnd, path)
			}
			if strings.EqualFold(filepath.Clean(path), filepath.Clean(dir)) {
				found = hwnd
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	if found == 0 {
		t.Fatalf("folder %s not reported by ExplorerFolders", dir)
	}
	defer procPostMessageW.Call(found, 0x0010, 0, 0) // WM_CLOSE
	// right after opening the window may still be cloaked by the open animation
	if !waitTrue(func() bool { return IsAppWindow(found) }, 5*time.Second) {
		t.Fatalf("an Explorer folder window must count as an app window: class=%q visible=%v owner=%v cloaked=%v exstyle=%#x",
			ClassName(found), isVisible(found), owner(found), isCloaked(found), exStyle(found))
	}
	if ClassName(found) != "CabinetWClass" {
		t.Fatalf("class = %q", ClassName(found))
	}
}

// Read-only: if a packaged app is running, its AUMID must be readable.
func TestAppUserModelIDOfRunningPackagedApp(t *testing.T) {
	for _, h := range TopLevelWindows() {
		pid := WindowPID(h)
		if strings.Contains(strings.ToLower(ExePath(pid)), `\windowsapps\`) {
			id := AppUserModelID(pid)
			if !strings.Contains(id, "!") {
				t.Fatalf("packaged app %s: AUMID = %q", ExePath(pid), id)
			}
			t.Logf("%s -> %s", filepath.Base(ExePath(pid)), id)
			return
		}
	}
	t.Skip("no packaged app is running")
}

func TestNormalizeFinalPath(t *testing.T) {
	cases := map[string]string{
		`\\?\C:\Users\me\a.docx`:      `C:\Users\me\a.docx`,
		`\\?\UNC\server\share\b.xlsx`: `\\server\share\b.xlsx`,
		`\\?\Volume{1234}\x.txt`:      ``,
		`C:\plain.txt`:                `C:\plain.txt`,
	}
	for in, want := range cases {
		if got := NormalizeFinalPath(in); got != want {
			t.Errorf("NormalizeFinalPath(%q) = %q, want %q", in, got, want)
		}
	}
}
