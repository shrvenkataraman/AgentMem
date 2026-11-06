package e2e

import "testing"

func TestT02_ClaudeToolBlocks(t *testing.T) {
	m := parsed(t, "claude", "raw", "claude/s1.jsonl")
	eq(t, "count", len(m), 9)
	refs := []string{"1#0", "2#1", "2#2", "3#0", "4#0", "5#0", "8#0", "9#0", "11#0"}
	for i, r := range refs {
		eq(t, "ref", m[i]["source_ref"], r)
	}
	call := m[2]
	eq(t, "call role", call["role"], "tool_call")
	eq(t, "call text", call["text"], "go test ./...")
	eq(t, "call tool_id", call["tool_id"], "toolu_1")
	eq(t, "call tool_name", call["tool_name"], "Bash")
	res := m[3]
	eq(t, "res role", res["role"], "tool_result")
	eq(t, "res text", res["text"], "FAIL: TestFoo")
	eq(t, "res tool_id", res["tool_id"], "toolu_1")
	eq(t, "res is_error", res["is_error"], true)
	eq(t, "list-content result", m[5]["text"], "poison result")
	eq(t, "non-Bash call", m[8]["text"], `Read {"file_path":"/etc/hosts"}`)
	eq(t, "non-Bash name", m[8]["tool_name"], "Read")
}
