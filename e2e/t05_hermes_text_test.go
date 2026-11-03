package e2e

import "testing"

func TestT05_HermesTextRows(t *testing.T) {
	out, se, code := runEnv(t, t.TempDir(), nil, "parse", "--store", "hermes", fixture("hermes/state.db"))
	eq(t, "exit", code, 0)
	has(t, se, "max_id=9")
	m := jsonLines(t, out)
	var tx []map[string]any
	for _, x := range m {
		if r := x["role"]; r == "user" || r == "assistant" {
			tx = append(tx, x)
		}
	}
	m = tx
	eq(t, "count", len(m), 4)
	refs := []string{"1#0", "2#0", "8#0", "9#0"}
	for i, r := range refs {
		eq(t, "ref", m[i]["source_ref"], r)
	}
	eq(t, "text1", m[0]["text"], "deploy the app")
	eq(t, "ts1", m[0]["ts"], 1790000000500)
	eq(t, "seq1", m[0]["seq"], 100)
	eq(t, "role1", m[0]["role"], "user")
	eq(t, "session1", m[0]["session_id"], "h1")
	eq(t, "agent1", m[0]["agent"], "hermes")
	eq(t, "store1", m[0]["source_store"], "hermes/cli")
	eq(t, "project1", m[0]["project"], "/home/u/Work/proj")
	eq(t, "path1", m[0]["source_path"], fixture("hermes/state.db"))
	eq(t, "compacted row kept", m[2]["text"], "compacted msg")
	eq(t, "subagent store", m[3]["source_store"], "hermes/subagent")
	eq(t, "subagent session", m[3]["session_id"], "h2")
}
