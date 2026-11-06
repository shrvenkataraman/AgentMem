package e2e

import (
	"os/exec"
	"strings"
	"testing"
)

func TestT71_AskRate(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", twoSessions(t, "s71a", "s71b", "/p71"), "--store", "claude")
	log := st + "/q.jsonl"
	writeFile(t, log, `{"ts":"2026-09-30T00:00:00Z","agent":"claude","session_id":"s71a","query":"x"}
{"ts":"2026-09-30T01:00:00Z","agent":"claude","session_id":"s71a","query":"y"}
{"ts":"2026-09-30T02:00:00Z","agent":"claude","query":"z"}
{"ts":"2026-09-30T03:00:00Z","agent":"unknown","session_id":"","query":"w"}
`)
	out, err := exec.Command("python3", "-I", "../scripts/askrate.py", log, st+"/index.sqlite").Output()
	if err != nil {
		t.Fatalf("script failed: %v", err)
	}
	eq(t, "stdout", strings.TrimSpace(string(out)),
		"agent=claude sessions=2 asking=1 asks=3\nagent=unknown sessions=0 asking=0 asks=1\nunattributed=2")
	if exec.Command("python3", "-I", "../scripts/askrate.py", log).Run() == nil {
		t.Fatal("one argument should exit non-zero")
	}
}
