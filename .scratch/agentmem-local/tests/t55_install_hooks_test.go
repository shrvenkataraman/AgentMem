package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// hookCommands returns the command strings registered under event in a settings.json file.
func hookCommands(t *testing.T, path, event string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var s struct {
		Hooks map[string][]struct {
			Matcher string `json:"matcher"`
			Hooks   []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, g := range s.Hooks[event] {
		for _, h := range g.Hooks {
			out = append(out, g.Matcher+"|"+h.Command)
		}
	}
	return out
}

func TestT55_InstallHooksMergesAndIsIdempotent(t *testing.T) {
	st := t.TempDir()
	p := filepath.Join(t.TempDir(), "settings.json")
	orig := `{"model":"opus","hooks":{"PreCompact":[{"matcher":"","hooks":[{"type":"command","command":"other"}]}]}}`
	writeFile(t, p, orig)
	ok(t, st, nil, "install-hooks", "--settings", p, "--bin", "/x/agentmem")
	first, _ := os.ReadFile(p)
	has(t, string(first), `"model": "opus"`)
	eq(t, "PreCompact", hookCommands(t, p, "PreCompact"), []string{"|other", "|/x/agentmem hook precompact"})
	eq(t, "SessionStart", hookCommands(t, p, "SessionStart"), []string{"compact|/x/agentmem hook session-start"})
	eq(t, "UserPromptSubmit", hookCommands(t, p, "UserPromptSubmit"), []string{"|/x/agentmem hook user-prompt"})
	ok(t, st, nil, "install-hooks", "--settings", p, "--bin", "/x/agentmem")
	second, _ := os.ReadFile(p)
	eq(t, "second run changes nothing", string(second), string(first))
	bak, _ := os.ReadFile(p + ".bak")
	eq(t, "backup keeps the original", string(bak), orig)
}

func TestT55_InstallHooksCreatesFile(t *testing.T) {
	st := t.TempDir()
	p := filepath.Join(t.TempDir(), "settings.json")
	ok(t, st, nil, "install-hooks", "--settings", p, "--bin", "/x/agentmem")
	eq(t, "created", hookCommands(t, p, "SessionStart"), []string{"compact|/x/agentmem hook session-start"})
}
