package search

import (
	"database/sql"
	"sort"
	"strings"
)

// Collapse removes chunks whose ContentHash repeats, so identical text from several sessions is
// shown once. In each group the kept member is the one with the smallest TS (ties broken by the
// smallest ID); it takes the position of the group's first (best ranked) member, and SeenIn is set
// to the number of distinct SessionID values in the group.
func Collapse(hits []Hit) []Hit {
	type group struct {
		first    int
		best     int
		sessions map[string]bool
	}
	groups := make(map[string]*group, len(hits))
	for i, h := range hits {
		g := groups[h.ContentHash]
		if g == nil {
			g = &group{first: i, best: i, sessions: map[string]bool{}}
			groups[h.ContentHash] = g
		}
		g.sessions[h.SessionID] = true
		if h.TS < hits[g.best].TS || (h.TS == hits[g.best].TS && h.ID < hits[g.best].ID) {
			g.best = i
		}
	}
	out := make([]Hit, 0, len(hits))
	for i, h := range hits {
		g := groups[h.ContentHash]
		if g.first != i {
			continue
		}
		keep := hits[g.best]
		keep.SeenIn = len(g.sessions)
		out = append(out, keep)
	}
	return out
}

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
