package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT65_RejectsTraversalSession(t *testing.T) {
	st := t.TempDir()
	outside := filepath.Join(st, "keep.md")
	writeFile(t, outside, "keep")
	for _, id := range []string{"../keep", "a/b", `a\b`, "a\tb", "a\nb"} {
		_, _, code := run(t, st, "forget", "--session", id)
		eq(t, "exit for "+id, code, 2)
	}
	if _, err := os.Stat(outside); err != nil {
		t.Fatal("file outside the spool must survive")
	}
	if _, err := os.Stat(filepath.Join(st, "forgotten")); !os.IsNotExist(err) {
		t.Fatal("no skip-list entry for a rejected id")
	}
}

func TestT65_RejectsNewlineProject(t *testing.T) {
	st := t.TempDir()
	_, _, code := run(t, st, "forget", "--project", "/p\nsession\tx")
	eq(t, "exit", code, 2)
}

func TestT65_ZeroMatchWarns(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s65a", "s65b", "/p65"), "--store", "claude")
	out, errOut, code := run(t, st, "forget", "--project", "/nowhere65")
	eq(t, "exit", code, 0)
	has(t, out, "forgot=0")
	has(t, errOut, "no chunks matched")
}
