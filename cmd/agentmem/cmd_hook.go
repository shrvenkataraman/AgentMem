package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"agentmem/internal/facts"
	"agentmem/internal/ingest"
	"agentmem/internal/model"
	"agentmem/internal/redact"
	"agentmem/internal/store"
)

// hookPayload is the JSON Claude Code writes to a hook's stdin (only the fields used here).
type hookPayload struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
}

// cmdHook: agentmem hook precompact, reading the hook payload from stdin.
// A hook never fails its host: errors go to stderr and the exit code stays 0.
func cmdHook(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: agentmem hook precompact|session-start|user-prompt")
		return 2
	}
	var p hookPayload
	if err := json.NewDecoder(os.Stdin).Decode(&p); err != nil || p.SessionID == "" ||
		strings.ContainsAny(p.SessionID, `/\`) || p.SessionID == ".." {
		fmt.Fprintln(os.Stderr, "agentmem hook: bad payload")
		return 0
	}
	switch args[0] {
	case "precompact":
		if err := spoolDigest(p); err != nil {
			fmt.Fprintln(os.Stderr, "agentmem hook:", err)
		}
		return 0
	case "session-start":
		deliverSpool(p.SessionID, "SessionStart")
		return 0
	}
	fmt.Fprintf(os.Stderr, "agentmem hook: unknown hook %q\n", args[0])
	return 2
}

// spoolDigest writes the digest of the payload's transcript to <state>/spool/<session_id>.md.
func spoolDigest(p hookPayload) error {
	lines, _, err := ingest.ReadLines(p.TranscriptPath, 0, 1)
	if err != nil {
		return err
	}
	var msgs []model.Message
	for _, l := range lines {
		m, _ := ingest.ParseClaudeLine(p.TranscriptPath, l)
		msgs = append(msgs, m...)
	}
	msgs = ingest.DropSelfPoison(msgs, nil)
	for i := range msgs {
		msgs[i].Text = redact.Text(msgs[i].Text)
	}
	dir := filepath.Join(store.StateDir(), "spool")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, p.SessionID+".md"), []byte(facts.Text(facts.Extract(msgs))), 0o600)
}
