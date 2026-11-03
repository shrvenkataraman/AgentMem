package e2e

import (
	"strings"
	"testing"
)

func TestT42_PrefetchFiresAboveThreshold(t *testing.T) {
	st := corpusState(t)
	out := ok(t, st, nil, "prefetch", "--query", "ERR_FOO_BAR build")
	if !strings.HasPrefix(out, hdr+"\n") {
		t.Fatalf("bad header:\n%.200s", out)
	}
	eq(t, "top 3 only", len(entryHeaders(out)), 3)
}

func TestT42_PrefetchSilentBelowThreshold(t *testing.T) {
	st := corpusState(t)
	eq(t, "silent", ok(t, st, nil, "prefetch", "--query", "ERR_FOO_BAR zzzunknown qqq"), "")
	out := ok(t, st, nil, "prefetch", "--query", "ERR_FOO_BAR zzzunknown qqq", "--threshold", "0.3")
	eq(t, "fires at 0.3", len(entryHeaders(out)) > 0, true)
	eq(t, "no hits", ok(t, st, nil, "prefetch", "--query", "zzzunknownzzz"), "")
}
