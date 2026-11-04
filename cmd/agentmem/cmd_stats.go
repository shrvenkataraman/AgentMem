package main

import (
	"database/sql"
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
	lag := computeLag(db)
	if !jsonOut {
		fmt.Printf("chunks=%d sessions=%d\n", chunks, sessions)
		return 0
	}
	b, err := json.Marshal(map[string]any{
		"schema_version": store.SchemaVersion,
		"chunks":         chunks,
		"sessions":       sessions,
		"by_agent":       byAgent,
		"skipped_lines":  store.SkippedByStore(db),
		"lag_seconds":    lag,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem stats:", err)
		return 1
	}
	fmt.Println(string(b))
	return 0
}

// computeLag returns, per agent present in chunks, how far the index is behind the sources in
// seconds: max(0, newestSourceMTime - newestIndexedTS) / 1000.
func computeLag(db *sql.DB) map[string]int64 {
	lag := map[string]int64{}
	// Collect DISTINCT source_path values per agent.
	paths := map[string][]string{}
	rows, err := db.Query("SELECT DISTINCT agent, source_path FROM chunks")
	if err != nil {
		return lag
	}
	defer rows.Close()
	for rows.Next() {
		var agent, path string
		if err := rows.Scan(&agent, &path); err != nil {
			return lag
		}
		paths[agent] = append(paths[agent], path)
	}
	if err := rows.Err(); err != nil {
		return lag
	}
	rows.Close()

	for agent, ps := range paths {
		// newest source mtime: max os.Stat(source_path).ModTime() (missing files ignored)
		var newest int64
		found := false
		for _, p := range ps {
			fi, err := os.Stat(p)
			if err != nil {
				continue
			}
			if ms := fi.ModTime().UnixMilli(); !found || ms > newest {
				newest, found = ms, true
			}
		}
		// newest indexed ts: SELECT max(ts) FROM chunks WHERE agent = ? (ms)
		var maxTS int64
		if err := db.QueryRow("SELECT COALESCE(max(ts), 0) FROM chunks WHERE agent = ?", agent).Scan(&maxTS); err != nil {
			continue
		}
		var l int64
		if found && newest-maxTS > 0 {
			l = (newest - maxTS) / 1000
		}
		lag[agent] = l
	}
	return lag
}
