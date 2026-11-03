package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// cmdStats: agentmem stats --json prints index stats as JSON.
func cmdStats(args []string) int {
	if len(args) != 1 || args[0] != "--json" {
		fmt.Fprintln(os.Stderr, "usage: agentmem stats --json")
		return 2
	}
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	defer db.Close()
	var chunks int
	if err := db.QueryRow("SELECT count(*) FROM chunks").Scan(&chunks); err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	b, err := json.Marshal(map[string]any{"schema_version": 1, "chunks": chunks})
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	fmt.Println(string(b))
	return 0
}
