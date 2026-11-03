package e2e

import "testing"

func TestT08_RedactTokens(t *testing.T) {
	m := texts(parsed(t, "claude", "redacted", "redact/secrets.jsonl"))
	eq(t, "count", len(m), 6)
	eq(t, "aws", m[0], "my key is [REDACTED:aws] ok")
	eq(t, "github", m[1], "token [REDACTED:github]")
	eq(t, "jwt", m[2], "jwt [REDACTED:jwt] done")
	eq(t, "private key", m[3], "[REDACTED:private-key] after")
	eq(t, "clean line", m[5], "nothing secret here, just the flag --max-old-space-size=4096")
}

func TestT08_RedactIsOnlyAtRedactedStage(t *testing.T) {
	m := texts(parsed(t, "claude", "filtered", "redact/secrets.jsonl"))
	has(t, m[0], "AKIAIOSFODNN7EXAMPLE")
}
