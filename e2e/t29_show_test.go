package e2e

import (
	"fmt"
	"testing"
)

func TestT29_ShowAndContext(t *testing.T) {
	st := corpusState(t)
	r := asks(t, st, nil, "--max-old-space-size=4096")
	eq(t, "one hit", len(r), 1)
	id := fmt.Sprint(r[0]["id"])
	out := ok(t, st, nil, "show", id)
	has(t, out, "## "+id+" · 2026-10-01T09:00:05Z · claude · /home/u/Work/proj · sc-a · assistant\n")
	has(t, out, "source: "+fixture("corpus/claude/-home-u-Work-proj/sA.jsonl")+":2#0\n")
	has(t, out, "Set --max-old-space-size=4096 to avoid ERR_FOO_BAR")
	lacks(t, out, "how do I fix")
	out = ok(t, st, nil, "show", "--context", "1", id)
	has(t, out, "▶ ## "+id+" ·")
	has(t, out, "how do I fix ERR_FOO_BAR in the build")
	has(t, out, "npm run build")
	lacks(t, out, "Error: ERR_FOO_BAR at line 3")
}

func TestT29_ShowErrors(t *testing.T) {
	st := corpusState(t)
	_, se, code := runEnv(t, st, nil, "show", "999999")
	eq(t, "missing exit", code, 1)
	has(t, se, "no such chunk")
	_, _, code = runEnv(t, st, nil, "show", "abc")
	eq(t, "bad id exit", code, 2)
}
