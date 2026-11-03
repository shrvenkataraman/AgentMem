package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"

	"agentmem/internal/search"
)

// showOpts is the parsed `show` command line.
type showOpts struct {
	id      int64 // the chunk to print
	context int   // --context N: also print N neighbours on each side
}

// parseShow parses `show [--context N] ID` by hand: anything that is not the --context option is
// taken as the chunk id, and exactly one id is required.
func parseShow(args []string) (o showOpts, err error) {
	var ids []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--context":
			if i+1 >= len(args) {
				return o, errors.New("bad --context")
			}
			n, cerr := strconv.Atoi(args[i+1])
			if cerr != nil || n < 0 {
				return o, errors.New("bad --context")
			}
			o.context = n
			i++
		default:
			ids = append(ids, args[i])
		}
	}
	if len(ids) != 1 {
		return o, errors.New("usage: agentmem show [--context N] ID")
	}
	id, cerr := strconv.ParseInt(ids[0], 10, 64)
	if cerr != nil {
		return o, errors.New("bad id " + ids[0])
	}
	o.id = id
	return o, nil
}

// showChunk prints one chunk: its header, the source line and the whole text.
func showChunk(h search.Hit) {
	fmt.Print(search.EntryHeader(h) + "\n" + "source: " + h.SourcePath + ":" + h.SourceRef +
		"\n" + h.Text + "\n\n")
}

// cmdShow: agentmem show [--context N] ID prints the full chunk ID and, with --context, its
// nearest N neighbours in the same session.
func cmdShow(args []string) int {
	o, err := parseShow(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "show:", err)
		return 2
	}
	db, err := openDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem show:", err)
		return 1
	}
	defer db.Close()

	h, err := search.Get(db, o.id)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Fprintln(os.Stderr, "show: no such chunk")
		return 1
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem show:", err)
		return 1
	}
	if o.context <= 0 {
		showChunk(h)
		return 0
	}
	list, err := search.Neighbors(db, h, o.context)
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem show:", err)
		return 1
	}
	for _, n := range list {
		if n.ID == h.ID {
			fmt.Print("▶ ")
		}
		showChunk(n)
	}
	return 0
}
