package e2e

import "testing"

func TestT03_PiTextBlocks(t *testing.T) {
	m := parsed(t, "pi", "raw", "pi/s1.jsonl")
	eq(t, "count", len(m), 2)
	eq(t, "ref1", m[0]["source_ref"], "4#0")
	eq(t, "seq1", m[0]["seq"], 400)
	eq(t, "ts1", m[0]["ts"], 1790848801000)
	eq(t, "role1", m[0]["role"], "user")
	eq(t, "text1", m[0]["text"], "list the files")
	eq(t, "session", m[0]["session_id"], "sess-pi-1")
	eq(t, "agent", m[0]["agent"], "pi")
	eq(t, "store", m[0]["source_store"], "pi")
	eq(t, "project", m[0]["project"], "/home/u/Work/proj")
	eq(t, "ref2", m[1]["source_ref"], "5#1")
	eq(t, "text2", m[1]["text"], "Listing.")
	eq(t, "role2", m[1]["role"], "assistant")
}
