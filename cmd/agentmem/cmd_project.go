package main

import (
	"fmt"
	"os"

	"agentmem/internal/ingest"
)

// cmdProject: agentmem project DIR prints the normalized project directory for DIR.
func cmdProject(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: agentmem project DIR")
		return 2
	}
	fmt.Println(ingest.NormalizeProject(args[0]))
	return 0
}
