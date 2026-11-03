package search

import (
	"database/sql"
	"strings"
	"unicode/utf8"
)

// Terms splits q into query terms, de-duplicated keeping the first occurrence of each.
func Terms(q string) []string {
	var out []string
	seen := map[string]bool{}
	for _, t := range strings.Fields(q) {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

// FTSQuery builds an FTS5 MATCH expression from terms: each term quoted (so FTS operators in the
// user's query are plain text) and joined with OR.
func FTSQuery(terms []string) string {
	q := make([]string, len(terms))
	for i, t := range terms {
		q[i] = `"` + strings.ReplaceAll(t, `"`, `""`) + `"`
	}
	return strings.Join(q, " OR ")
}

// Candidates returns the chunks matching any term, best rank first, at most limit of them.
// No terms means no candidates and no error.
func Candidates(db *sql.DB, terms []string, f Filter, limit int) ([]Hit, error) {
	if len(terms) == 0 {
		return nil, nil
	}
	where, fargs := FilterSQL(f)

	query := func(table, match string) ([]Hit, error) {
		args := append([]any{match}, fargs...)
		args = append(args, limit)
		rows, err := db.Query("SELECT "+Cols+" FROM "+table+" JOIN chunks c ON c.id = "+table+".rowid"+
			" WHERE "+table+" MATCH ?"+where+" ORDER BY rank LIMIT ?", args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var out []Hit
		for rows.Next() {
			var (
				h     Hit
				isErr int
			)
			if err := rows.Scan(&h.ID, &h.SessionID, &h.SourceStore, &h.Agent, &h.Project, &h.SourcePath,
				&h.SourceRef, &h.Seq, &h.WindowIdx, &h.TS, &h.Role, &h.Text, &h.ContentHash,
				&isErr, &h.ToolID, &h.ToolName); err != nil {
				return nil, err
			}
			h.IsError = isErr != 0
			out = append(out, h)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return out, nil
	}

	out, err := query("chunks_fts", FTSQuery(terms))
	if err != nil {
		return nil, err
	}

	seen := make(map[int64]bool, len(out))
	for _, h := range out {
		seen[h.ID] = true
	}
	for _, t := range terms {
		if utf8.RuneCountInString(t) < 3 {
			continue
		}
		trgm, err := query("chunks_trgm", `"`+strings.ReplaceAll(t, `"`, `""`)+`"`)
		if err != nil {
			return nil, err
		}
		for _, h := range trgm {
			if !seen[h.ID] {
				seen[h.ID] = true
				out = append(out, h)
			}
		}
	}
	return out, nil
}
