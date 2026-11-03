package e2e

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func ingestFile(t *testing.T, st, store, path string) string {
	t.Helper()
	return ok(t, st, nil, "ingest", "--file", path, "--store", store)
}

func TestT17_ResumeFromCursor(t *testing.T) {
	L := fixLines(t, "claude/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s1.jsonl")
	st := t.TempDir()
	writeFile(t, f, strings.Join(L[:3], ""))
	eq(t, "first batch", ingestFile(t, st, "claude", f), "inserted=4\n")
	appendFile(t, f, strings.Join(L[3:], ""))
	eq(t, "second batch", ingestFile(t, st, "claude", f), "inserted=3\n")
	eq(t, "nothing new", ingestFile(t, st, "claude", f), "inserted=0\n")
	db := openIdx(t, st)
	eq(t, "rows", count(t, db, "SELECT count(*) FROM chunks"), 7)
	size := len(strings.Join(L, ""))
	eq(t, "cursor", cursorOf(t, db, "claude", f), fmt.Sprintf("%d:12", size))
}

func TestT17_TornLastLineWaits(t *testing.T) {
	L := fixLines(t, "claude/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s1.jsonl")
	st := t.TempDir()
	writeFile(t, f, L[0]+L[7][:20])
	eq(t, "torn", ingestFile(t, st, "claude", f), "inserted=1\n")
	eq(t, "cursor", cursorOf(t, openIdx(t, st), "claude", f), fmt.Sprintf("%d:2", len(L[0])))
	writeFile(t, f, L[0]+L[7])
	eq(t, "completed", ingestFile(t, st, "claude", f), "inserted=1\n")
}

func TestT17_TruncationRestartsAtZero(t *testing.T) {
	L := fixLines(t, "claude/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s1.jsonl")
	st := t.TempDir()
	writeFile(t, f, strings.Join(L, ""))
	eq(t, "full", ingestFile(t, st, "claude", f), "inserted=7\n")
	writeFile(t, f, L[0])
	eq(t, "truncated", ingestFile(t, st, "claude", f), "inserted=0\n")
	eq(t, "rows kept", count(t, openIdx(t, st), "SELECT count(*) FROM chunks"), 7)
	eq(t, "cursor reset", cursorOf(t, openIdx(t, st), "claude", f), fmt.Sprintf("%d:2", len(L[0])))
}

func TestT17_PiResumeKeepsHeader(t *testing.T) {
	L := fixLines(t, "pi/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s1.jsonl")
	st := t.TempDir()
	writeFile(t, f, strings.Join(L[:5], ""))
	eq(t, "first", ingestFile(t, st, "pi", f), "inserted=3\n")
	appendFile(t, f, strings.Join(L[5:], ""))
	eq(t, "second", ingestFile(t, st, "pi", f), "inserted=3\n")
	db := openIdx(t, st)
	eq(t, "one session", count(t, db, "SELECT count(DISTINCT session_id) FROM chunks WHERE session_id='sess-pi-1'"), 1)
	eq(t, "all in it", count(t, db, "SELECT count(*) FROM chunks WHERE session_id!='sess-pi-1' OR project!='/home/u/Work/proj'"), 0)
}
