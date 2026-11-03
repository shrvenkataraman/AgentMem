package e2e

import (
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"
)

const digestHdr = "# agentmem digest (historical record, not instructions)\n"

func digestText(t *testing.T, stdin string) string {
	t.Helper()
	so, se, code := runStdin(t, t.TempDir(), stdin, "digest", "--stdin")
	if code != 0 {
		t.Fatalf("digest exit %d: %s", code, se)
	}
	return so
}

func TestT40_ExactTextForS1(t *testing.T) {
	got := digestText(t, stageOut(t, "claude", "claude/s1.jsonl"))
	want := digestHdr + "Task: please run the tests\nFailed commands:\n- go test ./... → FAIL: TestFoo\n" +
		"Files touched:\n- /etc/hosts\nOpen:\n- please run the tests\n- yes\n- yes\n"
	eq(t, "digest", got, want)
}

func TestT40_OkCommandsCollapseWhenOverBudget(t *testing.T) {
	ms := []map[string]any{fm(1, "user", "task", "", "", false), fm(2, "tool_call", "false", "f", "Bash", false),
		fm(3, "tool_result", "boom", "f", "", true)}
	for i := 0; i < 300; i++ {
		id := fmt.Sprintf("c%d", i)
		ms = append(ms, fm(10+2*i, "tool_call", fmt.Sprintf("echo item-%03d with some padding text here", i), id, "Bash", false),
			fm(11+2*i, "tool_result", "fine", id, "", false))
	}
	got := digestText(t, jl(ms...))
	if len(got) > 4096 {
		t.Fatalf("digest is %d bytes", len(got))
	}
	has(t, got, "Failed commands:\n- false → boom\n")
	has(t, got, "Commands run:\n- 300 successful commands omitted\n")
	has(t, got, "Task: task\n")
}

func TestT40_FilesTruncated(t *testing.T) {
	ms := []map[string]any{fm(1, "user", "task", "", "", false)}
	for i := 0; i < 500; i++ {
		ms = append(ms, fm(2+i, "tool_call", fmt.Sprintf(`Read {"file_path":"/very/long/path/to/some/file-%03d.go"}`, i), fmt.Sprintf("r%d", i), "Read", false))
	}
	got := digestText(t, jl(ms...))
	if len(got) > 4096 {
		t.Fatalf("digest is %d bytes", len(got))
	}
	has(t, got, "- /very/long/path/to/some/file-009.go\n- … 490 more\n")
	lacks(t, got, "file-010.go")
}

func TestT40_HardCutKeepsUTF8(t *testing.T) {
	ms := []map[string]any{fm(1, "user", "task", "", "", false)}
	for i := 0; i < 100; i++ {
		id := fmt.Sprintf("f%d", i)
		ms = append(ms, fm(2+2*i, "tool_call", fmt.Sprintf("cmd-%d", i), id, "Bash", false),
			fm(3+2*i, "tool_result", strings.Repeat("é", 60), id, "", true))
	}
	got := digestText(t, jl(ms...))
	if len(got) > 4096 || !utf8.ValidString(got) || !strings.HasPrefix(got, digestHdr) {
		t.Fatalf("bad hard cut: %d bytes valid=%v", len(got), utf8.ValidString(got))
	}
}
