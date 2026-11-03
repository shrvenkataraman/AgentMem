package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"agentmem/internal/facts"
	"agentmem/internal/model"
)

// cmdDigest: agentmem digest --stdin [--format json]
// Reads JSON-lines model.Message from stdin and prints the extracted facts.
func cmdDigest(args []string) int {
	fs := flag.NewFlagSet("digest", flag.ContinueOnError)
	stdin := fs.Bool("stdin", false, "read JSON-lines model.Message from stdin")
	format := fs.String("format", "text", "text|json")
	if err := fs.Parse(args); err != nil || !*stdin {
		fmt.Fprintln(os.Stderr, "usage: agentmem digest --stdin [--format json]")
		return 2
	}
	msgs, err := readMessages(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	f := facts.Extract(msgs)
	if *format != "json" {
		fmt.Fprintln(os.Stderr, "agentmem digest: --format text is not implemented")
		return 3
	}
	b, err := json.Marshal(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	os.Stdout.Write(append(b, '\n'))
	return 0
}

// readMessages decodes newline-delimited model.Message objects.
func readMessages(r io.Reader) ([]model.Message, error) {
	var msgs []model.Message
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		ln := sc.Bytes()
		if len(ln) == 0 {
			continue
		}
		var m model.Message
		if err := json.Unmarshal(ln, &m); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	return msgs, sc.Err()
}
