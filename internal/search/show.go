package search

import (
	"database/sql"
)

// Get returns the chunk with the given id. sql.ErrNoRows passes through when there is none.
func Get(db *sql.DB, id int64) (Hit, error) {
	var (
		h     Hit
		isErr int
	)
	err := db.QueryRow("SELECT "+Cols+" FROM chunks c WHERE c.id = ?", id).
		Scan(&h.ID, &h.SessionID, &h.SourceStore, &h.Agent, &h.Project, &h.SourcePath,
			&h.SourceRef, &h.Seq, &h.WindowIdx, &h.TS, &h.Role, &h.Text, &h.ContentHash,
			&isErr, &h.ToolID, &h.ToolName)
	if err != nil {
		return Hit{}, err
	}
	h.IsError = isErr != 0
	return h, nil
}

// Neighbors returns the chunks of h.SessionID in (seq, window_idx) order: the n before h, h itself
// and the n after. Fewer are returned at the edges of the session. When h is not in the session it
// is returned alone.
func Neighbors(db *sql.DB, h Hit, n int) ([]Hit, error) {
	rows, err := db.Query("SELECT "+Cols+" FROM chunks c WHERE c.session_id = ?"+
		" ORDER BY c.seq, c.window_idx", h.SessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var all []Hit
	for rows.Next() {
		var (
			g     Hit
			isErr int
		)
		if err := rows.Scan(&g.ID, &g.SessionID, &g.SourceStore, &g.Agent, &g.Project, &g.SourcePath,
			&g.SourceRef, &g.Seq, &g.WindowIdx, &g.TS, &g.Role, &g.Text, &g.ContentHash,
			&isErr, &g.ToolID, &g.ToolName); err != nil {
			return nil, err
		}
		g.IsError = isErr != 0
		all = append(all, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	pos := -1
	for i, g := range all {
		if g.ID == h.ID {
			pos = i
			break
		}
	}
	if pos < 0 {
		return []Hit{h}, nil
	}
	from, to := pos-n, pos+n+1
	if from < 0 {
		from = 0
	}
	if to > len(all) {
		to = len(all)
	}
	return all[from:to], nil
}
