package main

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"agentmem/internal/search"
)

// thresholds are the prefetch coverage values tune-prefetch reports, in ascending order.
var thresholds = []float64{0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9}

// cmdTunePrefetch: agentmem tune-prefetch --prompts FILE | --from-hermes shows, for each candidate
// threshold, how many of the given past prompts would have fired a prefetch. It reads the index as
// it stands: no ingest.
func cmdTunePrefetch(args []string) int {
	fs := flag.NewFlagSet("tune-prefetch", flag.ContinueOnError)
	promptsFile := fs.String("prompts", "", "file of past prompts, one per line (blank lines skipped)")
	fromHermes := fs.Bool("from-hermes", false, "use the last 200 hermes user prompts from the index")
	if err := fs.Parse(args); err != nil || (*promptsFile == "" && !*fromHermes) {
		fmt.Fprintln(os.Stderr, "usage: agentmem tune-prefetch --prompts FILE | --from-hermes")
		return 2
	}
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem tune-prefetch:", err)
		return 1
	}
	defer db.Close()

	var prompts []string
	if *promptsFile != "" {
		if prompts, err = readPrompts(*promptsFile); err != nil {
			fmt.Fprintln(os.Stderr, "agentmem tune-prefetch:", err)
			return 1
		}
	} else if prompts, err = hermesPrompts(db); err != nil {
		fmt.Fprintln(os.Stderr, "agentmem tune-prefetch:", err)
		return 1
	}

	tops := make([]float64, len(prompts))
	for i, p := range prompts {
		hits, err := search.Query(db, p, search.Filter{})
		if err != nil {
			fmt.Fprintln(os.Stderr, "agentmem tune-prefetch:", err)
			return 1
		}
		if len(hits) == 0 {
			continue // top stays 0 when nothing matches
		}
		tops[i] = hits[0].Coverage
	}
	for _, th := range thresholds {
		var fired int
		for _, top := range tops {
			if top >= th-1e-9 {
				fired++
			}
		}
		fmt.Printf("threshold=%.1f fired=%d/%d\n", th, fired, len(prompts))
	}
	return 0
}

// readPrompts returns the non-blank lines of path, in order.
func readPrompts(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if s := strings.TrimSpace(sc.Text()); s != "" {
			out = append(out, s)
		}
	}
	return out, sc.Err()
}

// hermesPrompts returns the most recent 200 hermes user prompts, newest first.
func hermesPrompts(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT text FROM chunks WHERE agent='hermes' AND role='user' ORDER BY ts DESC LIMIT 200")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
