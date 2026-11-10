package snapshot

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const fileFormatVersion = 2

type timelineFile struct {
	Version  int          `json:"version"`
	Timeline *appTimeline `json:"timeline"`
}

// Load reads persisted timelines. Missing or corrupt files are skipped.
func (e *Engine) Load() error {
	if e.cfg.StorageDir == "" {
		return nil
	}
	files, err := filepath.Glob(filepath.Join(e.cfg.StorageDir, "*.json.gz"))
	if err != nil {
		return err
	}
	leftovers, _ := filepath.Glob(filepath.Join(e.cfg.StorageDir, "*.json.gz.tmp"))
	for _, f := range leftovers {
		_ = os.Remove(f) // from an interrupted save
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, f := range files {
		tl, err := readTimeline(f)
		if err != nil || tl == nil || tl.AppID == "" {
			continue
		}
		e.trimLocked(tl)
		tl.dirty = false
		if len(tl.Snapshots) == 0 {
			_ = os.Remove(f)
			continue
		}
		e.apps[tl.AppID] = tl
	}
	return nil
}

// Save writes every changed timeline to disk atomically.
func (e *Engine) Save() error {
	if e.cfg.StorageDir == "" {
		return nil
	}
	type job struct {
		path string
		data []byte
		tl   *appTimeline
	}
	var jobs []job
	e.mu.Lock()
	for _, tl := range e.apps {
		if !tl.dirty {
			continue
		}
		raw, err := json.Marshal(timelineFile{Version: fileFormatVersion, Timeline: tl})
		if err != nil {
			continue
		}
		tl.dirty = false
		jobs = append(jobs, job{path: e.fileFor(tl.AppID), data: raw, tl: tl})
	}
	e.mu.Unlock()
	if len(jobs) == 0 {
		return nil
	}
	if err := os.MkdirAll(e.cfg.StorageDir, 0o755); err != nil {
		for _, j := range jobs {
			e.markDirty(j.tl)
		}
		return err
	}

	var firstErr error
	for _, j := range jobs {
		if err := writeGzipAtomic(j.path, j.data); err != nil {
			e.markDirty(j.tl)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (e *Engine) Close() error { return e.Save() }

// DeleteApp removes an app's whole timeline from memory and disk.
func (e *Engine) DeleteApp(appID string) error {
	e.mu.Lock()
	delete(e.apps, appID)
	delete(e.restores, appID)
	e.mu.Unlock()
	if e.cfg.StorageDir == "" {
		return nil
	}
	if err := os.Remove(e.fileFor(appID)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (e *Engine) markDirty(tl *appTimeline) {
	e.mu.Lock()
	tl.dirty = true
	e.mu.Unlock()
}

func (e *Engine) fileFor(appID string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			return r
		}
		return '_'
	}, appID)
	return filepath.Join(e.cfg.StorageDir, safe+".json.gz")
}

func readTimeline(path string) (*appTimeline, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer zr.Close()
	raw, err := io.ReadAll(io.LimitReader(zr, 64<<20))
	if err != nil {
		return nil, err
	}
	var tf timelineFile
	if err := json.Unmarshal(raw, &tf); err != nil {
		return nil, err
	}
	if tf.Version != fileFormatVersion {
		return nil, errors.New("unsupported timeline version")
	}
	return tf.Timeline, nil
}

func writeGzipAtomic(path string, data []byte) error {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	if _, err := zw.Write(data); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		// Some file system filters (sandboxes, sync clients, AV) refuse the rename.
		// Fall back to a direct write: a torn file is skipped by Load anyway.
		_ = os.Remove(tmp)
		return os.WriteFile(path, buf.Bytes(), 0o644)
	}
	return nil
}
