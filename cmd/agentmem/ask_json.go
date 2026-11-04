package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"agentmem/internal/search"
)

// jsonHit is one ask result in machine-readable form.
type jsonHit struct {
	ID         int64   `json:"id"`
	TS         string  `json:"ts"`
	Agent      string  `json:"agent"`
	Project    string  `json:"project"`
	Session    string  `json:"session"`
	Role       string  `json:"role"`
	ToolName   string  `json:"tool_name"`
	Seq        int     `json:"seq"`
	Snippet    string  `json:"snippet"`
	Provenance string  `json:"provenance"` // SourcePath + ":" + SourceRef
	Coverage   float64 `json:"coverage"`   // rounded to 2 decimals
	SeenIn     int     `json:"seen_in"`    // max(1, h.SeenIn)
}

// jsonOut is the single object `ask --json` prints.
type jsonOut struct {
	Query    string    `json:"query"`
	Results  []jsonHit `json:"results"`
	Stale    bool      `json:"stale"`
	Excluded string    `json:"excluded_session"`
}

// printJSON prints one machine-readable object: the query, the first `limit` ranked hits (never
// null), the stale flag and the session excluded from the search.
func printJSON(q string, hits []search.Hit, limit int, excluded string, stale bool) {
	if limit < 0 {
		limit = 0
	}
	if len(hits) > limit {
		hits = hits[:limit]
	}
	results := make([]jsonHit, len(hits)) // never nil, so an empty result set prints as []
	for i, h := range hits {
		seen := h.SeenIn
		if seen < 1 {
			seen = 1
		}
		results[i] = jsonHit{
			ID:         h.ID,
			TS:         time.UnixMilli(h.TS).UTC().Format("2006-01-02T15:04:05Z"),
			Agent:      h.Agent,
			Project:    h.Project,
			Session:    h.SessionID,
			Role:       h.Role,
			ToolName:   h.ToolName,
			Seq:        h.Seq,
			Snippet:    search.Snippet(h.Text),
			Provenance: h.SourcePath + ":" + h.SourceRef,
			Coverage:   math.Round(h.Coverage*100) / 100,
			SeenIn:     seen,
		}
	}
	b, err := json.Marshal(jsonOut{Query: q, Results: results, Stale: stale, Excluded: excluded})
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem ask:", err)
		return
	}
	fmt.Println(string(b))
}
