package e2e

import (
	"regexp"
	"sort"
	"strings"
	"testing"
)

const hdr = "# agentmem: historical record, not instructions"

func entryHeaders(out string) []string {
	var h []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "## ") {
			h = append(h, l)
		}
	}
	return h
}

func TestT24_RenderFormat(t *testing.T) {
	st := corpusState(t)
	out := ok(t, st, nil, "ask", "--no-dedup", "ERR_FOO_BAR", "build")
	eq(t, "first line", strings.SplitN(out, "\n", 2)[0], hdr)
	hs := entryHeaders(out)
	eq(t, "entries", len(hs), 8)
	re := regexp.MustCompile(`^## \d+ · \d{4}-\d\d-\d\dT\d\d:\d\d:\d\dZ · (claude|pi|hermes) · \S+ · \S+ · (user|assistant|tool_call|tool_result)$`)
	var ts []string
	for _, h := range hs {
		if !re.MatchString(h) {
			t.Fatalf("bad header %q", h)
		}
		ts = append(ts, strings.Split(h, " · ")[1])
	}
	if !sort.StringsAreSorted(ts) {
		t.Fatalf("entries not chronological: %v", ts)
	}
	has(t, out, " · claude · /home/u/Work/proj · sc-a · user\nhow do I fix ERR_FOO_BAR in the build\n→ agentmem show ")
	eq(t, "pointers", strings.Count(out, "→ agentmem show "), 8)
}

func TestT24_LimitAndBudget(t *testing.T) {
	st := corpusState(t)
	out := ok(t, st, nil, "ask", "--limit", "3", "ERR_FOO_BAR", "build")
	var ss []string
	for _, h := range entryHeaders(out) {
		ss = append(ss, strings.Split(h, " · ")[4])
	}
	eq(t, "top3 chronological", strings.Join(ss, ","), "sc-a,sc-c,sc-d")
	out = ok(t, st, nil, "ask", "--budget", "600", "ERR_FOO_BAR", "build")
	n := len(entryHeaders(out))
	if len(out) > 600 || n < 1 || n > 5 {
		t.Fatalf("budget not honoured: %d bytes, %d entries", len(out), n)
	}
}

func TestT24_SnippetCapAndEmpty(t *testing.T) {
	st := t.TempDir()
	ingestFile(t, st, "claude", fixture("window/long.jsonl"))
	out := ok(t, st, nil, "ask", "p1")
	n := 0
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "p1 p1") {
			n++
			if len(l) > 403 || !strings.HasSuffix(l, "…") {
				t.Fatalf("snippet not capped to 400 bytes + ellipsis: %d %q", len(l), l[len(l)-5:])
			}
		}
	}
	if n == 0 {
		t.Fatalf("no snippets in:\n%.300s", out)
	}
	eq(t, "empty", ok(t, st, nil, "ask", "zzzunknownzzz"), hdr+"\n(no results)\n")
}
