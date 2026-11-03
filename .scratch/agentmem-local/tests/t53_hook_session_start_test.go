package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// spoolFile writes a spool digest for sid under state and returns its path.
func spoolFile(t *testing.T, state, sid, body string) string {
	t.Helper()
	p := filepath.Join(state, "spool", sid+".md")
	writeFile(t, p, body)
	return p
}

// injected decodes a hook's stdout and returns its event name and additionalContext.
func injected(t *testing.T, so string) (string, string) {
	t.Helper()
	var out struct {
		H struct {
			Event string `json:"hookEventName"`
			Ctx   string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal([]byte(so), &out); err != nil {
		t.Fatalf("stdout is not hook JSON: %q (%v)", so, err)
	}
	return out.H.Event, out.H.Ctx
}

func TestT53_SessionStartDeliversOnceThenDeletes(t *testing.T) {
	st := t.TempDir()
	p := spoolFile(t, st, "sess-1", "DIGEST BODY")
	so, _, code := runStdin(t, st, hookPayload("sess-1", "", "SessionStart"), "hook", "session-start")
	eq(t, "exit", code, 0)
	ev, ctx := injected(t, so)
	eq(t, "event", ev, "SessionStart")
	has(t, ctx, "DIGEST BODY")
	_, err := os.Stat(p)
	eq(t, "spool deleted", os.IsNotExist(err), true)
	so, _, _ = runStdin(t, st, hookPayload("sess-1", "", "SessionStart"), "hook", "session-start")
	eq(t, "second call is silent", so, "")
}

func TestT53_SessionStartNeverCrossesSessions(t *testing.T) {
	st := t.TempDir()
	p := spoolFile(t, st, "sess-1", "DIGEST BODY")
	so, _, code := runStdin(t, st, hookPayload("sess-2", "", "SessionStart"), "hook", "session-start")
	eq(t, "exit", code, 0)
	eq(t, "silent", so, "")
	if _, err := os.Stat(p); err != nil {
		t.Fatal("other session's spool must survive:", err)
	}
}

func TestT53_SessionStartDropsExpiredSpool(t *testing.T) {
	st := t.TempDir()
	p := spoolFile(t, st, "sess-1", "OLD DIGEST")
	old := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(p, old, old); err != nil {
		t.Fatal(err)
	}
	so, _, code := runStdin(t, st, hookPayload("sess-1", "", "SessionStart"), "hook", "session-start")
	eq(t, "exit", code, 0)
	eq(t, "silent", so, "")
	_, err := os.Stat(p)
	eq(t, "expired spool deleted", os.IsNotExist(err), true)
}
