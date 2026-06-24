// Package store is the SQLite-backed persistence layer for aqnod. It uses the
// pure-Go modernc.org/sqlite driver (no cgo) so the daemon stays a single
// self-contained sidecar binary. See docs/context.md §9–10.
package store

import (
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Store wraps the database handle and exposes typed domain operations.
type Store struct {
	db      *sql.DB
	hasFTS  bool
	dataDir string
}

// DefaultPath returns the on-disk database location, creating its directory.
// macOS: ~/Library/Application Support/io.aqno/aqno.db
func DefaultPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "aqno.db"), nil
}

// DataDir returns (and creates) the app data directory.
func DataDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "io.aqno")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Open initialises the database at path (":memory:" for tests), applies the
// recommended PRAGMAs, runs migrations and seeds defaults.
func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=busy_timeout(5000)&_pragma=foreign_keys(ON)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // WAL + single writer keeps it simple and safe
	if err := db.Ping(); err != nil {
		return nil, err
	}
	s := &Store{db: db}
	if dir, derr := DataDir(); derr == nil {
		s.dataDir = dir
	}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	if err := s.seed(); err != nil {
		db.Close()
		return nil, fmt.Errorf("seed: %w", err)
	}
	return s, nil
}

// Close releases the database handle.
func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schemaSQL); err != nil {
		return err
	}
	// FTS5 is an optional build feature; create the virtual table only if the
	// driver supports it, otherwise fall back to LIKE-based search.
	_, err := s.db.Exec(`CREATE VIRTUAL TABLE IF NOT EXISTS busca USING fts5(tipo, ref_id UNINDEXED, texto)`)
	s.hasFTS = err == nil
	return nil
}

// HasFTS reports whether full-text search is available.
func (s *Store) HasFTS() bool { return s.hasFTS }
