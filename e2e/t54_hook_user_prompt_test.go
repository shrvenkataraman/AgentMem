package e2e

import (
	"os"
	"testing"
)

func TestT54_UserPromptFallbackDeliversOnce(t *testing.T) {
	st := t.TempDir()
	p := spoolFile(t, st, "sess-1", "DIGEST BODY")
	so, _, code := runStdin(t, st, hookPayload("sess-1", "", "UserPromptSubmit"), "hook", "user-prompt")
	eq(t, "exit", code, 0)
	ev, ctx := injected(t, so)
	eq(t, "event", ev, "UserPromptSubmit")
	has(t, ctx, "DIGEST BODY")
	_, err := os.Stat(p)
	eq(t, "spool deleted", os.IsNotExist(err), true)
	so, _, _ = runStdin(t, st, hookPayload("sess-1", "", "UserPromptSubmit"), "hook", "user-prompt")
	eq(t, "second call is silent", so, "")
}

func TestT54_UserPromptIgnoresOtherSessions(t *testing.T) {
	st := t.TempDir()
	spoolFile(t, st, "sess-1", "DIGEST BODY")
	so, _, _ := runStdin(t, st, hookPayload("sess-2", "", "UserPromptSubmit"), "hook", "user-prompt")
	eq(t, "silent", so, "")
}
