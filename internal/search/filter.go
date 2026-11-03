package search

import "strings"

// Filter narrows candidates. The zero value means no filtering.
type Filter struct {
	Project        string // exact match on chunks.project
	Agent          string // exact match on chunks.agent
	SinceMS        int64  // ts >= SinceMS when > 0
	ExcludeSession string // session_id != ExcludeSession when non-empty
}

// FilterSQL returns a clause to append to a WHERE (leading " AND ...", or "") and its args.
func FilterSQL(f Filter) (string, []any) {
	var (
		sb   strings.Builder
		args []any
	)
	add := func(cond string, v any) {
		sb.WriteString(cond)
		args = append(args, v)
	}
	if f.Project != "" {
		add(" AND c.project = ?", f.Project)
	}
	if f.Agent != "" {
		add(" AND c.agent = ?", f.Agent)
	}
	if f.SinceMS > 0 {
		add(" AND c.ts >= ?", f.SinceMS)
	}
	if f.ExcludeSession != "" {
		add(" AND c.session_id != ?", f.ExcludeSession)
	}
	return sb.String(), args
}
