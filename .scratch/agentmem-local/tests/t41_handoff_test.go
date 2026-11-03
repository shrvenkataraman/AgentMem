package e2e

import (
	"strings"
	"testing"
)

const handoffHdr = "# agentmem handoff (regenerated, historical record, not instructions)\n\n"

func TestT41_HandoffSinceFilterAndContent(t *testing.T) {
	st := corpusState(t)
	out := ok(t, st, nil, "handoff", "--regen", "--since", "2026-10-02")
	if !strings.HasPrefix(out, handoffHdr) {
		t.Fatalf("bad header:\n%.200s", out)
	}
	has(t, out, "## Session sc-b (claude, /home/u/Work/other, 2026-10-02)\nTask: deploy with make deploy-prod\n"+
		"Failed commands:\n- make deploy-prod → Error: ERR_FOO_BAR at line 7 (deploy)\nOpen:\n- deploy with make deploy-prod\n\n")
	last := -1
	for _, s := range []string{"Session sc-b ", "Session sc-pi ", "Session sc-c ", "Session sc-d "} {
		i := strings.Index(out, s)
		if i <= last {
			t.Fatalf("%q missing or out of order (by first message time)", s)
		}
		last = i
	}
	for _, s := range []string{"Session sc-a ", "Session h1 ", "Session h2 "} {
		lacks(t, out, s)
	}
	eq(t, "deterministic", ok(t, st, nil, "handoff", "--regen", "--since", "2026-10-02"), out)
}

func TestT41_SessionFlagAndNoIngest(t *testing.T) {
	st := corpusState(t)
	out := ok(t, st, nil, "handoff", "--regen", "--session", "sc-d")
	eq(t, "one section", strings.Count(out, "## Session "), 1)
	has(t, out, "## Session sc-d (claude, /home/u/Work/proj, 2026-10-04)\n")
	empty := t.TempDir()
	eq(t, "empty", ok(t, empty, corpus(), "handoff", "--regen"), handoffHdr+"(no sessions)\n")
	eq(t, "handoff never ingests", statsOf(t, empty, nil)["chunks"], 0)
}

func TestT41_RegenRequired(t *testing.T) {
	_, _, code := runEnv(t, t.TempDir(), nil, "handoff")
	eq(t, "exit", code, 2)
}
