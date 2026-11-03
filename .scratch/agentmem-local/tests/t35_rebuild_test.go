package e2e

import (
	"os"
	"testing"
)

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func TestT35_RebuildSwaps(t *testing.T) {
	st := corpusState(t)
	eq(t, "out", ok(t, st, corpus(), "rebuild"), "rebuilt old=21 new=21 swapped=true\n")
	if !exists(st + "/index.sqlite.bak") {
		t.Fatal("previous index must be kept as index.sqlite.bak")
	}
	if exists(st + "/index.new.sqlite") {
		t.Fatal("index.new.sqlite must not remain")
	}
	eq(t, "chunks", statsOf(t, st, nil)["chunks"], 21)
	eq(t, "queryable", len(asks(t, st, nil, "--no-refresh", "ERR_FOO_BAR")), 6)
}

func TestT35_RegressionGuardRefusesSmallerIndex(t *testing.T) {
	env, root := tmpCorpus(t)
	st := t.TempDir()
	ok(t, st, env, "ingest")
	sB := root + "/claude/-home-u-Work-other/sB.jsonl"
	L, _ := os.ReadFile(sB)
	n := 0
	for i, c := range L {
		if c == '\n' {
			if n++; n == 2 {
				os.WriteFile(sB, L[:i+1], 0o644)
				break
			}
		}
	}
	so, _, code := runEnv(t, st, env, "rebuild")
	eq(t, "refused exit", code, 1)
	eq(t, "refused out", so, "rebuilt old=21 new=19 swapped=false\n")
	eq(t, "index unchanged", statsOf(t, st, nil)["chunks"], 21)
	if exists(st+"/index.sqlite.bak") || exists(st+"/index.new.sqlite") {
		t.Fatal("a refused rebuild must leave no .bak or .new file")
	}
	eq(t, "force", ok(t, st, env, "rebuild", "--force"), "rebuilt old=21 new=19 swapped=true\n")
	eq(t, "forced chunks", statsOf(t, st, nil)["chunks"], 19)
	if !exists(st + "/index.sqlite.bak") {
		t.Fatal("forced rebuild must keep .bak")
	}
}
