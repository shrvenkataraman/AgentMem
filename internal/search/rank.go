package search

import (
	"database/sql"
	"sort"
	"strings"
)

// Query runs q against the index and returns the reranked hits.
func Query(db *sql.DB, q string, f Filter) ([]Hit, error) {
	terms := Terms(q)
	cands, err := Candidates(db, terms, f, 512)
	if err != nil {
		return nil, err
	}
	return Rerank(cands, terms), nil
}

// maxHits is how many reranked hits Query returns.
const maxHits = 128

// Rerank scores each hit by the fraction of terms whose lowercase form appears in the lowercase
// hit text, then orders by coverage desc, ts desc, id asc, and keeps the first 128. It sets
// h.Coverage on the returned hits.
func Rerank(hits []Hit, terms []string) []Hit {
	lower := make([]string, len(terms))
	for i, t := range terms {
		lower[i] = strings.ToLower(t)
	}
	out := make([]Hit, len(hits))
	copy(out, hits)
	for i := range out {
		text := strings.ToLower(out[i].Text)
		var n int
		for _, t := range lower {
			if strings.Contains(text, t) {
				n++
			}
		}
		if len(terms) > 0 {
			out[i].Coverage = float64(n) / float64(len(terms))
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Coverage != out[j].Coverage {
			return out[i].Coverage > out[j].Coverage
		}
		if out[i].TS != out[j].TS {
			return out[i].TS > out[j].TS
		}
		return out[i].ID < out[j].ID
	})
	if len(out) > maxHits {
		out = out[:maxHits]
	}
	return out
}
