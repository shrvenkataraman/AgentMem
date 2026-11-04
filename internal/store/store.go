package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ErrNeedsRebuild is returned by Open/OpenFile when the index on disk cannot be used: its schema
// version is newer than SchemaVersion or the file is corrupt ("file is not
// a database" / "malformed"). Callers rebuild from the sources.
var ErrNeedsRebuild = errors.New("index needs rebuild")

// needsRebuild maps a sqlite error message to ErrNeedsRebuild when it says the file is not a
// database or is malformed, and otherwise wraps err unchanged.
func needsRebuild(err error) error {
	if err == nil {
		return nil
	}
	m := err.Error()
	if strings.Contains(m, "file is not a database") || strings.Contains(m, "malformed") {
		return ErrNeedsRebuild
	}
	return fmt.Errorf("open index: %w", err)
}

// StateDir is $AGENTMEM_STATE_DIR, else ~/.local/share/agentmem.
func StateDir() string {
	if d := os.Getenv("AGENTMEM_STATE_DIR"); d != "" {
		return d
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".local", "share", "agentmem")
}

// Open opens (creating if needed) the index in dir.
func Open(dir string) (*sql.DB, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return OpenFile(filepath.Join(dir, "index.sqlite"))
}

// OpenFile opens the index database at path with WAL enabled.
func OpenFile(path string) (*sql.DB, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	f.Close()
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=secure_delete(ON)")
	if err != nil {
		return nil, needsRebuild(err)
	}
	db.SetMaxOpenConns(1)
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		db.Close()
		return nil, needsRebuild(err)
	}
	if v > SchemaVersion {
		db.Close()
		return nil, ErrNeedsRebuild
	}
	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, needsRebuild(err)
	}
	if v != 0 && v < SchemaVersion {
		if _, err := Reredact(db); err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.Exec("INSERT INTO chunks_fts(chunks_fts) VALUES('optimize')"); err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.Exec("INSERT INTO chunks_trgm(chunks_trgm) VALUES('optimize')"); err != nil {
			db.Close()
			return nil, err
		}
		if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
			db.Close()
			return nil, err
		}
		os.Remove(path + ".bak") // ignore: best-effort scrub
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version=%d", SchemaVersion)); err != nil {
		db.Close()
		return nil, needsRebuild(err)
	}
	return db, nil
}
