package main

import (
	"fmt"
	"os"
	"strings"

	"agentmem/internal/search"
)

// askOpts is the parsed `ask` command line.
type askOpts struct {
	raw    bool     // print one TSV line per FTS candidate
	limit  int      // max entries to render
	budget int      // max bytes of rendered output
	words  []string // query words, in the order given
}

// parseAsk parses ask args by hand: the query itself may start with `--`, so package flag is not
// usable here. Later tickets add more cases to this loop.
func parseAsk(args []string) (o askOpts) {
	o.limit, o.budget = defaultAskLimit, defaultAskBudget
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--raw":
			o.raw = true
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
	return o
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
	o := parseAsk(args)
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem ask:", err)
		return 1
	}
	defer db.Close()
	hits, err := search.Query(db, o.query(), search.Filter{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem ask:", err)
		return 1
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
