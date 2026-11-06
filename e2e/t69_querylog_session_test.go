package e2e

import (
	"os"
	"testing"
)

func TestT69_QueryLogSession(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s69a", "s69b", "/p69"), "--store", "claude")
	ok(t, st, []string{"CLAUDE_CODE_SESSION_ID=c69"}, "ask", "s69a")
	ok(t, st, []string{"AGENTMEM_AGENT=pi", "AGENTMEM_SESSION=p69", "CLAUDE_CODE_SESSION_ID=c69"}, "ask", "s69a")
	ok(t, st, nil, "ask", "s69a")
	b, err := os.ReadFile(st + "/query-log.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	e := jsonLines(t, string(b))
	eq(t, "entries", len(e), 3)
	eq(t, "claude agent", e[0]["agent"], "claude")
	eq(t, "claude session", e[0]["session_id"], "c69")
	eq(t, "pi agent", e[1]["agent"], "pi")
	eq(t, "pi session", e[1]["session_id"], "p69")
	eq(t, "unknown agent", e[2]["agent"], "unknown")
	eq(t, "no session", e[2]["session_id"], "")
}
