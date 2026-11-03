package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
)

func TestT11_ToolResultCaps(t *testing.T) {
	c := parsed(t, "claude", "chunks", "caps/results.jsonl")
	eq(t, "count", len(c), 4)
	mark := "\n[...truncated...]\n"
	ok1 := c[1]["text"].(string)
	eq(t, "success len", len(ok1), 659)
	eq(t, "success text", ok1, strings.Repeat("A", 512)+mark+strings.Repeat("B", 128))
	bad := c[3]["text"].(string)
	eq(t, "error len", len(bad), 2531)
	eq(t, "error text", bad, strings.Repeat("C", 2000)+mark+strings.Repeat("D", 512))
	eq(t, "error windows", c[3]["window_idx"], 0)
	sum := sha256.Sum256([]byte(ok1))
	eq(t, "hash of capped text", c[1]["content_hash"], hex.EncodeToString(sum[:]))
	eq(t, "tool_call untouched", c[0]["text"], "echo ok")
}

func TestT11_ShortResultUncapped(t *testing.T) {
	c := parsed(t, "claude", "chunks", "claude/s1.jsonl")
	// chunk 2 is the tool_call; its tool_result is chunk 3 (13 bytes, under the 640 threshold).
	eq(t, "uncapped role", c[3]["role"], "tool_result")
	eq(t, "uncapped text", c[3]["text"], "FAIL: TestFoo")
}
