package e2e

import (
	"strings"
	"testing"
)

func strs(v any) string {
	var o []string
	for _, x := range v.([]any) {
		o = append(o, x.(string))
	}
	return strings.Join(o, "|")
}

func TestT39_FilesAndOpenFromS1(t *testing.T) {
	d := digestJSON(t, stageOut(t, "claude", "claude/s1.jsonl"))
	eq(t, "files", strs(d["files"]), "/etc/hosts")
	eq(t, "open", strs(d["open"]), "please run the tests|yes|yes")
}

func TestT39_FilesDedupAndToolNames(t *testing.T) {
	d := digestJSON(t, jl(
		fm(1, "user", "u1", "", "", false), fm(2, "user", "u2", "", "", false),
		fm(3, "user", "u3", "", "", false), fm(4, "user", strings.Repeat("w", 300), "", "", false),
		fm(5, "user", "u5", "", "", false),
		fm(6, "tool_call", `edit {"path":"/a/b.go","edits":[]}`, "e", "edit", false),
		fm(7, "tool_call", `Write {"file_path":"/a/c.go","content":"x"}`, "w", "Write", false),
		fm(8, "tool_call", `Read {"file_path":"/a/b.go"}`, "r", "Read", false),
		fm(9, "tool_call", `Grep {"pattern":"x","path":"/a"}`, "g", "Grep", false),
	))
	eq(t, "files", strs(d["files"]), "/a/b.go|/a/c.go")
	open := d["open"].([]any)
	eq(t, "open count", len(open), 3)
	eq(t, "open capped", len(open[1].(string)), 200)
	eq(t, "open last", open[2], "u5")
}
