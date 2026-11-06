package e2e

import (
	"os"
	"syscall"
	"testing"
)

func holdLock(t *testing.T, st string) func() {
	t.Helper()
	f, err := os.OpenFile(st+"/ingest.lock", os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	return func() { f.Close() }
}

func TestT31_AskAnswersStaleWhenLocked(t *testing.T) {
	st := corpusState(t)
	env := append(corpus(), "AGENTMEM_LOCK_WAIT_MS=100")
	release := holdLock(t, st)
	m := jsonLines(t, ok(t, st, env, "ask", "--json", "ERR_FOO_BAR"))[0]
	eq(t, "stale", m["stale"], true)
	eq(t, "answers from existing index", len(m["results"].([]any)), 6)
	release()
	m = jsonLines(t, ok(t, st, env, "ask", "--json", "ERR_FOO_BAR"))[0]
	eq(t, "fresh again", m["stale"], false)
}

func TestT31_IngestFailsWhenLocked(t *testing.T) {
	st := t.TempDir()
	ok(t, st, nil, "stats", "--json")
	release := holdLock(t, st)
	defer release()
	_, se, code := runEnv(t, st, append(corpus(), "AGENTMEM_LOCK_WAIT_MS=100"), "ingest")
	eq(t, "exit", code, 1)
	has(t, se, "ingest lock busy")
	eq(t, "nothing ingested", statsOf(t, st, nil)["chunks"], 0)
}
