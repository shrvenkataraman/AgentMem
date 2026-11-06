package e2e

import (
	"path/filepath"
	"testing"
)

func TestT43_ReplayPromptsFile(t *testing.T) {
	st := corpusState(t)
	f := filepath.Join(t.TempDir(), "prompts.txt")
	writeFile(t, f, "ERR_FOO_BAR build\nERR_FOO_BAR zzzunknown qqq\n\nnothing matches qqqqzzzz\n")
	want := "threshold=0.3 fired=2/3\nthreshold=0.4 fired=1/3\nthreshold=0.5 fired=1/3\nthreshold=0.6 fired=1/3\n" +
		"threshold=0.7 fired=1/3\nthreshold=0.8 fired=1/3\nthreshold=0.9 fired=1/3\n"
	eq(t, "table", ok(t, st, nil, "tune-prefetch", "--prompts", f), want)
}

func TestT43_FromHermes(t *testing.T) {
	st := corpusState(t)
	out := ok(t, st, corpus(), "tune-prefetch", "--from-hermes")
	has(t, out, "threshold=0.3 fired=2/2\n")
	has(t, out, "threshold=0.9 fired=2/2\n")
}

func TestT43_NeedsASource(t *testing.T) {
	_, _, code := runEnv(t, t.TempDir(), nil, "tune-prefetch")
	eq(t, "exit", code, 2)
}
