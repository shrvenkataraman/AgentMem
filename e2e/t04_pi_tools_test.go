package e2e

import "testing"

func TestT04_PiToolBlocks(t *testing.T) {
	m := parsed(t, "pi", "raw", "pi/s1.jsonl")
	eq(t, "count", len(m), 8)
	refs := []string{"4#0", "5#1", "5#2", "6#0", "7#0", "8#0", "9#0", "10#0"}
	for i, r := range refs {
		eq(t, "ref", m[i]["source_ref"], r)
	}
	eq(t, "call role", m[2]["role"], "tool_call")
	eq(t, "call text", m[2]["text"], "ls")
	eq(t, "call id", m[2]["tool_id"], "call_1")
	eq(t, "call name", m[2]["tool_name"], "bash")
	eq(t, "res role", m[3]["role"], "tool_result")
	eq(t, "res text", m[3]["text"], "a.txt\nb.txt")
	eq(t, "res id", m[3]["tool_id"], "call_1")
	if _, has := m[3]["is_error"]; has {
		t.Fatal("non-error result must omit is_error")
	}
	eq(t, "err result", m[7]["is_error"], true)
	eq(t, "err text", m[7]["text"], "exit 1")
}
