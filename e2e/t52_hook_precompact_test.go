package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// hookTranscript writes a small claude transcript (one task, one failing command) and returns its path.
func hookTranscript(t *testing.T) string {
	t.Helper()
	lines := `{"type":"user","uuid":"u0","sessionId":"sess-1","timestamp":"2026-10-01T10:00:00.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":"fix the flaky login test"}}` + "\n" +
		`{"type":"assistant","uuid":"a1","sessionId":"sess-1","timestamp":"2026-10-01T10:00:01.000Z","cwd":"/p","isSidechain":false,"message":{"role":"assistant","content":[{"type":"tool_use","id":"k1","name":"Bash","input":{"command":"go test ./login"}}]}}` + "\n" +
		`{"type":"user","uuid":"u1","sessionId":"sess-1","timestamp":"2026-10-01T10:00:02.000Z","cwd":"/p","isSidechain":false,"message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"k1","is_error":true,"content":"FAIL TestLogin timeout"}]}}` + "\n"
	f := filepath.Join(t.TempDir(), "t.jsonl")
	writeFile(t, f, lines)
	return f
}

// hookPayload is the JSON Claude Code writes to a hook's stdin.
func hookPayload(sid, transcript, event string) string {
	return fmt.Sprintf(`{"session_id":%q,"transcript_path":%q,"hook_event_name":%q}`, sid, transcript, event)
}

func TestT52_PrecompactSpoolsDigest(t *testing.T) {
	st := t.TempDir()
	so, _, code := runStdin(t, st, hookPayload("sess-1", hookTranscript(t), "PreCompact"), "hook", "precompact")
	eq(t, "exit", code, 0)
	eq(t, "stdout", so, "")
	p := filepath.Join(st, "spool", "sess-1.md")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	has(t, string(b), "go test ./login")
	has(t, string(b), "fix the flaky login test")
	fi, _ := os.Stat(p)
	eq(t, "mode", fi.Mode().Perm(), os.FileMode(0o600))
}

func TestT52_PrecompactNeverFailsCompaction(t *testing.T) {
	st := t.TempDir()
	_, _, code := runStdin(t, st, "not json", "hook", "precompact")
	eq(t, "bad payload exits 0", code, 0)
	_, _, code = runStdin(t, st, hookPayload("sess-9", "/nonexistent/t.jsonl", "PreCompact"), "hook", "precompact")
	eq(t, "missing transcript exits 0", code, 0)
	_, err := os.Stat(filepath.Join(st, "spool", "sess-9.md"))
	eq(t, "no spool written", os.IsNotExist(err), true)
}
