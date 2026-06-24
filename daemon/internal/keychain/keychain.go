// Package keychain stores secrets (LLM API keys, SSH credentials) in the macOS
// Keychain via the system `security` tool, so they never touch the SQLite file.
// On non-macOS platforms it falls back to an in-memory store (dev only) and
// reports unavailability, keeping the rest of the daemon portable.
package keychain

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
	"sync"
)

// ErrNotFound is returned when no secret exists for the service/account pair.
var ErrNotFound = errors.New("keychain: item not found")

const securityBin = "/usr/bin/security"

// Available reports whether the OS keychain backend is usable.
func Available() bool {
	if runtime.GOOS != "darwin" {
		return false
	}
	_, err := exec.LookPath(securityBin)
	return err == nil
}

// Set stores (or replaces) a generic password secret.
func Set(service, account, secret string) error {
	if !Available() {
		mem.set(service, account, secret)
		return nil
	}
	// -U updates in place if the item already exists.
	cmd := exec.Command(securityBin, "add-generic-password",
		"-s", service, "-a", account, "-w", secret, "-U")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New("keychain set: " + strings.TrimSpace(string(out)))
	}
	return nil
}

// Get returns the stored secret, or ErrNotFound.
func Get(service, account string) (string, error) {
	if !Available() {
		return mem.get(service, account)
	}
	cmd := exec.Command(securityBin, "find-generic-password",
		"-s", service, "-a", account, "-w")
	out, err := cmd.Output()
	if err != nil {
		return "", ErrNotFound
	}
	return strings.TrimRight(string(out), "\n"), nil
}

// Delete removes a secret; missing items are not an error.
func Delete(service, account string) error {
	if !Available() {
		mem.del(service, account)
		return nil
	}
	cmd := exec.Command(securityBin, "delete-generic-password",
		"-s", service, "-a", account)
	_ = cmd.Run()
	return nil
}

// memStore is the non-darwin dev fallback.
type memStore struct {
	mu sync.RWMutex
	m  map[string]string
}

var mem = &memStore{m: map[string]string{}}

func key(s, a string) string { return s + "\x00" + a }

func (s *memStore) set(svc, acc, secret string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key(svc, acc)] = secret
}

func (s *memStore) get(svc, acc string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.m[key(svc, acc)]
	if !ok {
		return "", ErrNotFound
	}
	return v, nil
}

func (s *memStore) del(svc, acc string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, key(svc, acc))
}
