package search

// Filter narrows candidates. The zero value means no filtering.
type Filter struct {
	Project        string // exact match on chunks.project
	Agent          string // exact match on chunks.agent
	SinceMS        int64  // ts >= SinceMS when > 0
	ExcludeSession string // session_id != ExcludeSession when non-empty
}

// FilterSQL returns a clause to append to a WHERE (leading " AND ...", or "") and its args.
// Stub: applies nothing until the "filters" ticket.
func FilterSQL(f Filter) (string, []any) { return "", nil }
