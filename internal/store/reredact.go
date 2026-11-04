package store

import (
	"database/sql"

	"agentmem/internal/redact"
)

// SchemaVersion is PRAGMA user_version of an index written by this build. Bump it whenever
// redaction rules change so existing rows are re-redacted on next open.
const SchemaVersion = 2

// Reredact runs redact.Text over every chunk and updates the rows whose text changed, in one
// transaction (the chunks_au trigger keeps both FTS tables in sync). It returns how many rows changed.
func Reredact(db *sql.DB) (int, error) {
	type row struct {
		id   int64
		text string
	}
	rows, err := db.Query("SELECT id, text FROM chunks")
	if err != nil {
		return 0, err
	}
	var all []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.text); err != nil {
			rows.Close()
			return 0, err
		}
		all = append(all, r)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	changed := 0
	for _, r := range all {
		new := redact.Text(r.text)
		if new == r.text {
			continue
		}
		if _, err := tx.Exec("UPDATE chunks SET text=? WHERE id=?", new, r.id); err != nil {
			return 0, err
		}
		changed++
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return changed, nil
}
