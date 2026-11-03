package e2e

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestT34_QueryLogEntry(t *testing.T) {
	st := corpusState(t)
	ok(t, st, []string{"AGENTMEM_AGENT=claude", "AGENTMEM_CWD=/home/u/Work/proj"}, "ask", "--include-current", "ERR_FOO_BAR")
	ok(t, st, nil, "ask", "zzzunknownzzz")
	b, err := os.ReadFile(st + "/query-log.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	e := jsonLines(t, string(b))
	eq(t, "entries", len(e), 2)
	eq(t, "query", e[0]["query"], "ERR_FOO_BAR")
	eq(t, "agent", e[0]["agent"], "claude")
	eq(t, "project", e[0]["project"], "/home/u/Work/proj")
	eq(t, "hits", e[0]["hits"], 6)
	eq(t, "zero_result", e[0]["zero_result"], false)
	eq(t, "sources", e[0]["sources_returned"].([]any)[0], "claude")
	eq(t, "sources len", len(e[0]["sources_returned"].([]any)), 1)
	if _, err := time.Parse(time.RFC3339, e[0]["ts"].(string)); err != nil {
		t.Fatal(err)
	}
	eq(t, "default agent", e[1]["agent"], "unknown")
	eq(t, "zero hits", e[1]["hits"], 0)
	eq(t, "zero_result", e[1]["zero_result"], true)
	eq(t, "no sources", len(e[1]["sources_returned"].([]any)), 0)
	fi, _ := os.Stat(st + "/query-log.jsonl")
	eq(t, "mode", fi.Mode().Perm(), os.FileMode(0o600))
}

func TestT34_QueryLogRotates(t *testing.T) {
	st := corpusState(t)
	env := []string{"AGENTMEM_QUERYLOG_MAX=100"}
	for i := 0; i < 4; i++ {
		ok(t, st, env, "ask", "ERR_FOO_BAR")
	}
	if _, err := os.Stat(st + "/query-log.jsonl.1"); err != nil {
		t.Fatal("expected rotated file: ", err)
	}
	b, _ := os.ReadFile(st + "/query-log.jsonl")
	eq(t, "current file has the newest entry only", len(strings.Split(strings.TrimSpace(string(b)), "\n")), 1)
}

// TestT34_QueryLogHonoursLimit pins the review-gate fix: Hits and SourcesReturned describe the
// hits actually returned, so --limit truncates them, and a negative limit answers empty (exit 0)
// instead of panicking the render path.
func TestT34_QueryLogHonoursLimit(t *testing.T) {
	st := corpusState(t)
	ok(t, st, []string{"AGENTMEM_AGENT=claude"}, "ask", "--include-current", "--limit", "1", "ERR_FOO_BAR")
	b, err := os.ReadFile(st + "/query-log.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	e := jsonLines(t, string(b))
	eq(t, "entries", len(e), 1)
	eq(t, "hits follows --limit", e[0]["hits"], 1)
	eq(t, "sources len", len(e[0]["sources_returned"].([]any)), 1)
	_, errOut, rc := runEnv(t, st, nil, "ask", "--include-current", "--limit", "-1", "ERR_FOO_BAR")
	eq(t, "negative limit exit", rc, 0)
	lacks(t, errOut, "panic")
}
