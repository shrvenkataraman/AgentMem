package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestT10_WindowSplit(t *testing.T) {
	c := parsed(t, "claude", "chunks", "window/long.jsonl")
	eq(t, "count", len(c), 5)
	wantLen := []int{2400, 1199, 2560, 2560, 880}
	wantIdx := []int{0, 1, 0, 1, 2}
	wantRef := []string{"1#0", "1#0", "2#0", "2#0", "2#0"}
	for i := range c {
		eq(t, "len", len(c[i]["text"].(string)), wantLen[i])
		eq(t, "window_idx", c[i]["window_idx"], wantIdx[i])
		eq(t, "ref", c[i]["source_ref"], wantRef[i])
	}
	sum := sha256.Sum256([]byte(c[0]["text"].(string)))
	eq(t, "hash", c[0]["content_hash"], hex.EncodeToString(sum[:]))
}

func TestT10_ShortTextIsOneWindow(t *testing.T) {
	c := parsed(t, "claude", "chunks", "claude/s1.jsonl")
	eq(t, "count", len(c), 7)
	for _, x := range c {
		eq(t, "window_idx", x["window_idx"], 0)
	}
}
