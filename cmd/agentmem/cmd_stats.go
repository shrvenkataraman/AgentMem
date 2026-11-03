package main

import (
	"encoding/json"
	"fmt"
	"os"

	"agentmem/internal/store"
)

// cmdStats: agentmem stats [--json] prints index stats as JSON or text.
func cmdStats(args []string) int {
	jsonOut := false
	switch len(args) {
	case 0:
	case 1:
		if args[0] != "--json" {
			fmt.Fprintln(os.Stderr, "usage: agentmem stats [--json]")
			return 2
		}
		jsonOut = true
	default:
		fmt.Fprintln(os.Stderr, "usage: agentmem stats [--json]")
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
	var sessions int
	if err := db.QueryRow("SELECT count(DISTINCT session_id) FROM chunks").Scan(&sessions); err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	byAgent := map[string]int{}
	rows, err := db.Query("SELECT agent, count(*) FROM chunks GROUP BY agent")
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	defer rows.Close()
	for rows.Next() {
		var agent string
		var n int
		if err := rows.Scan(&agent, &n); err != nil {
			fmt.Fprintln(os.Stderr, "agentmem stats:", err)
			return 1
		}
		byAgent[agent] = n
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	if !jsonOut {
		fmt.Printf("chunks=%d sessions=%d\n", chunks, sessions)
		return 0
	}
	b, err := json.Marshal(map[string]any{
		"schema_version": 1,
		"chunks":         chunks,
		"sessions":       sessions,
		"by_agent":       byAgent,
		"skipped_lines":  store.SkippedByStore(db),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	fmt.Println(string(b))
	return 0
}
