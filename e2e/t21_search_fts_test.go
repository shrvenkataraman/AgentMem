package e2e

import (
	"strings"
	"testing"
)

func TestT21_FTSCandidates(t *testing.T) {
	st := corpusState(t)
	l := rawLines(t, st, nil, "ERR_FOO_BAR")
	eq(t, "ERR_FOO_BAR hits", len(l), 6)
	eq(t, "sessions", strings.Join(sessionsOf(mapRaw(l)), ","), "sc-a,sc-b,sc-c,sc-d")
	l = rawLines(t, st, nil, "deploy-prod")
	eq(t, "deploy-prod hits", len(l), 2)
	eq(t, "deploy-prod session", strings.Join(rawField(l, 1), ","), "sc-b,sc-b")
	l = rawLines(t, st, nil, "--max-old-space-size=4096")
	eq(t, "flag token hits", len(l), 1)
	eq(t, "flag session", rawField(l, 1)[0], "sc-a")
	eq(t, "OR semantics", len(rawLines(t, st, nil, "deploy-prod", "noted")), 3)
	eq(t, "no match", len(rawLines(t, st, nil, "zzzunknownzzz")), 0)
}

func TestT21_QueryNeverBreaksSyntax(t *testing.T) {
	st := corpusState(t)
	for _, q := range []string{`AND OR NEAR "`, `a* (b) -c`, `"unterminated`} {
		ok(t, st, nil, "ask", "--raw", q)
	}
}

func mapRaw(lines []string) []map[string]any {
	var o []map[string]any
	for _, l := range lines {
		f := strings.Split(l, "\t")
		o = append(o, map[string]any{"session_id": f[1]})
	}
	return o
}
