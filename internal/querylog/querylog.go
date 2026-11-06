// Package querylog records every `ask` as one JSON line in a private, size-rotated log.
package querylog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
)

// Entry is one recorded question.
type Entry struct {
	TS              string   `json:"ts"`
	Agent           string   `json:"agent"`
	SessionID       string   `json:"session_id"`
	Project         string   `json:"project"`
	Query           string   `json:"query"`
	Hits            int      `json:"hits"`
	ZeroResult      bool     `json:"zero_result"`
	SourcesReturned []string `json:"sources_returned"`
}

// defaultMax is the log size in bytes that triggers a rotation when AGENTMEM_QUERYLOG_MAX is unset.
const defaultMax = 5242880

// name is the log file's name inside the state dir; rotation moves it to name + ".1".
const name = "query-log.jsonl"

// maxBytes is AGENTMEM_QUERYLOG_MAX, else defaultMax. Unparseable values fall back to defaultMax.
func maxBytes() int64 {
	if v := os.Getenv("AGENTMEM_QUERYLOG_MAX"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultMax
}

// Append writes e as one JSON line to <dir>/query-log.jsonl (0600, appended). When the file has
// already grown to AGENTMEM_QUERYLOG_MAX bytes it is renamed to query-log.jsonl.1 first. All
// errors are ignored: a query log must never break or slow down an answer.
// AGENTMEM_QUERYLOG=off disables it.
func Append(dir string, e Entry) {
	if os.Getenv("AGENTMEM_QUERYLOG") == "off" {
		return // eval runs must not count as real asks
	}
	path := filepath.Join(dir, name)
	if fi, err := os.Stat(path); err == nil && fi.Size() >= maxBytes() {
		os.Rename(path, path+".1")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	b, err := json.Marshal(e)
	if err != nil {
		return
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		return
	}
	os.Chmod(path, 0o600)
}
