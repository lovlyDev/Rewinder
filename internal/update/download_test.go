package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func serve(t *testing.T, exe []byte, sums string) *Release {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/exe":
			_, _ = w.Write(exe)
		case "/sums":
			_, _ = w.Write([]byte(sums))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return &Release{Version: "9.9.9", exeURL: srv.URL + "/exe", sumsURL: srv.URL + "/sums"}
}

func TestDownloadVerifiesChecksum(t *testing.T) {
	exe := []byte("MZ pretend this is a new Rewinder")
	sum := sha256.Sum256(exe)
	r := serve(t, exe, hex.EncodeToString(sum[:])+"  Rewinder.exe\n")
	dest := filepath.Join(t.TempDir(), "Rewinder.exe")

	var last int
	path, err := Download(context.Background(), r, dest, func(p int) { last = p })
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(exe) || last != 100 {
		t.Fatalf("downloaded %q, progress %d", got, last)
	}
}

func TestDownloadRejectsTamperedFile(t *testing.T) {
	r := serve(t, []byte("evil"), "0000000000000000000000000000000000000000000000000000000000000000  Rewinder.exe\n")
	dest := filepath.Join(t.TempDir(), "Rewinder.exe")
	if _, err := Download(context.Background(), r, dest, nil); err == nil || err.Error() != "update.checksum" {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(dest + ".new"); !os.IsNotExist(err) {
		t.Fatal("a file that failed verification was left behind")
	}
}

func TestSwapKeepsOldUntilCleanup(t *testing.T) {
	dir := t.TempDir()
	cur := filepath.Join(dir, "Rewinder.exe")
	_ = os.WriteFile(cur, []byte("old"), 0o755)
	_ = os.WriteFile(cur+".new", []byte("new"), 0o755)
	if err := Swap(cur, cur+".new"); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(cur); string(b) != "new" {
		t.Fatalf("current = %q", b)
	}
	CleanupOld(cur)
	if _, err := os.Stat(cur + ".old"); !os.IsNotExist(err) {
		t.Fatal(".old not cleaned up")
	}
}
