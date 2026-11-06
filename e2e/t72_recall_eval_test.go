package e2e

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestT72_RecallEval(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s72a", "s72b", "/p72"), "--store", "claude")
	q := st + "/queries.tsv"
	writeFile(t, q, "# class\tsession\tquery\nidentifier\ts72a\ts72a\n\nidentifier\ts72b\tzzznothing72\nsemantic\ts72b\tnote s72b\n")
	cmd := exec.Command("python3", "-I", "../scripts/recall_eval.py", q)
	cmd.Env = append(os.Environ(), "AGENTMEM_BIN="+bin(t), "AGENTMEM_STATE_DIR="+st,
		"AGENTMEM_CLAUDE_DIR="+st+"/none-claude", "AGENTMEM_PI_DIR="+st+"/none-pi",
		"AGENTMEM_HERMES_DB="+st+"/none-hermes.db", "AGENTMEM_SESSION=", "CLAUDE_CODE_SESSION_ID=",
		"AGENTMEM_AGENT=", "AGENTMEM_CWD=/nonexistent/cwd")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("script failed: %v", err)
	}
	eq(t, "stdout", strings.TrimSpace(string(out)),
		"miss line=4\nclass=identifier n=2 hit=1 recall@5=0.50\nclass=semantic n=1 hit=1 recall@5=1.00\nclass=all n=3 hit=2 recall@5=0.67")
	if _, err := os.Stat(st + "/query-log.jsonl"); !os.IsNotExist(err) {
		t.Fatal("eval runs must not write the query log")
	}
}
