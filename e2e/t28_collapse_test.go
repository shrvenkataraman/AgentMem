package e2e

import "testing"

func TestT28_CollapseIdenticalContent(t *testing.T) {
	st := t.TempDir()
	env := []string{"AGENTMEM_CLAUDE_DIR=" + fixture("dups/claude")}
	ok(t, st, env, "ingest")
	r := asks(t, st, nil, "ERR_DUP_X")
	eq(t, "collapsed", len(r), 1)
	eq(t, "seen_in", r[0]["seen_in"], 3)
	eq(t, "earliest kept", r[0]["session"], "dup-1")
	eq(t, "earliest ts", r[0]["ts"], "2026-10-01T09:00:02Z")
	r = asks(t, st, nil, "--no-dedup", "ERR_DUP_X")
	eq(t, "no-dedup", len(r), 3)
	eq(t, "no-dedup seen_in", r[0]["seen_in"], 1)
	has(t, ok(t, st, nil, "ask", "ERR_DUP_X"), " · tool_result · seen in 3 sessions\n")
}

func TestT28_DistinctTextNotCollapsed(t *testing.T) {
	eq(t, "corpus", len(asks(t, corpusState(t), nil, "ERR_FOO_BAR")), 6)
}
