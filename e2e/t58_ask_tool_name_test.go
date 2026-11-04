package e2e

import (
	"path/filepath"
	"testing"
)

func TestT58_AskShowsToolName(t *testing.T) {
	st := t.TempDir()
	f := filepath.Join(t.TempDir(), "w.jsonl")
	writeFile(t, f, `{"type":"assistant","uuid":"a1","sessionId":"s58","timestamp":"2026-10-01T10:00:01.000Z","cwd":"/p","isSidechain":false,"message":{"role":"assistant","content":[{"type":"tool_use","id":"k1","name":"WebFetch","input":{"url":"https://example.com/zqx58marker"}}]}}`+"\n"+
		`{"type":"user","uuid":"u1","sessionId":"s58","timestamp":"2026-10-01T10:00:02.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":"plain zqx58marker note"}}`+"\n")
	ok(t, st, nil, "ingest", "--file", f, "--store", "claude")

	text := ok(t, st, nil, "ask", "zqx58marker")
	has(t, text, " · tool_call:WebFetch\n")
	has(t, text, " · s58 · user\n")
	has(t, ok(t, st, nil, "ask", "--json", "zqx58marker"), `"tool_name":"WebFetch"`)
}
