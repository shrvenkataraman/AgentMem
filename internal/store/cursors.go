package store

import (
	"database/sql"
)

// GetCursor returns cursors.cursor for (store, key) and whether a row exists.
func GetCursor(db *sql.DB, store, key string) (string, bool) {
	var c string
	if err := db.QueryRow("SELECT cursor FROM cursors WHERE store=? AND key=?", store, key).Scan(&c); err != nil {
		return "", false
	}
	return c, true
}

// SetCursor upserts the cursor row for (store, key).
func SetCursor(db *sql.DB, store, key, cursor string) error {
	_, err := db.Exec(`INSERT INTO cursors(store,key,cursor) VALUES(?,?,?)
ON CONFLICT(store,key) DO UPDATE SET cursor=excluded.cursor`, store, key, cursor)
	return err
}
