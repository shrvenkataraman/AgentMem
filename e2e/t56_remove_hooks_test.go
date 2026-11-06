package e2e

import (
	"path/filepath"
	"testing"
)

func TestT56_RemoveHooksKeepsOthers(t *testing.T) {
	st := t.TempDir()
	p := filepath.Join(t.TempDir(), "settings.json")
	writeFile(t, p, `{"hooks":{"PreCompact":[{"matcher":"","hooks":[{"type":"command","command":"other"}]}]}}`)
	ok(t, st, nil, "install-hooks", "--settings", p, "--bin", "/x/agentmem")
	ok(t, st, nil, "install-hooks", "--settings", p, "--bin", "/x/agentmem", "--remove")
	eq(t, "PreCompact", hookCommands(t, p, "PreCompact"), []string{"|other"})
	eq(t, "SessionStart", len(hookCommands(t, p, "SessionStart")), 0)
	eq(t, "UserPromptSubmit", len(hookCommands(t, p, "UserPromptSubmit")), 0)
}
