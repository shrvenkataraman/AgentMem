package main

import (
	"fmt"
	"os"
)

// cmdStats is a stub until its ticket lands.
func cmdStats(args []string) int {
	fmt.Fprintln(os.Stderr, "agentmem stats: not implemented")
	return 3
}
