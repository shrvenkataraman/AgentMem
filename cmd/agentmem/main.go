package main

import (
	"fmt"
	"os"
)

// commands maps a subcommand name to its implementation. Each lives in its own cmd_<name>.go.
var commands = map[string]func(args []string) int{
	"parse":         cmdParse,
	"ingest":        cmdIngest,
	"ask":           cmdAsk,
	"show":          cmdShow,
	"stats":         cmdStats,
	"rebuild":       cmdRebuild,
	"handoff":       cmdHandoff,
	"digest":        cmdDigest,
	"prefetch":      cmdPrefetch,
	"tune-prefetch": cmdTunePrefetch,
	"audit":         cmdAudit,
	"install-rules": cmdInstallRules,
	"install-hooks": cmdInstallHooks,
	"project":       cmdProject,
	"hook":          cmdHook,
	"forget":        cmdForget,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: agentmem <command> [args]")
		os.Exit(2)
	}
	run, ok := commands[os.Args[1]]
	if !ok {
		fmt.Fprintf(os.Stderr, "agentmem: unknown command %q\n", os.Args[1])
		os.Exit(2)
	}
	os.Exit(run(os.Args[2:]))
}
