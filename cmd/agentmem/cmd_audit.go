package main

import (
	"fmt"
	"os"
)

// cmdAudit is a stub until its ticket lands.
func cmdAudit(args []string) int {
	fmt.Fprintln(os.Stderr, "agentmem audit: not implemented")
	return 3
}
