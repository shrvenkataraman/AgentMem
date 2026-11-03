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

// AddSkipped adds n to the skipped counter of the (store, key) cursor row.
func AddSkipped(db *sql.DB, store, key string, n int) error {
	_, err := db.Exec(`UPDATE cursors SET skipped = skipped + ? WHERE store=? AND key=?`, n, store, key)
	return err
}

// SkippedByStore returns the total skipped lines per store, omitting stores with zero.
func SkippedByStore(db *sql.DB) map[string]int {
	res := map[string]int{}
	rows, err := db.Query("SELECT store, sum(skipped) FROM cursors WHERE skipped > 0 GROUP BY store")
	if err != nil {
		return res
	}
	defer rows.Close()
	for rows.Next() {
		var s string
		var n int
		if err := rows.Scan(&s, &n); err != nil {
			return res
		}
		res[s] = n
	}
	return res
}
