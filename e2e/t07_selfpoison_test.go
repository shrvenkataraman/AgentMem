package e2e

import (
	"strings"
	"testing"
)

func TestT07_SelfPoisonFiltered(t *testing.T) {
	c := parsed(t, "claude", "filtered", "claude/s1.jsonl")
	eq(t, "claude count", len(c), 7)
	for _, x := range texts(c) {
		if x == "poison result" || x == "agentmem ask 'foo'" {
			t.Fatalf("poison survived: %q", x)
		}
	}
	eq(t, "pi count", len(parsed(t, "pi", "filtered", "pi/s1.jsonl")), 6)
	eq(t, "hermes count", len(parsed(t, "hermes", "filtered", "hermes/state.db")), 6)
}

func TestT07_SelfPoisonCommandForms(t *testing.T) {
	m := parsed(t, "claude", "raw", "poison/cases.jsonl")
	eq(t, "raw count", len(m), 10)
	f := texts(parsed(t, "claude", "filtered", "poison/cases.jsonl"))
	eq(t, "filtered", strings.Join(f, "|"), "grep agentmem notes.txt|result-3|echo agentmem|result-4")
}
