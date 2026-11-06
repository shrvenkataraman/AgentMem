package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

const rulesBlock = "<!-- agentmem:begin -->\nBefore re-deriving a decision, command, or error message, run `agentmem ask '<query>'`.\n<!-- agentmem:end -->\n"

func read(t *testing.T, p string) string {
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestT45_InstallIsIdempotentAndRemovable(t *testing.T) {
	st := corpusState(t)
	f := filepath.Join(t.TempDir(), "CLAUDE.md")
	orig := "# my rules\n\nkeep this\n"
	writeFile(t, f, orig)
	ok(t, st, corpus(), "install-rules", "--file", f)
	eq(t, "installed", read(t, f), orig+"\n"+rulesBlock)
	ok(t, st, corpus(), "install-rules", "--file", f)
	eq(t, "idempotent", read(t, f), orig+"\n"+rulesBlock)
	writeFile(t, f, orig+"\n<!-- agentmem:begin -->\nold text\n<!-- agentmem:end -->\n")
	ok(t, st, corpus(), "install-rules", "--file", f)
	eq(t, "replaced in place", read(t, f), orig+"\n"+rulesBlock)
	ok(t, st, nil, "install-rules", "--file", f, "--remove")
	eq(t, "removed", read(t, f), orig)
}

func TestT45_CreatesMissingFileAndRemoveIsSafe(t *testing.T) {
	st := corpusState(t)
	f := filepath.Join(t.TempDir(), "sub", "AGENTS.md")
	ok(t, st, nil, "install-rules", "--file", f, "--remove")
	if exists(f) {
		t.Fatal("--remove must not create the file")
	}
	ok(t, st, corpus(), "install-rules", "--file", f)
	eq(t, "created", read(t, f), rulesBlock)
}

func TestT45_RefusesWhenIndexEmpty(t *testing.T) {
	f := filepath.Join(t.TempDir(), "CLAUDE.md")
	_, se, code := runEnv(t, t.TempDir(), nil, "install-rules", "--file", f)
	eq(t, "exit", code, 1)
	has(t, se, "index is empty")
	if exists(f) {
		t.Fatal("must not write when refusing")
	}
}
