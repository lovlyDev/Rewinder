package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Repository releases are published to (see .github/workflows/release.yml).
const Repository = "lovlyDev/Rewinder"

const (
	// AssetExe is the portable executable attached to every release.
	AssetExe = "Rewinder.exe"
	// AssetSums lists "sha256  file" for every asset of the release.
	AssetSums = "SHA256SUMS.txt"

	maxExeSize  = 200 << 20
	maxSumsSize = 64 << 10
)

// Release is the newest published release.
type Release struct {
	Version string `json:"version"`
	Notes   string `json:"notes"`
	URL     string `json:"url"`
	exeURL  string
	sumsURL string
}

var client = &http.Client{Timeout: 60 * time.Second}

func get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	// A generic client name: nothing about the user leaves the machine.
	req.Header.Set("User-Agent", "Rewinder-Updater")
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
	}
	return resp, nil
}

// Latest asks GitHub for the newest non-draft, non-prerelease release.
func Latest(ctx context.Context) (*Release, error) {
	// REWINDER_UPDATE_API points the check at a test server (end-to-end tests of
	// the updater). Downloads are still verified against the release checksums.
	url := "https://api.github.com/repos/" + Repository + "/releases/latest"
	if v := os.Getenv("REWINDER_UPDATE_API"); v != "" {
		url = v
	}
	resp, err := get(ctx, url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var body struct {
		TagName string `json:"tag_name"`
		Body    string `json:"body"`
		HTMLURL string `json:"html_url"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&body); err != nil {
		return nil, err
	}
	r := &Release{Version: strings.TrimPrefix(body.TagName, "v"), Notes: body.Body, URL: body.HTMLURL}
	for _, a := range body.Assets {
		switch a.Name {
		case AssetExe:
			r.exeURL = a.URL
		case AssetSums:
			r.sumsURL = a.URL
		}
	}
	return r, nil
}

// Installable reports whether the release carries what Download needs.
func (r *Release) Installable() bool { return r.exeURL != "" && r.sumsURL != "" }

// Download fetches the release executable next to dest (dest + ".new"),
// checks its SHA-256 against the release's checksum file and returns the path.
// progress gets 0…100.
func Download(ctx context.Context, r *Release, dest string, progress func(int)) (string, error) {
	if !r.Installable() {
		return "", errors.New("update.assets")
	}
	sumsResp, err := get(ctx, r.sumsURL)
	if err != nil {
		return "", err
	}
	sums, err := io.ReadAll(io.LimitReader(sumsResp.Body, maxSumsSize))
	sumsResp.Body.Close()
	if err != nil {
		return "", err
	}
	want := checksumFor(string(sums), AssetExe)
	if want == "" {
		return "", errors.New("update.checksum")
	}

	resp, err := get(ctx, r.exeURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tmp := dest + ".new"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	pr := &progressReader{r: io.LimitReader(resp.Body, maxExeSize), total: resp.ContentLength, report: progress}
	_, copyErr := io.Copy(io.MultiWriter(f, h), pr)
	closeErr := f.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmp)
		return "", errors.Join(copyErr, closeErr)
	}
	if got := hex.EncodeToString(h.Sum(nil)); !strings.EqualFold(got, want) {
		_ = os.Remove(tmp)
		return "", errors.New("update.checksum")
	}
	return tmp, nil
}

// checksumFor finds a file's hash in sha256sum output ("hash  name" or "hash *name").
func checksumFor(sums, name string) string {
	for _, line := range strings.Split(sums, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			return fields[0]
		}
	}
	return ""
}

type progressReader struct {
	r      io.Reader
	total  int64
	done   int64
	last   int
	report func(int)
}

func (p *progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	p.done += int64(n)
	if p.report != nil && p.total > 0 {
		if pct := int(p.done * 100 / p.total); pct != p.last {
			p.last = pct
			p.report(pct)
		}
	}
	return n, err
}

// Swap replaces the running executable with the downloaded one. Windows lets a
// running exe be renamed, so: current -> .old, new -> current. The .old file is
// removed by CleanupOld at the next start.
func Swap(current, downloaded string) error {
	old := current + ".old"
	_ = os.Remove(old)
	if err := os.Rename(current, old); err != nil {
		return err
	}
	if err := os.Rename(downloaded, current); err != nil {
		_ = os.Rename(old, current) // put the working version back
		return err
	}
	return nil
}

// CleanupOld removes what a previous update left behind.
func CleanupOld(current string) {
	_ = os.Remove(current + ".old")
	_ = os.Remove(current + ".new")
	matches, _ := filepath.Glob(current + ".new*")
	for _, m := range matches {
		_ = os.Remove(m)
	}
}
