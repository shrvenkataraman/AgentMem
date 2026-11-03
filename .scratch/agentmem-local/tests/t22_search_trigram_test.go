package e2e

import (
	"strings"
	"testing"
)

func TestT22_TrigramSubstring(t *testing.T) {
	st := corpusState(t)
	l := rawLines(t, st, nil, "--max-old-space-size")
	eq(t, "flag prefix hits", len(l), 2)
	eq(t, "flag prefix sessions", strings.Join(sessionsOf(mapRaw(l)), ","), "sc-a,sc-pi")
	eq(t, "substring FOO_B", len(rawLines(t, st, nil, "FOO_B")), 6)
	eq(t, "no duplicates after union", len(rawLines(t, st, nil, "ERR_FOO_BAR")), 6)
	eq(t, "short term uses fts only", len(rawLines(t, st, nil, "pi")), 1)
}
