package ingest

import (
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"agentmem/internal/store"
)

// Result summarizes an IngestAll run.
type Result struct {
	Files    int
	Inserted int
	Skipped  int
}

// walkRoot ingests every *.jsonl file under root in walk order, updating r. A missing root is
// skipped, not an error.
func walkRoot(db *sql.DB, storeName, root string, r *Result) error {
	if _, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		ins, sk, err := IngestJSONL(db, storeName, path)
		if err != nil {
			return err
		}
		r.Files++
		r.Inserted += ins
		r.Skipped += sk
		return nil
	})
}

// IngestAll ingests every Claude Code and pi transcript under the configured roots:
// AGENTMEM_CLAUDE_DIR (default ~/.claude/projects) and AGENTMEM_PI_DIR (default
// ~/.pi/agent/sessions).
func IngestAll(db *sql.DB) (Result, error) {
	var r Result
	home, _ := os.UserHomeDir()
	roots := []struct{ storeName, dir string }{
		{"claude", envDir("AGENTMEM_CLAUDE_DIR", filepath.Join(home, ".claude", "projects"))},
		{"pi", envDir("AGENTMEM_PI_DIR", filepath.Join(home, ".pi", "agent", "sessions"))},
	}
	for _, root := range roots {
		if err := walkRoot(db, root.storeName, root.dir, &r); err != nil {
			return r, err
		}
	}
	dbf := envDir("AGENTMEM_HERMES_DB", filepath.Join(home, ".hermes", "state.db"))
	if _, err := os.Stat(dbf); err == nil {
		r.Files++
		after := int64(0)
		if c, ok := store.GetCursor(db, "hermes", dbf); ok {
			after, _ = strconv.ParseInt(c, 10, 64)
		}
		msgs, maxID, err := ParseHermes(dbf, after)
		if err != nil {
			return r, err
		}
		known := LoadKnown(db, dbf)
		ins, err := store.Upsert(db, Process(msgs, known))
		if err != nil {
			return r, err
		}
		if err := SaveKnown(db, dbf, known); err != nil {
			return r, err
		}
		if maxID > after {
			if err := store.SetCursor(db, "hermes", dbf, strconv.FormatInt(maxID, 10)); err != nil {
				return r, err
			}
		}
		r.Inserted += ins
	}
	return r, nil
}

// envDir returns the value of key, or def when unset or empty.
func envDir(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
