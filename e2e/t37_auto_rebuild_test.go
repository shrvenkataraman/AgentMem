package e2e

import (
	"database/sql"
	"os"
	"testing"
)

func TestT37_VersionMismatchRebuilds(t *testing.T) {
	st := corpusState(t)
	db, err := sql.Open("sqlite", "file:"+st+"/index.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA user_version=99"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	so, se, code := runEnv(t, st, corpus(), "stats", "--json")
	eq(t, "exit", code, 0)
	has(t, se, "rebuilding")
	eq(t, "chunks", jsonLines(t, so)[0]["chunks"], 21)
	if !exists(st + "/index.sqlite.bak") {
		t.Fatal("old index must be kept as .bak")
	}
}

func TestT37_CorruptFileRebuilds(t *testing.T) {
	st := corpusState(t)
	os.Remove(st + "/index.sqlite-wal")
	os.Remove(st + "/index.sqlite-shm")
	os.WriteFile(st+"/index.sqlite", []byte("this is definitely not a sqlite database file........................"), 0o600)
	so, se, code := runEnv(t, st, corpus(), "stats", "--json")
	eq(t, "exit", code, 0)
	has(t, se, "rebuilding")
	eq(t, "chunks", jsonLines(t, so)[0]["chunks"], 21)
	if !exists(st + "/index.sqlite.bak") {
		t.Fatal("corrupt index must be kept as .bak")
	}
}

func TestT37_CorruptWithNoSourcesGivesEmptyIndex(t *testing.T) {
	st := t.TempDir()
	os.WriteFile(st+"/index.sqlite", []byte("garbage garbage garbage garbage garbage garbage garbage garbage"), 0o600)
	eq(t, "chunks", statsOf(t, st, nil)["chunks"], 0)
}
