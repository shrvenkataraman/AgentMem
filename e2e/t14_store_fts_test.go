package e2e

import "testing"

const insChunk = `INSERT INTO chunks(session_id,source_store,agent,project,source_path,source_ref,seq,window_idx,ts,role,text,content_hash,is_error)
VALUES(?,'claude','claude','/p','/f',?,100,0,1000,'user',?,'h',0)`

func TestT14_FTSAndTrigramTriggers(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "stats", "--json")
	db := openIdx(t, st)
	if _, err := db.Exec(insChunk, "s", "1#0", "fix ERR_FOO_BAR with --max-old-space-size=4096 now"); err != nil {
		t.Fatal(err)
	}
	f := func(tbl, q string) int {
		return count(t, db, "SELECT count(*) FROM "+tbl+" WHERE "+tbl+" MATCH ?", q)
	}
	eq(t, "fts ERR_FOO_BAR", f("chunks_fts", `"ERR_FOO_BAR"`), 1)
	eq(t, "fts flag token", f("chunks_fts", `"--max-old-space-size=4096"`), 1)
	eq(t, "fts word", f("chunks_fts", `"fix"`), 1)
	eq(t, "trigram substring", f("chunks_trgm", `"OO_BA"`), 1)
	eq(t, "trigram case-insensitive", f("chunks_trgm", `"old-space"`), 1)
	if _, err := db.Exec("UPDATE chunks SET text='now about KUBECONFIG only' WHERE session_id='s'"); err != nil {
		t.Fatal(err)
	}
	eq(t, "fts after update old", f("chunks_fts", `"ERR_FOO_BAR"`), 0)
	eq(t, "fts after update new", f("chunks_fts", `"KUBECONFIG"`), 1)
	eq(t, "trigram after update old", f("chunks_trgm", `"OO_BA"`), 0)
	eq(t, "trigram after update new", f("chunks_trgm", `"BECONF"`), 1)
	if _, err := db.Exec("DELETE FROM chunks WHERE session_id='s'"); err != nil {
		t.Fatal(err)
	}
	eq(t, "fts after delete", f("chunks_fts", `"KUBECONFIG"`), 0)
	eq(t, "trigram after delete", f("chunks_trgm", `"BECONF"`), 0)
}

func TestT14_UniqueKey(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "stats", "--json")
	db := openIdx(t, st)
	if _, err := db.Exec(insChunk, "s", "1#0", "a"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(insChunk, "s", "1#0", "b"); err == nil {
		t.Fatal("duplicate (session_id, source_ref, window_idx) must be rejected")
	}
}
