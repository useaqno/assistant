package modelmanager

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

// TestEnsureVerifies downloads a small payload and checks the sha256 gate by
// temporarily overriding the registry entry for the small tier.
func TestEnsureVerifies(t *testing.T) {
	payload := []byte("fake-ggml-model-bytes")
	sum := sha256.Sum256(payload)
	hexSum := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write(payload)
	}))
	defer srv.Close()

	dir := t.TempDir()
	m := New(dir, srv.URL)

	// Override the small spec for the test.
	orig := Registry[TierSmall]
	Registry[TierSmall] = ModelSpec{Tier: TierSmall, Filename: "ggml-small.bin", SizeBytes: int64(len(payload)), SHA256: hexSum}
	defer func() { Registry[TierSmall] = orig }()

	path, err := m.Ensure(context.Background(), TierSmall, nil)
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if filepath.Base(path) != "ggml-small.bin" {
		t.Errorf("path = %s", path)
	}
	got, _ := os.ReadFile(path)
	if string(got) != string(payload) {
		t.Error("downloaded content mismatch")
	}
	if st := m.Status(TierSmall); !st.Present {
		t.Error("expected present after download")
	}
}

func TestEnsureRejectsBadChecksum(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("corrupted"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	m := New(dir, srv.URL)
	orig := Registry[TierSmall]
	Registry[TierSmall] = ModelSpec{Tier: TierSmall, Filename: "ggml-small.bin", SizeBytes: 999, SHA256: "deadbeef"}
	defer func() { Registry[TierSmall] = orig }()

	if _, err := m.Ensure(context.Background(), TierSmall, nil); err == nil {
		t.Fatal("expected checksum mismatch error")
	}
	if _, err := os.Stat(filepath.Join(dir, "ggml-small.bin")); !os.IsNotExist(err) {
		t.Error("bad download should have been removed")
	}
}
