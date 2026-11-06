package e2e

import (
	"strings"
	"testing"
)

func TestT25_Filters(t *testing.T) {
	st := corpusState(t)
	sess := func(args ...string) string {
		return strings.Join(sessionsOf(mapRaw(rawLines(t, st, nil, args...))), ",")
	}
	eq(t, "project", sess("--project", "/home/u/Work/other", "ERR_FOO_BAR"), "sc-b")
	eq(t, "agent pi", sess("--agent", "pi", "flag"), "sc-pi")
	eq(t, "agent claude", sess("--agent", "claude", "flag"), "sc-c")
	eq(t, "since", sess("--since", "2026-10-03", "ERR_FOO_BAR"), "sc-c,sc-d")
	eq(t, "combined count", len(rawLines(t, st, nil, "--project", "/home/u/Work/proj", "--agent", "claude", "ERR_FOO_BAR")), 5)
}

func TestT25_BadSinceIsUsageError(t *testing.T) {
	_, se, code := runEnv(t, t.TempDir(), nil, "ask", "--since", "bogus", "x")
	eq(t, "exit", code, 2)
	has(t, se, "--since")
}
