package e2e

import "testing"

func TestT19_DiscoverClaudeAndPi(t *testing.T) {
	env := []string{"AGENTMEM_CLAUDE_DIR=" + fixture("corpus/claude"), "AGENTMEM_PI_DIR=" + fixture("corpus/pi")}
	st := t.TempDir()
	eq(t, "first", ok(t, st, env, "ingest"), "files=5 inserted=15 skipped=0\n")
	eq(t, "second", ok(t, st, env, "ingest"), "files=5 inserted=0 skipped=0\n")
	by := statsOf(t, st, nil)["by_agent"].(map[string]any)
	eq(t, "claude", by["claude"], 13)
	eq(t, "pi", by["pi"], 2)
}

func TestT19_MissingRootsAreIgnored(t *testing.T) {
	eq(t, "none", ok(t, t.TempDir(), nil, "ingest"), "files=0 inserted=0 skipped=0\n")
}

func TestT19_DupsFixture(t *testing.T) {
	env := []string{"AGENTMEM_CLAUDE_DIR=" + fixture("dups/claude")}
	eq(t, "dups", ok(t, t.TempDir(), env, "ingest"), "files=3 inserted=9 skipped=0\n")
}
