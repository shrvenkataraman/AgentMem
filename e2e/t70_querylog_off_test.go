package e2e

import (
	"os"
	"testing"
)

func TestT70_QueryLogOff(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s70a", "s70b", "/p70"), "--store", "claude")
	ok(t, st, []string{"AGENTMEM_QUERYLOG=off"}, "ask", "s70a")
	if _, err := os.Stat(st + "/query-log.jsonl"); !os.IsNotExist(err) {
		t.Fatal("AGENTMEM_QUERYLOG=off must not write the log")
	}
	ok(t, st, nil, "ask", "s70a")
	if _, err := os.Stat(st + "/query-log.jsonl"); err != nil {
		t.Fatal("logging stays on by default")
	}
}
