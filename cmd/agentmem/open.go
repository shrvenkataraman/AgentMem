package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"agentmem/internal/ingest"
	"agentmem/internal/store"
)

// openDB opens the index in the state dir. Every command that needs the DB calls this. When the
// index on disk is unusable (unknown schema version or corruption) it is rebuilt from the sources
// first and the old file is kept as index.sqlite.bak.
func openDB() (*sql.DB, error) {
	db, err := store.Open(store.StateDir())
	if !errors.Is(err, store.ErrNeedsRebuild) {
		return db, err
	}
	fmt.Fprintln(os.Stderr, "agentmem: index unusable, rebuilding from sources")
	_, _, _, err = store.Rebuild(store.StateDir(), true, func(d *sql.DB) error { _, e := ingest.IngestAll(d); return e })
	if err != nil {
		return nil, err
	}
	return store.Open(store.StateDir())
}
