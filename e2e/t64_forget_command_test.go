package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT64_ForgetSession(t *testing.T) {
	st := t.TempDir()
	f := twoSessions(t, "s64a", "s64b", "/p64")
	ok(t, st, nil, "ingest", "--file", f, "--store", "claude")
	spool := filepath.Join(st, "spool", "s64a.md")
	if err := os.MkdirAll(filepath.Dir(spool), 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, spool, "digest")

	has(t, ok(t, st, nil, "forget", "--dry-run", "--session", "s64a"), "would-forget=")
	db := openIdx(t, st)
	if count(t, db, "SELECT count(*) FROM chunks WHERE session_id='s64a'") == 0 {
		t.Fatal("dry run must not delete")
	}
	db.Close()

	has(t, ok(t, st, nil, "forget", "--session", "s64a"), "forgot=")
	db = openIdx(t, st)
	eq(t, "rows", count(t, db, "SELECT count(*) FROM chunks WHERE session_id='s64a'"), 0)
	eq(t, "trigram", count(t, db, `SELECT count(*) FROM chunks_trgm WHERE chunks_trgm MATCH '"from s64a"'`), 0)
	if count(t, db, "SELECT count(*) FROM chunks WHERE session_id='s64b'") == 0 {
		t.Fatal("other session must stay")
	}
	db.Close()
	if _, err := os.Stat(spool); !os.IsNotExist(err) {
		t.Fatal("spool file should be removed")
	}
	has(t, ok(t, st, nil, "ask", "s64a"), "(no results)")

	ok(t, st, nil, "ingest", "--file", f, "--store", "claude")
	eq(t, "reingest", count(t, openIdx(t, st), "SELECT count(*) FROM chunks WHERE session_id='s64a'"), 0)
}

func TestT64_ForgetUsage(t *testing.T) {
	_, _, code := run(t, t.TempDir(), "forget")
	eq(t, "exit", code, 2)
}
