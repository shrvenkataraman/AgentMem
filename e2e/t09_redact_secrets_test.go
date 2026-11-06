package e2e

import "testing"

func TestT09_RedactAssignedSecrets(t *testing.T) {
	m := texts(parsed(t, "claude", "redacted", "redact/secrets.jsonl"))
	eq(t, "assigned secret", m[4], "export API_TOKEN=[REDACTED:secret] then run")
	eq(t, "clean line", m[5], "nothing secret here, just the flag --max-old-space-size=4096")
	eq(t, "aws still redacted", m[0], "my key is [REDACTED:aws] ok")
	eq(t, "private key still redacted", m[3], "[REDACTED:private-key] after")
}
