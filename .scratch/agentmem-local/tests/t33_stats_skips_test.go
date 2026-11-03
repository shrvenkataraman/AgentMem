package e2e

import (
	"path/filepath"
	"testing"
)

func skipped(t *testing.T, st string) map[string]any {
	t.Helper()
	m, _ := statsOf(t, st, nil)["skipped_lines"].(map[string]any)
	return m
}

func TestT33_SkippedLinesCounted(t *testing.T) {
	st := t.TempDir()
	ingestFile(t, st, "claude", fixture("claude/s1.jsonl"))
	eq(t, "claude", skipped(t, st)["claude"], 3)
	ingestFile(t, st, "claude", fixture("claude/s1.jsonl"))
	eq(t, "no double count", skipped(t, st)["claude"], 3)
	ingestFile(t, st, "pi", fixture("pi/s1.jsonl"))
	if _, present := skipped(t, st)["pi"]; present {
		t.Fatal("stores with zero skips must be omitted")
	}
}

func TestT33_SkipsAccumulateAcrossBatches(t *testing.T) {
	L := fixLines(t, "claude/s1.jsonl")
	f := filepath.Join(t.TempDir(), "s.jsonl")
	st := t.TempDir()
	writeFile(t, f, L[0]+"not json\n")
	ingestFile(t, st, "claude", f)
	eq(t, "first", skipped(t, st)["claude"], 1)
	appendFile(t, f, "also not json\n")
	ingestFile(t, st, "claude", f)
	eq(t, "second", skipped(t, st)["claude"], 2)
}
