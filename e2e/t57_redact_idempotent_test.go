package e2e

import "testing"

func TestT57_MarkersAreStable(t *testing.T) {
	same := []string{
		"TOKEN=[REDACTED:secret]; FOO=1",
		"`API_KEY=[REDACTED:secret]`, then go",
		"KEY_SECRET=[REDACTED:secret])",
	}
	got := redactedTexts(t, append(append([]string{}, same...), "TOKEN=abc;")...)
	for i := range same {
		eq(t, "stable "+string(rune('0'+i)), got[i], same[i])
	}
	eq(t, "fresh", got[len(same)], "TOKEN=[REDACTED:secret]")
}
