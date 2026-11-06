package e2e

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestT18_ResultInLaterBatchIsDropped(t *testing.T) {
	L := fixLines(t, "claude/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s1.jsonl")
	st := t.TempDir()
	writeFile(t, f, strings.Join(L[:4], ""))
	eq(t, "call batch", ingestFile(t, st, "claude", f), "inserted=4\n")
	db := openIdx(t, st)
	eq(t, "key persisted", count(t, db, "SELECT count(*) FROM cursors WHERE store='poison' AND key=?", f), 1)
	appendFile(t, f, strings.Join(L[4:], ""))
	eq(t, "result batch", ingestFile(t, st, "claude", f), "inserted=3\n")
	eq(t, "rows", count(t, db, "SELECT count(*) FROM chunks"), 7)
	eq(t, "poison text absent", count(t, db, "SELECT count(*) FROM chunks WHERE text='poison result'"), 0)
}

func TestT18_PiResultInLaterBatchIsDropped(t *testing.T) {
	L := fixLines(t, "pi/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s1.jsonl")
	st := t.TempDir()
	writeFile(t, f, strings.Join(L[:7], ""))
	ingestFile(t, st, "pi", f)
	appendFile(t, f, strings.Join(L[7:], ""))
	ingestFile(t, st, "pi", f)
	db := openIdx(t, st)
	eq(t, "rows", count(t, db, "SELECT count(*) FROM chunks"), 6)
	eq(t, "poison text absent", count(t, db, "SELECT count(*) FROM chunks WHERE text='poison'"), 0)
}
