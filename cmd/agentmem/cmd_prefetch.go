package main

import (
	"flag"
	"fmt"
	"os"

	"agentmem/internal/search"
)

// cmdPrefetch: agentmem prefetch --query Q [--threshold T] prints the top 3 hits only when the
// best one covers enough of the query. It reads the index as it stands: no ingest.
func cmdPrefetch(args []string) int {
	fs := flag.NewFlagSet("prefetch", flag.ContinueOnError)
	query := fs.String("query", "", "query text (required)")
	threshold := fs.Float64("threshold", 0.6, "minimum coverage of the best hit to print")
	if err := fs.Parse(args); err != nil || *query == "" {
		fmt.Fprintln(os.Stderr, "usage: agentmem prefetch --query Q [--threshold T]")
		return 2
	}
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem prefetch:", err)
		return 1
	}
	defer db.Close()
	hits, err := search.Query(db, *query, search.Filter{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem prefetch:", err)
		return 1
	}
	hits = search.Collapse(hits)
	if len(hits) == 0 || hits[0].Coverage < *threshold {
		return 0
	}
	fmt.Print(search.Render(hits, 3, 6144))
	return 0
}
