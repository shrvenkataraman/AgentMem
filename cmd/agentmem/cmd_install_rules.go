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

const beginMarker = "<!-- agentmem:begin -->\n"
const endMarker = "<!-- agentmem:end -->\n"

// block is the instruction block install-rules writes into an agent's rules file.
const block = beginMarker + "Before re-deriving a decision, command, or error message, run `agentmem ask '<query>'`.\n" + endMarker

// cmdInstallRules: agentmem install-rules --file PATH [--remove]
// Without --remove it ingests the sources, refuses when the index is empty, then writes the block
// into FILE (replacing an existing block in place, otherwise appending). With --remove it deletes
// the block, plus one preceding newline, from an existing FILE.
func cmdInstallRules(args []string) int {
	fs := flag.NewFlagSet("install-rules", flag.ContinueOnError)
	path := fs.String("file", "", "rules file to edit")
	remove := fs.Bool("remove", false, "remove the block instead of installing it")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, "usage: agentmem install-rules --file PATH [--remove]")
		return 2
	}
	if *path == "" {
		fmt.Fprintln(os.Stderr, "install-rules: --file is required")
		return 2
	}

	if *remove {
		return removeRules(*path)
	}

	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer db.Close()

	unlock, ok := store.TryLock(store.StateDir(), store.LockWait())
	if !ok {
		fmt.Fprintln(os.Stderr, "install-rules: ingest lock busy")
		return 1
	}
	defer unlock()

	if _, err := ingest.IngestAll(db); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var n int
	if err := db.QueryRow("SELECT count(*) FROM chunks").Scan(&n); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if n == 0 {
		fmt.Fprintln(os.Stderr, "install-rules: index is empty")
		return 1
	}

	content := ""
	if b, err := os.ReadFile(*path); err == nil {
		content = string(b)
	} else if !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	var updated string
	if i := strings.Index(content, beginMarker); i >= 0 {
		if j := strings.Index(content[i:], endMarker); j >= 0 {
			end := i + j + len(endMarker)
			updated = content[:i] + block + content[end:]
		} else {
			updated = content[:i] + block
		}
	} else if content == "" {
		updated = block
	} else {
		updated = content + "\n" + block
	}

	if err := os.MkdirAll(filepath.Dir(*path), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.WriteFile(*path, []byte(updated), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// removeRules deletes the block (and one preceding newline) from FILE. A missing FILE is left
// alone; nothing is created.
func removeRules(path string) int {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	content := string(b)
	i := strings.Index(content, beginMarker)
	if i < 0 {
		return 0
	}
	j := strings.Index(content[i:], endMarker)
	if j < 0 {
		return 0
	}
	end := i + j + len(endMarker)
	if i > 0 && content[i-1] == '\n' {
		i--
	}
	if err := os.WriteFile(path, []byte(content[:i]+content[end:]), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
