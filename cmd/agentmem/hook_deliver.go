package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"agentmem/internal/store"
)

const spoolTTL = 24 * time.Hour

// deliverSpool prints sid's spooled digest as hook additionalContext for event, then deletes the spool.
// A missing spool prints nothing; an expired one is deleted and prints nothing. It reads only sid's file.
func deliverSpool(sid, event string) {
	p := filepath.Join(store.StateDir(), "spool", sid+".md")
	fi, err := os.Stat(p)
	if err != nil {
		return
	}
	b, err := os.ReadFile(p)
	os.Remove(p)
	if err != nil || time.Since(fi.ModTime()) > spoolTTL {
		return
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"hookSpecificOutput": map[string]any{
		"hookEventName": event, "additionalContext": string(b)}})
}
