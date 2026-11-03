package e2e

import (
	"strings"
	"testing"
)

func TestT23_CoverageAndOrder(t *testing.T) {
	st := corpusState(t)
	l := rawLines(t, st, nil, "ERR_FOO_BAR", "build")
	eq(t, "hits", len(l), 8)
	eq(t, "coverage", strings.Join(rawField(l, 2), ","), "1.00,0.50,0.50,0.50,0.50,0.50,0.50,0.50")
	eq(t, "session order", strings.Join(rawField(l, 1), ","), "sc-a,sc-d,sc-c,sc-c,sc-b,sc-a,sc-a,sc-a")
}

func TestT23_SingleTermFullCoverage(t *testing.T) {
	st := corpusState(t)
	l := rawLines(t, st, nil, "deploy-prod")
	eq(t, "coverage", strings.Join(rawField(l, 2), ","), "1.00,1.00")
}
