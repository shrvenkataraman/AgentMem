package e2e

import "testing"

func TestT30_AskRefreshesFirst(t *testing.T) {
	st := t.TempDir()
	r := asks(t, st, corpus(), "ERR_FOO_BAR")
	eq(t, "results without prior ingest", len(r), 6)
	eq(t, "chunks", statsOf(t, st, nil)["chunks"], 21)
}

func TestT30_NoRefreshFlag(t *testing.T) {
	st := t.TempDir()
	eq(t, "no-refresh on empty index", len(asks(t, st, corpus(), "--no-refresh", "ERR_FOO_BAR")), 0)
	eq(t, "still empty", statsOf(t, st, nil)["chunks"], 0)
}
