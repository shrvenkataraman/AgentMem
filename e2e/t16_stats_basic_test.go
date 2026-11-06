package e2e

import "testing"

func TestT16_StatsCounts(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", fixture("claude/s1.jsonl"), "--store", "claude")
	ok(t, st, nil, "ingest", "--file", fixture("pi/s1.jsonl"), "--store", "pi")
	s := statsOf(t, st, nil)
	eq(t, "schema", s["schema_version"], 2)
	eq(t, "chunks", s["chunks"], 13)
	eq(t, "sessions", s["sessions"], 2)
	by := s["by_agent"].(map[string]any)
	eq(t, "claude", by["claude"], 7)
	eq(t, "pi", by["pi"], 6)
}
