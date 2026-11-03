package store

import (
	"database/sql"
	"os"
	"path/filepath"
)

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
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(Schema); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec("PRAGMA user_version=1"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
