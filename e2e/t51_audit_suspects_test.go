package e2e

import "testing"

func TestT51_SuspectsNeedMixedCaseAndDigit(t *testing.T) {
	st := corpusState(t)
	db := openIdx(t, st)
	for i, text := range []string{
		"slug abcdefghijklmnopqrstuvwxyz-abcdefghijklmnop end",
		"id AbcdefghIjklmnopQrstuvwxYzabcdefghIjklmno end",
		"path /home/user/projects/agent-memory/internal/ingest/x end",
	} {
		if _, err := db.Exec(insChunk, "noise", string(rune('a'+i))+"#0", text); err != nil {
			t.Fatal(err)
		}
	}
	eq(t, "identifiers and paths are not suspects", ok(t, st, nil, "audit"), "leaks=0 suspects=0\n")
}
