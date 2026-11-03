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

// Open opens (creating if needed) the index in dir. Stub until the "store open" ticket.
func Open(dir string) (*sql.DB, error) {
	panic("not implemented")
}
