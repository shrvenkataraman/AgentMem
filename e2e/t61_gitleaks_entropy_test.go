package e2e

import "testing"

func TestT61_LowEntropyKeptHighEntropyRedacted(t *testing.T) {
	low := `api_key: "aaaaaaaaaaaaaaaaaaaaaaaa"`
	high := `api_key: "Zq8vK2mX9pL4wR7tY1nB6cD3"`
	got := redactedTexts(t, low, high)
	eq(t, "low entropy", got[0], low)
	lacks(t, got[1], "Zq8vK2mX9pL4wR7tY1nB6cD3")
	has(t, got[1], "[REDACTED:")
}
