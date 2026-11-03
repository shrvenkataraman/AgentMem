package e2e

import "testing"

func TestT15_IngestFileClaude(t *testing.T) {
	st := t.TempDir()
	eq(t, "first", ok(t, st, nil, "ingest", "--file", fixture("claude/s1.jsonl"), "--store", "claude"), "inserted=7\n")
	eq(t, "second", ok(t, st, nil, "ingest", "--file", fixture("claude/s1.jsonl"), "--store", "claude"), "inserted=0\n")
	db := openIdx(t, st)
	eq(t, "rows", count(t, db, "SELECT count(*) FROM chunks"), 7)
	eq(t, "both yes rows kept", count(t, db, "SELECT count(*) FROM chunks WHERE text='yes'"), 2)
	eq(t, "poison absent", count(t, db, "SELECT count(*) FROM chunks WHERE text LIKE '%poison%' OR text LIKE '%agentmem ask%'"), 0)
	eq(t, "fts in sync", count(t, db, `SELECT count(*) FROM chunks_fts WHERE chunks_fts MATCH '"Running"'`), 1)
}

func TestT15_IngestFilePiAndHermes(t *testing.T) {
	st := t.TempDir()
	eq(t, "pi", ok(t, st, nil, "ingest", "--file", fixture("pi/s1.jsonl"), "--store", "pi"), "inserted=6\n")
	eq(t, "hermes", ok(t, st, nil, "ingest", "--file", fixture("hermes/state.db"), "--store", "hermes"), "inserted=6\n")
	db := openIdx(t, st)
	eq(t, "agents", count(t, db, "SELECT count(DISTINCT agent) FROM chunks"), 2)
	eq(t, "hermes stores", count(t, db, "SELECT count(DISTINCT source_store) FROM chunks WHERE agent='hermes'"), 2)
}

func TestT15_IngestRedactsAndKeepsUpdates(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "ingest", "--file", fixture("redact/secrets.jsonl"), "--store", "claude")
	db := openIdx(t, st)
	eq(t, "aws gone", count(t, db, "SELECT count(*) FROM chunks WHERE text LIKE '%AKIAIOSFODNN7EXAMPLE%'"), 0)
	eq(t, "marker present", count(t, db, "SELECT count(*) FROM chunks WHERE text LIKE '%[REDACTED:aws]%'"), 1)
}
