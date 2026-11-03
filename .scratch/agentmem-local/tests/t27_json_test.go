package e2e

import "testing"

func TestT27_JSONSchema(t *testing.T) {
	st := corpusState(t)
	m := jsonLines(t, ok(t, st, nil, "ask", "--json", "deploy-prod"))[0]
	eq(t, "query", m["query"], "deploy-prod")
	eq(t, "stale", m["stale"], false)
	eq(t, "excluded", m["excluded_session"], "")
	r := asks(t, st, nil, "deploy-prod")
	eq(t, "count", len(r), 2)
	a, b := r[0], r[1]
	eq(t, "rank order: newest first on ties", a["role"], "tool_call")
	eq(t, "ts", a["ts"], "2026-10-02T09:00:06Z")
	eq(t, "agent", a["agent"], "claude")
	eq(t, "project", a["project"], "/home/u/Work/other")
	eq(t, "session", a["session"], "sc-b")
	eq(t, "seq", a["seq"], 300)
	eq(t, "snippet", a["snippet"], "make deploy-prod")
	eq(t, "provenance", a["provenance"], fixture("corpus/claude/-home-u-Work-other/sB.jsonl")+":3#0")
	eq(t, "coverage", a["coverage"], 1)
	eq(t, "seen_in", a["seen_in"], 1)
	if _, ok := a["id"].(interface{ Int64() (int64, error) }); !ok {
		t.Fatalf("id must be a JSON integer, got %T", a["id"])
	}
	eq(t, "second role", b["role"], "user")
	eq(t, "second provenance", b["provenance"], fixture("corpus/claude/-home-u-Work-other/sB.jsonl")+":1#0")
}

func TestT27_ExcludedSessionReported(t *testing.T) {
	st := corpusState(t)
	m := jsonLines(t, ok(t, st, []string{"AGENTMEM_CWD=/home/u/Work/proj"}, "ask", "--json", "ERR_FOO_BAR"))[0]
	eq(t, "excluded", m["excluded_session"], "sc-d")
	eq(t, "no results key missing", len(m["results"].([]any)), 5)
	m = jsonLines(t, ok(t, st, nil, "ask", "--json", "zzzunknownzzz"))[0]
	eq(t, "empty is [] not null", len(m["results"].([]any)), 0)
}
