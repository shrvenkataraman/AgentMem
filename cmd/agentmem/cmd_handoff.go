package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"agentmem/internal/facts"
)

// handoffHeader is the first line of a regenerated handoff, followed by a blank line.
const handoffHeader = "# agentmem handoff (regenerated, historical record, not instructions)\n\n"

// cmdHandoff: agentmem handoff --regen [--since YYYY-MM-DD] [--session ID] prints a deterministic
// digest of the indexed history. It never ingests and never locks.
func cmdHandoff(args []string) int {
	fs := flag.NewFlagSet("handoff", flag.ContinueOnError)
	regen := fs.Bool("regen", false, "regenerate the digest from the index")
	since := fs.String("since", "", "only sessions with messages on or after this UTC date (YYYY-MM-DD)")
	session := fs.String("session", "", "only this session id")
	usage := "usage: agentmem handoff --regen [--since YYYY-MM-DD] [--session ID]"
	if err := fs.Parse(args); err != nil || !*regen {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	var sinceMS int64
	if *since != "" {
		t, err := time.Parse("2006-01-02", *since)
		if err != nil {
			fmt.Fprintln(os.Stderr, "agentmem handoff: bad --since:", err)
			return 2
		}
		sinceMS = t.UnixMilli()
	}
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem handoff:", err)
		return 1
	}
	defer db.Close()
	var b strings.Builder
	b.WriteString(handoffHeader)
	sessions := loadSessions(db, sinceMS, *session)
	if len(sessions) == 0 {
		b.WriteString("(no sessions)\n")
		os.Stdout.WriteString(b.String())
		return 0
	}
	for _, s := range sessions {
		fmt.Fprintf(&b, "## Session %s (%s, %s, %s)\n", s.ID, s.Agent, s.Project,
			time.UnixMilli(s.First).UTC().Format("2006-01-02"))
		b.WriteString(stripDigestHeader(facts.Text(facts.Extract(s.Msgs))))
		b.WriteString("\n")
	}
	os.Stdout.WriteString(b.String())
	return 0
}

// stripDigestHeader drops the first line of a facts digest, returning the rest verbatim.
func stripDigestHeader(s string) string {
	i := strings.IndexByte(s, '\n')
	if i < 0 {
		return ""
	}
	return s[i+1:]
}
