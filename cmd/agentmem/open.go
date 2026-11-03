package main

import (
	"database/sql"

	"agentmem/internal/store"
)

// openDB opens the index in the state dir. Every command that needs the DB calls this.
func openDB() (*sql.DB, error) {
	return store.Open(store.StateDir())
}
