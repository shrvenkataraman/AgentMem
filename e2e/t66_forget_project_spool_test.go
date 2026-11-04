package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestT66_ForgetProjectRemovesSpool(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s66a", "s66b", "/p66"), "--store", "claude")
	sp := filepath.Join(st, "spool")
	if err := os.MkdirAll(sp, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, s := range []string{"s66a", "s66b", "other66"} {
		writeFile(t, filepath.Join(sp, s+".md"), "digest")
	}
	has(t, ok(t, st, nil, "forget", "--project", "/p66"), "forgot=2")
	for _, s := range []string{"s66a", "s66b"} {
		if _, err := os.Stat(filepath.Join(sp, s+".md")); !os.IsNotExist(err) {
			t.Fatalf("spool %s should be removed", s)
		}
	}
	if _, err := os.Stat(filepath.Join(sp, "other66.md")); err != nil {
		t.Fatal("other spool must stay")
	}
}
