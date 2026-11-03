package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT13_OpenCreatesIndex(t *testing.T) {
	st := filepath.Join(t.TempDir(), "state")
	s := jsonLines(t, ok(t, st, nil, "stats", "--json"))[0]
	eq(t, "schema_version", s["schema_version"], 1)
	eq(t, "chunks", s["chunks"], 0)
	di, err := os.Stat(st)
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "dir mode", di.Mode().Perm(), os.FileMode(0o700))
	fi, err := os.Stat(filepath.Join(st, "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	eq(t, "file mode", fi.Mode().Perm(), os.FileMode(0o600))
	db := openIdx(t, st)
	eq(t, "user_version", count(t, db, "PRAGMA user_version"), 1)
	var jm string
	db.QueryRow("PRAGMA journal_mode").Scan(&jm)
	eq(t, "journal_mode", jm, "wal")
	eq(t, "chunks table", count(t, db, "SELECT count(*) FROM chunks"), 0)
	eq(t, "cursors table", count(t, db, "SELECT count(*) FROM cursors"), 0)
}
