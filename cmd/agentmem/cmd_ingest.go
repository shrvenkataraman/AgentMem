package main

import (
	"flag"
	"fmt"
	"os"

	"agentmem/internal/ingest"
	"agentmem/internal/model"
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

	var msgs []model.Message
	switch *srcStore {
	case "claude", "pi":
		lines, _, err := ingest.ReadLines(*path, 0, 1)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		pp := &ingest.PiParser{Path: *path}
		for _, l := range lines {
			if *srcStore == "claude" {
				m, _ := ingest.ParseClaudeLine(*path, l)
				msgs = append(msgs, m...)
			} else {
				m, _ := pp.ParseLine(l)
				msgs = append(msgs, m...)
			}
		}
	case "hermes":
		m, _, err := ingest.ParseHermes(*path, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		msgs = m
	default:
		fmt.Fprintln(os.Stderr, "unknown --store")
		return 2
	}

	n, err := store.Upsert(db, ingest.Process(msgs, nil))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("inserted=%d\n", n)
	return 0
}
