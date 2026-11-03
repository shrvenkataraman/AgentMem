package e2e

import "testing"

func TestT50_TokenCountersAreNotSecrets(t *testing.T) {
	got := redactedTexts(t,
		"max_tokens=4096 input_tokens=12",
		"GH_TOKEN=123456",
		"GH_TOKENS=abc",
	)
	want := []string{
		"max_tokens=4096 input_tokens=12",
		"GH_TOKEN=[REDACTED:secret]",
		"GH_TOKENS=[REDACTED:secret]",
	}
	for i := range want {
		eq(t, "case "+string(rune('0'+i)), got[i], want[i])
	}
}
