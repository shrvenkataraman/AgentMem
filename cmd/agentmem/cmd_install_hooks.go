package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

// hookSpecs lists the hooks installed: settings event, matcher, and the `agentmem hook` subcommand.
var hookSpecs = []struct{ event, matcher, sub string }{
	{"PreCompact", "", "precompact"},
	{"SessionStart", "compact", "session-start"},
	{"UserPromptSubmit", "", "user-prompt"},
}

// cmdInstallHooks: agentmem install-hooks --settings PATH [--bin PATH]
// Our entry for each event is removed, then appended last, so reruns change nothing.
// Invalid JSON in an existing file is never rewritten.
func cmdInstallHooks(args []string) int {
	fs := flag.NewFlagSet("install-hooks", flag.ContinueOnError)
	settings := fs.String("settings", "", "Claude settings.json to edit")
	binPath := fs.String("bin", "", "agentmem binary path (default: this executable)")
	remove := fs.Bool("remove", false, "remove the hooks instead of installing them")
	if err := fs.Parse(args); err != nil || *settings == "" {
		fmt.Fprintln(os.Stderr, "usage: agentmem install-hooks --settings PATH [--bin PATH] [--remove]")
		return 2
	}
	if *binPath == "" {
		exe, err := os.Executable()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		*binPath = exe
	}
	root := map[string]any{}
	if b, err := os.ReadFile(*settings); err == nil {
		if err := json.Unmarshal(b, &root); err != nil {
			fmt.Fprintln(os.Stderr, "install-hooks: settings is not valid JSON:", err)
			return 1
		}
		if _, err := os.Stat(*settings + ".bak"); err != nil {
			os.WriteFile(*settings+".bak", b, 0o600)
		}
	}
	hooks, _ := root["hooks"].(map[string]any)
	if hooks == nil {
		hooks = map[string]any{}
	}
	for _, s := range hookSpecs {
		groups, _ := hooks[s.event].([]any)
		var kept []any
		for _, g := range groups {
			if !isOurs(g, " hook "+s.sub) {
				kept = append(kept, g)
			}
		}
		if !*remove {
			kept = append(kept, map[string]any{"matcher": s.matcher, "hooks": []any{
				map[string]any{"type": "command", "command": *binPath + " hook " + s.sub}}})
		}
		if len(kept) == 0 {
			delete(hooks, s.event)
		} else {
			hooks[s.event] = kept
		}
	}
	if len(hooks) == 0 {
		delete(root, "hooks")
	} else {
		root["hooks"] = hooks
	}
	out, err := json.MarshalIndent(root, "", "  ")
	if err == nil {
		err = os.WriteFile(*settings, append(out, '\n'), 0o600)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// isOurs reports whether hook group g has a command ending with suffix.
func isOurs(g any, suffix string) bool {
	m, _ := g.(map[string]any)
	hs, _ := m["hooks"].([]any)
	for _, h := range hs {
		hm, _ := h.(map[string]any)
		if c, _ := hm["command"].(string); strings.HasSuffix(c, suffix) {
			return true
		}
	}
	return false
}
