package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT68_UnreadableSkipListFailsClosed(t *testing.T) {
	st := t.TempDir()
	if err := os.MkdirAll(filepath.Join(st, "forgotten"), 0o700); err != nil {
		t.Fatal(err) // a directory makes ReadFile fail
	}
	_, errOut, code := run(t, st, "ingest", "--file", twoSessions(t, "s68a", "s68b", "/p68"), "--store", "claude")
	if code == 0 {
		t.Fatal("ingest must fail when the forgotten list is unreadable")
	}
	has(t, errOut, "forgotten list unreadable")
	_, errOut, code = run(t, st, "rebuild")
	if code == 0 {
		t.Fatal("rebuild must fail when the forgotten list is unreadable")
	}
	has(t, errOut, "forgotten list unreadable")
}
