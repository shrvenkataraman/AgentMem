package main

import (
	"flag"
	"fmt"
	"os"

	"agentmem/internal/ingest"
	"agentmem/internal/store"
)

// cmdIngest: agentmem ingest --file F --store claude|pi|hermes
// Reads one harness transcript file, runs the ingest pipeline and upserts the chunks.
// Prints "inserted=N".
func cmdIngest(args []string) int {
	fs := flag.NewFlagSet("ingest", flag.ContinueOnError)
	path := fs.String("file", "", "transcript file")
	srcStore := fs.String("store", "", "claude|pi|hermes")
	if err := fs.Parse(args); err != nil || *path == "" || *srcStore == "" {
		fmt.Fprintln(os.Stderr, "usage: agentmem ingest --file F --store claude|pi|hermes")
		return 2
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer db.Close()

	switch *srcStore {
	case "claude", "pi":
		n, _, err := ingest.IngestJSONL(db, *srcStore, *path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("inserted=%d\n", n)
	case "hermes":
		m, _, err := ingest.ParseHermes(*path, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		n, err := store.Upsert(db, ingest.Process(m, nil))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		fmt.Printf("inserted=%d\n", n)
	default:
		fmt.Fprintln(os.Stderr, "unknown --store")
		return 2
	}
	return 0
}
