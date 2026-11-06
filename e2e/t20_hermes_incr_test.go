package e2e

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestT20_FullCorpus(t *testing.T) {
	st := t.TempDir()
	eq(t, "first", ok(t, st, corpus(), "ingest"), "files=6 inserted=21 skipped=0\n")
	eq(t, "second", ok(t, st, corpus(), "ingest"), "files=6 inserted=0 skipped=0\n")
	by := statsOf(t, st, nil)["by_agent"].(map[string]any)
	eq(t, "hermes", by["hermes"], 6)
	eq(t, "total", statsOf(t, st, nil)["chunks"], 21)
}

func TestT20_HermesIncrementalAndPoisonCarry(t *testing.T) {
	src, _ := os.ReadFile(fixture("hermes/state.db"))
	dbf := filepath.Join(t.TempDir(), "state.db")
	os.WriteFile(dbf, src, 0o644)
	env := []string{"AGENTMEM_HERMES_DB=" + dbf}
	st := t.TempDir()
	eq(t, "base", ok(t, st, env, "ingest"), "files=1 inserted=6 skipped=0\n")
	idx := openIdx(t, st)
	eq(t, "cursor", cursorOf(t, idx, "hermes", dbf), "9")
	h, err := sql.Open("sqlite", "file:"+dbf)
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()
	ins := func(q string, a ...any) {
		if _, err := h.Exec(q, a...); err != nil {
			t.Fatal(err)
		}
	}
	ins(`INSERT INTO messages(session_id,role,content,timestamp) VALUES('h1','user','new msg',1790000200.0)`)
	ins(`INSERT INTO messages(session_id,role,content,tool_calls,timestamp) VALUES('h1','assistant','',?,1790000201.0)`,
		`[{"id":"c9","type":"function","function":{"name":"terminal","arguments":"{\"command\": \"agentmem ask z\"}"}}]`)
	eq(t, "new rows", ok(t, st, env, "ingest"), "files=1 inserted=1 skipped=0\n")
	eq(t, "cursor after", cursorOf(t, idx, "hermes", dbf), "11")
	ins(`INSERT INTO messages(session_id,role,content,tool_call_id,tool_name,timestamp) VALUES('h1','tool','poison-late','c9','terminal',1790000202.0)`)
	eq(t, "late result", ok(t, st, env, "ingest"), "files=1 inserted=0 skipped=0\n")
	eq(t, "poison absent", count(t, idx, "SELECT count(*) FROM chunks WHERE text='poison-late'"), 0)
	eq(t, "rows", count(t, idx, "SELECT count(*) FROM chunks"), 7)
}
