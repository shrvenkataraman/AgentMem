package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	if *session != "" && (strings.ContainsAny(*session, "/\\\t\n\r") || strings.Contains(*session, "..")) {
		fmt.Fprintln(os.Stderr, "invalid id")
		return 2
	}
	if *project != "" && strings.ContainsAny(*project, "\t\n\r") {
		fmt.Fprintln(os.Stderr, "invalid id")
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

	var ids []string
	if kind == "session" {
		ids = []string{val}
	} else {
		rows, err := db.Query("SELECT DISTINCT session_id FROM chunks WHERE "+col+" = ?", val)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		for rows.Next() {
			var id string
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			ids = append(ids, id)
		}
		rows.Close()
	}

	if err := store.AddForgotten(store.StateDir(), kind, val); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	res, err := db.Exec("DELETE FROM chunks WHERE "+col+" = ?", val)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	n64, _ := res.RowsAffected()
	if n64 == 0 {
		fmt.Fprintln(os.Stderr, "no chunks matched; entry recorded anyway")
	}
	for _, id := range ids {
		os.Remove(filepath.Join(store.StateDir(), "spool", id+".md")) // ignore: may not exist
	}
	if _, err := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if _, err := db.Exec("VACUUM"); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	fmt.Printf("forgot=%d\n", n64)
	return 0
}
