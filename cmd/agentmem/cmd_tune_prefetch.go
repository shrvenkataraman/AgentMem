package main

import (
	"fmt"
	"os"
)

// cmdTunePrefetch is a stub until its ticket lands.
func cmdTunePrefetch(args []string) int {
	fmt.Fprintln(os.Stderr, "agentmem tune-prefetch: not implemented")
	return 3
}
