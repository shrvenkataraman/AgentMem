package e2e

import (
	"fmt"
	"testing"
)

func TestT44_CleanCorpus(t *testing.T) {
	st := corpusState(t)
	eq(t, "clean", ok(t, st, nil, "audit"), "leaks=0 suspects=0\n")
}

func TestT44_FindsLeaksAndSuspectsWithoutPrintingThem(t *testing.T) {
	st := corpusState(t)
	db := openIdx(t, st)
	leak, err := db.Exec(insChunk, "leaky", "1#0", "key AKIAIOSFODNN7EXAMPLE leaked")
	if err != nil {
		t.Fatal(err)
	}
	susp, err := db.Exec(insChunk, "suspicious", "1#0", "tok x9Zq3Lm8Vb2Nc7Rt5Yw1Kd4Hf6Jg0Ps9Aa3Ee7UuZ1Qq end")
	if err != nil {
		t.Fatal(err)
	}
	lid, _ := leak.LastInsertId()
	sid, _ := susp.LastInsertId()
	so, _, code := runEnv(t, st, nil, "audit")
	eq(t, "exit", code, 1)
	has(t, so, "leaks=1 suspects=1\n")
	has(t, so, fmt.Sprintf("leak %d\n", lid))
	has(t, so, fmt.Sprintf("suspect %d x9Zq…Qq\n", sid))
	lacks(t, so, "AKIAIOSFODNN7EXAMPLE")
	lacks(t, so, "x9Zq3Lm8Vb2Nc7Rt5Yw1Kd4Hf6Jg0Ps9Aa3Ee7UuZ1Qq")
}

func TestT44_SuspectsOnlyExitZero(t *testing.T) {
	st := corpusState(t)
	openIdx(t, st).Exec(insChunk, "suspicious", "1#0", "tok x9Zq3Lm8Vb2Nc7Rt5Yw1Kd4Hf6Jg0Ps9Aa3Ee7UuZ1Qq end")
	_, _, code := runEnv(t, st, nil, "audit")
	eq(t, "suspects alone do not fail", code, 0)
}
