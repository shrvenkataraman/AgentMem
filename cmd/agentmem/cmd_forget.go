package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"agentmem/internal/ingest"
	"agentmem/internal/store"
)

// cmdForget: agentmem forget --session ID | --project PATH [--dry-run]
// Removes that memory from the index and the spool and records it on the forgotten skip-list so
// future ingests and rebuilds keep it out. Prints "would-forget=N" with --dry-run, "forgot=N" otherwise.
func cmdForget(args []string) int {
	fs := flag.NewFlagSet("forget", flag.ContinueOnError)
	session := fs.String("session", "", "session id to forget")
	project := fs.String("project", "", "project path to forget")
	dryRun := fs.Bool("dry-run", false, "print the chunk count and change nothing")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "usage: agentmem forget --session ID | --project PATH [--dry-run]")
		return 2
	}
	if (*session == "") == (*project == "") {
		fmt.Fprintln(os.Stderr, "usage: agentmem forget --session ID | --project PATH [--dry-run]")
		return 2
	}

	kind, col, val := "session", "session_id", *session
	if *project != "" {
		kind, col, val = "project", "project", ingest.NormalizeProject(*project)
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer db.Close()

	var n int
	if err := db.QueryRow("SELECT count(*) FROM chunks WHERE "+col+" = ?", val).Scan(&n); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *dryRun {
		fmt.Printf("would-forget=%d\n", n)
		return 0
	}

	unlock, ok := store.TryLock(store.StateDir(), store.LockWait())
	if !ok {
		fmt.Fprintln(os.Stderr, "ingest lock busy")
		return 1
	}
	defer unlock()

	if err := store.AddForgotten(store.StateDir(), kind, val); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if _, err := db.Exec("DELETE FROM chunks WHERE "+col+" = ?", val); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if kind == "session" {
		os.Remove(filepath.Join(store.StateDir(), "spool", val+".md"))
	}
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if _, err := db.Exec("VACUUM"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("forgot=%d\n", n)
	return 0
}
