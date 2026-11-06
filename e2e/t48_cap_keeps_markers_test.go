package e2e

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// capped runs one tool_result of the given text through `parse --stage chunks` and returns its chunk text.
func capped(t *testing.T, result string) string {
	t.Helper()
	c, _ := json.Marshal(result)
	lines := `{"type":"assistant","uuid":"a1","sessionId":"s","timestamp":"2026-10-01T10:00:00.000Z","cwd":"/p","isSidechain":false,"message":{"role":"assistant","content":[{"type":"tool_use","id":"k1","name":"Bash","input":{"command":"env"}}]}}` + "\n" +
		fmt.Sprintf(`{"type":"user","uuid":"u1","sessionId":"s","timestamp":"2026-10-01T10:00:01.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"k1","content":%s}]}}`, c) + "\n"
	f := filepath.Join(t.TempDir(), "c.jsonl")
	writeFile(t, f, lines)
	for _, m := range jsonLines(t, ok(t, t.TempDir(), nil, "parse", "--store", "claude", "--stage", "chunks", f)) {
		if m["role"] == "tool_result" {
			return m["text"].(string)
		}
	}
	t.Fatal("no tool_result chunk")
	return ""
}

func TestT48_CapNeverSplitsARedactionMarker(t *testing.T) {
	mark := "\n[...truncated...]\n"
	// The 512-byte head cut would land inside the marker that API_TOKEN=... becomes.
	head := capped(t, strings.Repeat("x", 495)+" API_TOKEN=supersecretvalue "+strings.Repeat("y", 300))
	eq(t, "head cut backs up to the marker start", head, strings.Repeat("x", 495)+" API_TOKEN="+mark+strings.Repeat("y", 128))
	// The tail start (len-128) would land inside the marker.
	tail := capped(t, strings.Repeat("x", 600)+" API_TOKEN=supersecretvalue "+strings.Repeat("z", 120))
	eq(t, "tail start moves past the marker", tail, strings.Repeat("x", 512)+mark+" "+strings.Repeat("z", 120))
}
