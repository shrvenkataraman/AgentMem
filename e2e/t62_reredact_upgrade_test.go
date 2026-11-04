package e2e

import (
	"path/filepath"
	"testing"
)

func TestT62_OldIndexReredactedInPlace(t *testing.T) {
	st := t.TempDir()
	f := filepath.Join(t.TempDir(), "r.jsonl")
	writeFile(t, f, `{"type":"user","uuid":"u0","sessionId":"s62","timestamp":"2026-10-01T10:00:00.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":"first note"}}`+"\n"+
		`{"type":"user","uuid":"u1","sessionId":"s62","timestamp":"2026-10-01T10:00:01.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":"second note"}}`+"\n")
	ok(t, st, nil, "ingest", "--file", f, "--store", "claude")

	db := openIdx(t, st)
	before := count(t, db, "SELECT count(*) FROM chunks")
	if _, err := db.Exec("UPDATE chunks SET text='key AKIAIOSFODNN7EXAMPLE leaked' WHERE text LIKE 'first%'"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("PRAGMA user_version=1"); err != nil {
		t.Fatal(err)
	}
	db.Close()

	so, _, _ := run(t, st, "audit")
	has(t, so, "leaks=0 ")

	db = openIdx(t, st)
	eq(t, "rows kept", count(t, db, "SELECT count(*) FROM chunks"), before)
	eq(t, "version", count(t, db, "PRAGMA user_version"), 2)
	eq(t, "fts clean", count(t, db, "SELECT count(*) FROM chunks_trgm WHERE chunks_trgm MATCH '\"AKIAIOSFODNN7\"'"), 0)
}

func TestT62_NewerVersionStillRebuilds(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "stats")
	db := openIdx(t, st)
	if _, err := db.Exec("PRAGMA user_version=9"); err != nil {
		t.Fatal(err)
	}
	db.Close()
	_, se, _ := run(t, st, "stats")
	has(t, se, "rebuilding")
}
