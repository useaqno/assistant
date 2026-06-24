// Package modelmanager downloads and verifies whisper.cpp GGML models on demand
// by quality tier. Checksums are pinned (verified against the HuggingFace API on
// 2026-06-23) so a mirror can be untrusted transport. Pure Go, no cgo.
package modelmanager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Tier is a whisper model quality tier.
type Tier string

const (
	TierSmall        Tier = "small"
	TierMedium       Tier = "medium"
	TierLargeV3Turbo Tier = "large-v3-turbo"
)

// ModelSpec describes a downloadable model.
type ModelSpec struct {
	Tier      Tier
	Filename  string
	SizeBytes int64
	SHA256    string // lowercase hex
}

// Registry maps tiers to their pinned spec.
var Registry = map[Tier]ModelSpec{
	TierSmall:        {TierSmall, "ggml-small.bin", 487601967, "1be3a9b2063867b937e64e2ec7483364a79917e157fa98c5d94b5c1fffea987b"},
	TierMedium:       {TierMedium, "ggml-medium.bin", 1533763059, "6c14d5adee5f86394037b4e4e8b59f1673b6cee10e3cf0b11bbdbee79c156208"},
	TierLargeV3Turbo: {TierLargeV3Turbo, "ggml-large-v3-turbo.bin", 1624555275, "1fc70f774d38eb169993ac391eea357ef47c88757ef72ee5943879b7e8e2bc69"},
}

const defaultMirror = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main"

// Manager downloads and tracks models in a directory.
type Manager struct {
	dir        string
	mirrorBase string
	client     *http.Client
}

// New creates a manager. mirrorBase may be "" (uses the default HF mirror).
func New(dir, mirrorBase string) *Manager {
	if mirrorBase == "" {
		mirrorBase = defaultMirror
	}
	return &Manager{dir: dir, mirrorBase: strings.TrimRight(mirrorBase, "/"), client: http.DefaultClient}
}

// Status reports presence/verification of a tier on disk.
type Status struct {
	Tier      Tier   `json:"tier"`
	Present   bool   `json:"present"`
	Verified  bool   `json:"verified"`
	Bytes     int64  `json:"bytes"`
	SizeBytes int64  `json:"sizeBytes"`
}

// Path returns the local file path for a tier (whether or not it exists).
func (m *Manager) Path(t Tier) (string, error) {
	spec, ok := Registry[t]
	if !ok {
		return "", fmt.Errorf("unknown tier %q", t)
	}
	return filepath.Join(m.dir, spec.Filename), nil
}

// Status returns disk status without hashing (cheap).
func (m *Manager) Status(t Tier) Status {
	spec, ok := Registry[t]
	st := Status{Tier: t}
	if !ok {
		return st
	}
	st.SizeBytes = spec.SizeBytes
	p := filepath.Join(m.dir, spec.Filename)
	if fi, err := os.Stat(p); err == nil {
		st.Present = true
		st.Bytes = fi.Size()
		st.Verified = fi.Size() == spec.SizeBytes
	}
	return st
}

// AllStatus returns the status for every tier.
func (m *Manager) AllStatus() []Status {
	return []Status{m.Status(TierSmall), m.Status(TierMedium), m.Status(TierLargeV3Turbo)}
}

// Progress is streamed during a download.
type Progress struct {
	Tier       Tier
	Downloaded int64
	Total      int64
	Done       bool
	Err        error
}

// Ensure downloads (if needed) and verifies a model, returning its path.
// Progress is emitted on prog when non-nil (closed on return).
func (m *Manager) Ensure(ctx context.Context, t Tier, prog chan<- Progress) (string, error) {
	if prog != nil {
		defer close(prog)
	}
	spec, ok := Registry[t]
	if !ok {
		return "", fmt.Errorf("unknown tier %q", t)
	}
	if err := os.MkdirAll(m.dir, 0o755); err != nil {
		return "", err
	}
	final := filepath.Join(m.dir, spec.Filename)

	// Already present and the right size? Trust it (full re-hash is expensive).
	if fi, err := os.Stat(final); err == nil && fi.Size() == spec.SizeBytes {
		return final, nil
	}

	url := m.mirrorBase + "/" + spec.Filename
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	res, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: status %d", spec.Filename, res.StatusCode)
	}

	tmp := final + ".partial"
	f, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	hasher := sha256.New()
	total := res.ContentLength
	if total <= 0 {
		total = spec.SizeBytes
	}
	written, err := copyWithProgress(f, io.TeeReader(res.Body, hasher), t, total, prog)
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return "", err
	}

	sum := hex.EncodeToString(hasher.Sum(nil))
	if sum != spec.SHA256 {
		os.Remove(tmp)
		return "", fmt.Errorf("checksum mismatch for %s (got %s)", spec.Filename, sum)
	}
	_ = written
	if err := os.Rename(tmp, final); err != nil {
		return "", err
	}
	return final, nil
}

// Verify re-hashes a file against a spec.
func Verify(path string, spec ModelSpec) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != spec.SHA256 {
		return errors.New("checksum mismatch")
	}
	return nil
}

func copyWithProgress(dst io.Writer, src io.Reader, t Tier, total int64, prog chan<- Progress) (int64, error) {
	buf := make([]byte, 256*1024)
	var written int64
	for {
		n, err := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return written, werr
			}
			written += int64(n)
			emit(prog, Progress{Tier: t, Downloaded: written, Total: total})
		}
		if err == io.EOF {
			emit(prog, Progress{Tier: t, Downloaded: written, Total: total, Done: true})
			return written, nil
		}
		if err != nil {
			emit(prog, Progress{Tier: t, Err: err})
			return written, err
		}
	}
}

func emit(prog chan<- Progress, p Progress) {
	if prog == nil {
		return
	}
	select {
	case prog <- p:
	default:
	}
}
