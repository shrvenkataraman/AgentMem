package e2e

import "testing"

func TestT06_HermesToolRows(t *testing.T) {
	m := parsed(t, "hermes", "raw", "hermes/state.db")
	eq(t, "count", len(m), 8)
	refs := []string{"1#0", "2#0", "2#1", "3#0", "4#1", "5#0", "8#0", "9#0"}
	for i, r := range refs {
		eq(t, "ref", m[i]["source_ref"], r)
	}
	eq(t, "call role", m[2]["role"], "tool_call")
	eq(t, "call text", m[2]["text"], "make deploy")
	eq(t, "call id", m[2]["tool_id"], "c1")
	eq(t, "call name", m[2]["tool_name"], "terminal")
	eq(t, "res role", m[3]["role"], "tool_result")
	eq(t, "res text", m[3]["text"], `{"error": "make: *** No rule"}`)
	eq(t, "res id", m[3]["tool_id"], "c1")
	eq(t, "res name", m[3]["tool_name"], "terminal")
	eq(t, "call2 text", m[4]["text"], "agentmem ask foo")
	eq(t, "call2 seq", m[4]["seq"], 401)
}
