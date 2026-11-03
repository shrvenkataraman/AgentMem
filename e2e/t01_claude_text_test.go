package e2e

import "testing"

func TestT01_ClaudeTextBlocks(t *testing.T) {
	m := parsed(t, "claude", "raw", "claude/s1.jsonl")
	eq(t, "count", len(m), 4)
	first := m[0]
	eq(t, "source_ref", first["source_ref"], "1#0")
	eq(t, "seq", first["seq"], 100)
	eq(t, "ts", first["ts"], 1790848800000)
	eq(t, "role", first["role"], "user")
	eq(t, "text", first["text"], "please run the tests")
	eq(t, "session", first["session_id"], "sess-claude-1")
	eq(t, "agent", first["agent"], "claude")
	eq(t, "store", first["source_store"], "claude")
	eq(t, "project", first["project"], "/home/u/Work/proj")
	eq(t, "path", first["source_path"], fixture("claude/s1.jsonl"))
	// thinking block (idx 0) is skipped, the text block is idx 1
	eq(t, "ref2", m[1]["source_ref"], "2#1")
	eq(t, "seq2", m[1]["seq"], 201)
	eq(t, "role2", m[1]["role"], "assistant")
	eq(t, "text2", m[1]["text"], "Running tests now.")
	eq(t, "ref3", m[2]["source_ref"], "8#0")
	eq(t, "ref4", m[3]["source_ref"], "9#0")
	eq(t, "text4", m[3]["text"], "yes")
}
