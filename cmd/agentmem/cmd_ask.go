package main

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"agentmem/internal/ingest"
	"agentmem/internal/querylog"
	"agentmem/internal/search"
	"agentmem/internal/store"
)

// askOpts is the parsed `ask` command line.
type askOpts struct {
	raw    bool          // print one TSV line per FTS candidate
	limit  int           // max entries to render
	budget int           // max bytes of rendered output
	words  []string      // query words, in the order given
	filter search.Filter // narrowing options (--project, --agent, --since)

	session        string // --session S: the session to treat as the current one
	includeCurrent bool   // --include-current: do not exclude the current session
	json           bool   // --json: print one machine-readable object instead of the rendering

	noDedup bool // --no-dedup: do not collapse identical chunks from different sessions

	noRefresh bool // --no-refresh: do not ingest new transcript lines before answering
}

// parseAsk parses ask args by hand: the query itself may start with `--`, so package flag is not
// usable here. Later tickets add more cases to this loop. A malformed option returns an error.
func parseAsk(args []string) (o askOpts, err error) {
	o.limit, o.budget = defaultAskLimit, defaultAskBudget
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			o.raw = true
		case "--project", "--agent":
			name := args[i]
			if i+1 < len(args) {
				if name == "--project" {
					o.filter.Project = args[i+1]
				} else {
					o.filter.Agent = args[i+1]
				}
				i++
			}
		case "--since":
			if i+1 < len(args) {
				t, perr := time.Parse("2006-01-02", args[i+1])
				if perr != nil {
					return o, errors.New("bad --since")
				}
				o.filter.SinceMS = t.UTC().UnixMilli()
				i++
			}
		case "--session":
			if i+1 < len(args) {
				o.session = args[i+1]
				i++
			}
		case "--include-current":
			o.includeCurrent = true
		case "--json":
			o.json = true
		case "--no-dedup":
			o.noDedup = true
		case "--no-refresh":
			o.noRefresh = true
		case "--limit", "--budget":
			name := args[i]
			if i+1 < len(args) {
				var n int
				if _, err := fmt.Sscanf(args[i+1], "%d", &n); err == nil {
					if name == "--limit" {
						o.limit = n
					} else {
						o.budget = n
					}
				}
				i++
			}
		default:
			o.words = append(o.words, args[i])
		}
	}
	return o, nil
}

// logQuery records this answered question in the private query log. The agent is AGENTMEM_AGENT
// (or "unknown"), the project is the normalized cwd, sources are the distinct agents of the
// returned hits (never nil, so an empty result set logs []).
func logQuery(o askOpts, hits []search.Hit) {
	agent := os.Getenv("AGENTMEM_AGENT")
	if agent == "" {
		agent = "unknown"
	}
	cwd := os.Getenv("AGENTMEM_CWD")
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	seen := map[string]bool{}
	sources := make([]string, 0, len(hits))
	for _, h := range hits {
		if !seen[h.Agent] {
			seen[h.Agent] = true
			sources = append(sources, h.Agent)
		}
	}
	sort.Strings(sources)
	querylog.Append(store.StateDir(), querylog.Entry{
		TS:              time.Now().UTC().Format(time.RFC3339),
		Agent:           agent,
		Project:         ingest.NormalizeProject(cwd),
		Query:           o.query(),
		Hits:            len(hits),
		ZeroResult:      len(hits) == 0,
		SourcesReturned: sources,
	})
}

// Defaults for the rendered output size.
const (
	defaultAskLimit  = 8
	defaultAskBudget = 6144
)

// query joins the query words back into one string.
func (o askOpts) query() string { return strings.Join(o.words, " ") }

// cmdAsk: agentmem ask [--raw] WORDS prints the chunks matching WORDS.
func cmdAsk(args []string) int {
	o, err := parseAsk(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ask:", err)
		return 2
	}
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem ask:", err)
		return 1
	}
	defer db.Close()
	stale := false
	if !o.noRefresh {
		if unlock, ok := store.TryLock(store.StateDir(), store.LockWait()); ok {
			defer unlock()
			if _, err := ingest.IngestAll(db); err != nil {
				fmt.Fprintln(os.Stderr, "ask: refresh:", err)
			}
		} else {
			// Another ingest holds the lock: answer from the index as it stands.
			stale = true
		}
	}
	if !o.includeCurrent {
		o.filter.ExcludeSession = search.CurrentSession(db, o.session)
	}
	hits, err := search.Query(db, o.query(), o.filter)
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem ask:", err)
		return 1
	}
	if !o.noDedup {
		hits = search.Collapse(hits)
	}
	logQuery(o, hits)
	if o.json {
		printJSON(o.query(), hits, o.limit, o.filter.ExcludeSession, stale)
		return 0
	}
	if !o.raw {
		fmt.Print(search.Render(hits, o.limit, o.budget))
		return 0
	}
	for _, h := range hits {
		fmt.Printf("%d\t%s\t%.2f\n", h.ID, h.SessionID, h.Coverage)
	}
	return 0
}
