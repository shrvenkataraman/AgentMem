package e2e

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// redactedTexts feeds each string through `parse --stage redacted` as a claude user line.
func redactedTexts(t *testing.T, in ...string) []string {
	t.Helper()
	var sb strings.Builder
	for i, s := range in {
		c, _ := json.Marshal(s)
		fmt.Fprintf(&sb, `{"type":"user","uuid":"u%d","sessionId":"s","timestamp":"2026-10-01T10:00:%02d.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":%s}}`+"\n", i, i, c)
	}
	f := filepath.Join(t.TempDir(), "r.jsonl")
	writeFile(t, f, sb.String())
	return texts(jsonLines(t, ok(t, t.TempDir(), nil, "parse", "--store", "claude", "--stage", "redacted", f)))
}

func TestT46_AssignmentKindsAndIdempotency(t *testing.T) {
	got := redactedTexts(t,
		"DB_PASSWORD=hunter2 go",
		"export MY_SECRET=s3cr3t",
		"PASSWD=zz",
		"openai_api_key=sk-abc",
		"PRIVATEKEY=k1",
		"API_KEY=[REDACTED:secret]",
		"GH_TOKEN=AKIA0123456789ABCDEF",
		"HOME=/root PATH=/usr/bin",
	)
	want := []string{
		"DB_PASSWORD=[REDACTED:secret] go",
		"export MY_SECRET=[REDACTED:secret]",
		"PASSWD=[REDACTED:secret]",
		"openai_api_key=[REDACTED:secret]",
		"PRIVATEKEY=[REDACTED:secret]",
		"API_KEY=[REDACTED:secret]",
		"GH_TOKEN=[REDACTED:secret]",
		"HOME=/root PATH=/usr/bin",
	}
	if len(got) != len(want) {
		t.Fatalf("messages: got %d, want %d", len(got), len(want))
	}
	for i := range want {
		eq(t, fmt.Sprintf("case %d", i), got[i], want[i])
	}
}
