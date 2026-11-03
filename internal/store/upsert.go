package store

import (
	"database/sql"

	"agentmem/internal/model"
)

// Upsert inserts or updates chunks in one transaction, keyed on
// (session_id, source_ref, window_idx). It returns the number of newly inserted rows
// (count after - count before); updated rows are not counted.
func Upsert(db *sql.DB, cs []model.Chunk) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var before int
	if err := tx.QueryRow("SELECT count(*) FROM chunks").Scan(&before); err != nil {
		return 0, err
	}
	for _, c := range cs {
		isErr := 0
		if c.IsError {
			isErr = 1
		}
		_, err := tx.Exec(`INSERT INTO chunks(session_id,source_store,agent,project,source_path,source_ref,seq,window_idx,ts,role,text,content_hash,is_error,tool_id,tool_name)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(session_id,source_ref,window_idx) DO UPDATE SET
 text=excluded.text, content_hash=excluded.content_hash, ts=excluded.ts, seq=excluded.seq, project=excluded.project`,
			c.SessionID, c.SourceStore, c.Agent, c.Project, c.SourcePath, c.SourceRef, c.Seq, c.WindowIdx,
			c.TS, c.Role, c.Text, c.ContentHash, isErr, c.ToolID, c.ToolName)
		if err != nil {
			return 0, err
		}
	}
	var after int
	if err := tx.QueryRow("SELECT count(*) FROM chunks").Scan(&after); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return after - before, nil
}
