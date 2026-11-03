package ingest

import (
	"database/sql"
	"fmt"
	"math"

	_ "modernc.org/sqlite"

	"agentmem/internal/model"
)

// ParseHermes reads user/assistant/tool rows of ~/.hermes/state.db style db at path (read-only)
// with messages.id > afterID, ordered by id. It returns the messages and the highest row id seen.
func ParseHermes(path string, afterID int64) (msgs []model.Message, maxID int64, err error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT m.id, m.session_id, m.role, m.content, s.source, s.cwd, m.timestamp, m._compressed_summary
FROM messages m JOIN sessions s ON s.id=m.session_id WHERE m.id > ? ORDER BY m.id`, afterID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id         int64
			sessionID  string
			role       string
			content    sql.NullString
			source     string
			cwd        sql.NullString
			timestamp  float64
			compressed int
		)
		if err := rows.Scan(&id, &sessionID, &role, &content, &source, &cwd, &timestamp, &compressed); err != nil {
			return nil, 0, err
		}
		// maxID counts every row seen, including the ones skipped below.
		if id > maxID {
			maxID = id
		}
		if role != "user" && role != "assistant" {
			continue
		}
		if !content.Valid || content.String == "" || compressed != 0 {
			continue
		}
		msgs = append(msgs, model.Message{
			Agent:       "hermes",
			SourceStore: "hermes/" + source,
			SessionID:   sessionID,
			SourcePath:  path,
			SourceRef:   fmt.Sprintf("%d#0", id),
			Seq:         int(id) * 100,
			TS:          int64(math.Round(timestamp * 1000)),
			Role:        role,
			Project:     cwd.String,
			Text:        content.String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return msgs, maxID, nil
}
