package e2e

import (
	"strings"
	"testing"
)

func TestT26_ExcludeCurrentSession(t *testing.T) {
	st := corpusState(t)
	cwd := "AGENTMEM_CWD=/home/u/Work/proj"
	sess := func(env []string, args ...string) (int, string) {
		l := rawLines(t, st, env, append(args, "ERR_FOO_BAR")...)
		return len(l), strings.Join(sessionsOf(mapRaw(l)), ",")
	}
	n, s := sess(nil)
	eq(t, "no cwd match: nothing excluded", n, 6)
	n, s = sess([]string{cwd})
	eq(t, "newest in project excluded", n, 5)
	eq(t, "sessions", s, "sc-a,sc-b,sc-c")
	n, _ = sess([]string{cwd}, "--include-current")
	eq(t, "include-current", n, 6)
	n, s = sess(nil, "--session", "sc-a")
	eq(t, "--session", n, 3)
	eq(t, "--session sessions", s, "sc-b,sc-c,sc-d")
	n, s = sess([]string{"AGENTMEM_SESSION=sc-b"})
	eq(t, "AGENTMEM_SESSION", n, 5)
	n, s = sess([]string{"CLAUDE_CODE_SESSION_ID=sc-c"})
	eq(t, "CLAUDE_CODE_SESSION_ID", s, "sc-a,sc-b,sc-d")
	n, s = sess([]string{"AGENTMEM_SESSION=sc-a", "CLAUDE_CODE_SESSION_ID=sc-b"})
	eq(t, "AGENTMEM_SESSION beats CLAUDE_CODE_SESSION_ID", s, "sc-b,sc-c,sc-d")
	n, s = sess([]string{"AGENTMEM_SESSION=sc-b"}, "--session", "sc-a")
	eq(t, "flag beats env", s, "sc-b,sc-c,sc-d")
}
