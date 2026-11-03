package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"agentmem/internal/ingest"
	"agentmem/internal/model"
	"agentmem/internal/redact"
)

// cmdParse: agentmem parse --store claude|pi|hermes [--stage raw|filtered|redacted|chunks] FILE
// Debug command: prints one JSON object per line (Message, or Chunk for stage chunks).
func cmdParse(args []string) int {
	fs := flag.NewFlagSet("parse", flag.ContinueOnError)
	store := fs.String("store", "", "claude|pi|hermes")
	stage := fs.String("stage", "raw", "raw|filtered|redacted|chunks")
	if err := fs.Parse(args); err != nil || fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: agentmem parse --store S [--stage ST] FILE")
		return 2
	}
	path := fs.Arg(0)
	var msgs []model.Message
	switch *store {
	case "claude", "pi":
		lines, _, err := ingest.ReadLines(path, 0, 1)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		pp := &ingest.PiParser{Path: path}
		for _, l := range lines {
			if *store == "claude" {
				m, _ := ingest.ParseClaudeLine(path, l)
				msgs = append(msgs, m...)
			} else {
				m, _ := pp.ParseLine(l)
				msgs = append(msgs, m...)
			}
		}
	case "hermes":
		m, maxID, err := ingest.ParseHermes(path, 0)
		fmt.Fprintf(os.Stderr, "max_id=%d\n", maxID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		msgs = m
	default:
		fmt.Fprintln(os.Stderr, "unknown --store")
		return 2
	}
	enc := json.NewEncoder(os.Stdout)
	if *stage != "raw" {
		msgs = ingest.DropSelfPoison(msgs, nil)
	}
	if *stage == "redacted" || *stage == "chunks" {
		for i := range msgs {
			msgs[i].Text = redact.Text(msgs[i].Text)
		}
	}
	if *stage == "chunks" {
		for _, c := range ingest.Window(msgs) {
			enc.Encode(c)
		}
		return 0
	}
	for _, m := range msgs {
		enc.Encode(m)
	}
	return 0
}
