package e2e

import (
	"strings"
	"testing"
	"time"
)

func TestT60_GitleaksSecretsRedacted(t *testing.T) {
	anth := "sk-ant-api03-" + strings.Repeat("Ab3dE", 18) + "fGh" + "AA" // 93 chars then AA
	secrets := []string{
		"xoxb-1234567890123-1234567890123-AbCdEfGhIjKlMnOpQrStUvWx",
		"sk_live_4eC39HqLyjWDarjtT1zdp7dc",
		anth,
		"AIzaSyD8xK3mP9qL2wR7tY1nB6cV4zX0aS5dF8g",
	}
	var in []string
	for _, s := range secrets {
		in = append(in, "config value "+s+" here")
	}
	got := redactedTexts(t, in...)
	for i, s := range secrets {
		lacks(t, got[i], s)
		has(t, got[i], "[REDACTED:")
	}
	again := redactedTexts(t, got...)
	for i := range got {
		eq(t, "idempotent", again[i], got[i])
	}
}

func TestT60_GitleaksFastEnough(t *testing.T) {
	prose := strings.Repeat("the quick brown fox jumps over the lazy dog and keeps running ", 16000) // ~1 MB
	start := time.Now()
	got := redactedTexts(t, prose)
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("redacting 1 MB took %v", d)
	}
	lacks(t, got[0], "[REDACTED:")
}
