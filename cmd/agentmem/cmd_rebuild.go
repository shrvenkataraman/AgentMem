package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"agentmem/internal/ingest"
	"agentmem/internal/store"
)

// cmdRebuild: agentmem rebuild [--force]
// Builds a fresh index beside the old one and swaps it in only when it is not smaller (--force
// swaps anyway). Prints "rebuilt old=N new=N swapped=B" and exits 1 when it did not swap.
func cmdRebuild(args []string) int {
	fs := flag.NewFlagSet("rebuild", flag.ContinueOnError)
	force := fs.Bool("force", false, "swap even if the new index is smaller")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "usage: agentmem rebuild [--force]")
		return 2
	}

	release, err := store.TryIngestLock(store.StateDir())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer release()

	oldN, newN, swapped, err := store.Rebuild(store.StateDir(), *force, func(db *sql.DB) error {
		_, err := ingest.IngestAll(db)
		return err
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("rebuilt old=%d new=%d swapped=%t\n", oldN, newN, swapped)
	if !swapped {
		return 1
	}
	return 0
}
