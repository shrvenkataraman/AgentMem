package e2e

import (
	"os"
	"testing"
)

func TestT36_UnionKeepsRowsWhoseSourceIsGone(t *testing.T) {
	env, root := tmpCorpus(t)
	st := t.TempDir()
	ok(t, st, env, "ingest")
	os.Remove(root + "/claude/-home-u-Work-proj/sD.jsonl")
	eq(t, "rebuild", ok(t, st, env, "rebuild"), "rebuilt old=21 new=21 swapped=true\n")
	eq(t, "sc-d survives", len(rawLines(t, st, nil, "talks")), 1)
	eq(t, "again", ok(t, st, env, "rebuild"), "rebuilt old=21 new=21 swapped=true\n")
	eq(t, "no duplicates", statsOf(t, st, nil)["chunks"], 21)
}

func TestT36_ForceSkipsUnion(t *testing.T) {
	env, root := tmpCorpus(t)
	st := t.TempDir()
	ok(t, st, env, "ingest")
	os.Remove(root + "/claude/-home-u-Work-proj/sD.jsonl")
	eq(t, "force", ok(t, st, env, "rebuild", "--force"), "rebuilt old=21 new=20 swapped=true\n")
	eq(t, "sc-d purged", len(rawLines(t, st, nil, "talks")), 0)
}
