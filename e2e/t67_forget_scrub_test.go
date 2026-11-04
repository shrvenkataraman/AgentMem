package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT67_ForgetScrubsFTSAndBak(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "zqxforgetme67", "keep67", "/p67"), "--store", "claude")
	bak := filepath.Join(st, "index.sqlite.bak")
	writeFile(t, bak, "old")
	ok(t, st, nil, "forget", "--session", "zqxforgetme67")
	db := openIdx(t, st)
	eq(t, "fts bytes", count(t, db, "SELECT count(*) FROM chunks_fts_data WHERE instr(block, CAST('zqxforgetme67' AS BLOB)) > 0"), 0)
	db.Close()
	if _, err := os.Stat(bak); !os.IsNotExist(err) {
		t.Fatal("forget must remove index.sqlite.bak")
	}
}

func TestT67_ReredactRemovesBak(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s67c", "s67d", "/p67"), "--store", "claude")
	db := openIdx(t, st)
	if _, err := db.Exec("PRAGMA user_version=1"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	bak := filepath.Join(st, "index.sqlite.bak")
	writeFile(t, bak, "old")
	ok(t, st, nil, "stats")
	if _, err := os.Stat(bak); !os.IsNotExist(err) {
		t.Fatal("re-redact upgrade must remove index.sqlite.bak")
	}
}
