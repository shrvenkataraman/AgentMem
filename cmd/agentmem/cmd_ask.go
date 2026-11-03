package main

import (
	"fmt"
	"os"
	"strings"

	"agentmem/internal/search"
)

// askOpts is the parsed `ask` command line.
type askOpts struct {
	raw   bool     // print one TSV line per FTS candidate
	words []string // query words, in the order given
}

// parseAsk parses ask args by hand: the query itself may start with `--`, so package flag is not
// usable here. Later tickets add more cases to this loop.
func parseAsk(args []string) (o askOpts) {
	for _, a := range args {
		switch a {
		case "--raw":
			o.raw = true
		default:
			o.words = append(o.words, a)
		}
	}
	return o
}

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
	hits, err := search.Candidates(db, search.Terms(o.query()), search.Filter{}, 512)
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem ask:", err)
		return 1
	}
	if !o.raw {
		return 0 // rendering lands in a later ticket
	}
	for _, h := range hits {
		fmt.Printf("%d\t%s\t0.00\n", h.ID, h.SessionID)
	}
	return 0
}
