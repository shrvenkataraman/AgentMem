package e2e

import (
	"strings"
	"testing"
)

func digestJSON(t *testing.T, stdin string) map[string]any {
	t.Helper()
	so, se, code := runStdin(t, t.TempDir(), stdin, "digest", "--stdin", "--format", "json")
	if code != 0 {
		t.Fatalf("digest exit %d: %s", code, se)
	}
	return jsonLines(t, so)[0]
}

func stageOut(t *testing.T, store, rel string) string {
	return ok(t, t.TempDir(), nil, "parse", "--store", store, "--stage", "filtered", fixture(rel))
}

func TestT38_TaskAndCommandsClaude(t *testing.T) {
	d := digestJSON(t, stageOut(t, "claude", "claude/s1.jsonl"))
	eq(t, "task", d["task"], "please run the tests")
	c := d["commands"].([]any)
	eq(t, "commands", len(c), 1)
	c0 := c[0].(map[string]any)
	eq(t, "cmd", c0["cmd"], "go test ./...")
	eq(t, "outcome", c0["outcome"], "fail")
	eq(t, "output", c0["output"], "FAIL: TestFoo")
}

func TestT38_CommandsPi(t *testing.T) {
	d := digestJSON(t, stageOut(t, "pi", "pi/s1.jsonl"))
	eq(t, "task", d["task"], "list the files")
	c := d["commands"].([]any)
	eq(t, "count", len(c), 2)
	eq(t, "ls", c[0], map[string]any{"cmd": "ls", "outcome": "ok", "output": "a.txt"})
	eq(t, "false", c[1], map[string]any{"cmd": "false", "outcome": "fail", "output": "exit 1"})
}

func TestT38_UnknownOutcomeAndCaps(t *testing.T) {
	d := digestJSON(t, jl(
		fm(1, "user", strings.Repeat("q", 400), "", "", false),
		fm(2, "tool_call", "sleep 1", "a", "Bash", false),
		fm(3, "tool_call", "cat big", "b", "Bash", false),
		fm(4, "tool_result", strings.Repeat("z", 200)+"\nsecond line", "b", "", false),
	))
	eq(t, "task capped", len(d["task"].(string)), 300)
	c := d["commands"].([]any)
	eq(t, "unknown", c[0].(map[string]any)["outcome"], "unknown")
	eq(t, "unknown output", c[0].(map[string]any)["output"], "")
	eq(t, "ok", c[1].(map[string]any)["outcome"], "ok")
	eq(t, "output capped first line", len(c[1].(map[string]any)["output"].(string)), 120)
}

func TestT38_NoCommandsIsEmptyArray(t *testing.T) {
	d := digestJSON(t, jl(fm(1, "user", "hi", "", "", false)))
	eq(t, "commands", len(d["commands"].([]any)), 0)
}
