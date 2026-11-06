package e2e

import (
	"fmt"
	"path/filepath"
	"testing"
)

// twoSessions writes a claude transcript with one user message each for sessions a and b, under cwd.
func twoSessions(t *testing.T, a, b, cwd string) string {
	t.Helper()
	line := `{"type":"user","uuid":"%s","sessionId":"%s","timestamp":"2026-10-01T10:00:0%d.000Z","cwd":%q,"isSidechain":false,"message":{"role":"user","content":"note from %s"}}` + "\n"
	f := filepath.Join(t.TempDir(), "two.jsonl")
	writeFile(t, f, fmt.Sprintf(line, "u1", a, 1, cwd, a)+fmt.Sprintf(line, "u2", b, 2, cwd, b))
	return f
}

func TestT63_ForgottenSessionSkipped(t *testing.T) {
	st := t.TempDir()
	writeFile(t, filepath.Join(st, "forgotten"), "session\ts63a\n")
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s63a", "s63b", "/p63"), "--store", "claude")
	db := openIdx(t, st)
	eq(t, "forgotten", count(t, db, "SELECT count(*) FROM chunks WHERE session_id='s63a'"), 0)
	if count(t, db, "SELECT count(*) FROM chunks WHERE session_id='s63b'") == 0 {
		t.Fatal("s63b should be indexed")
	}
}

func TestT63_ForgottenProjectSkipped(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s63c", "s63d", "/p63"), "--store", "claude")
	db := openIdx(t, st)
	var proj string
	if err := db.QueryRow("SELECT project FROM chunks LIMIT 1").Scan(&proj); err != nil {
		t.Fatal(err)
	}
	db.Close()

	st2 := t.TempDir()
	writeFile(t, filepath.Join(st2, "forgotten"), "project\t"+proj+"\n")
	ok(t, st2, nil, "ingest", "--file", twoSessions(t, "s63c", "s63d", "/p63"), "--store", "claude")
	eq(t, "project dropped", count(t, openIdx(t, st2), "SELECT count(*) FROM chunks"), 0)
}
