package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// countChunks returns SELECT count(*) FROM chunks at path, or 0 on any error (missing file,
// missing table, unreadable db). It opens read-only so it never creates -wal/-shm files.
func countChunks(path string) int {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return 0
	}
	defer db.Close()
	var n int
	if err := db.QueryRow("SELECT count(*) FROM chunks").Scan(&n); err != nil {
		return 0
	}
	return n
}

// removeDB removes a database file and its -wal/-shm siblings.
func removeDB(path string) {
	os.Remove(path)
	os.Remove(path + "-wal")
	os.Remove(path + "-shm")
}

// moveDB renames path to dst together with any -wal/-shm siblings so a sidecar WAL travels with
// its main file. The destination (and its sidecars) are replaced.
func moveDB(dst, path string) error {
	if err := os.Rename(path, dst); err != nil {
		return err
	}
	for _, suf := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(path + suf); err == nil {
			os.Remove(dst + suf)
			os.Rename(path+suf, dst+suf)
		}
	}
	return nil
}

// unionCols lists every chunks column except the id, in schema order.
const unionCols = "session_id,source_store,agent,project,source_path,source_ref,seq,window_idx,ts,role," +
	"text,content_hash,is_error,tool_id,tool_name"

// keepVanished copies rows out of the currently installed index (dir/index.sqlite) into db for each
// of its source_path values that no longer exists on disk, so a rebuild does not lose chunks whose
// source file was deleted. It is a no-op when the old index does not exist. Rows already present in
// db are skipped by INSERT OR IGNORE on the chunks unique key.
func keepVanished(db *sql.DB, dir string) error {
	old := filepath.Join(dir, "index.sqlite")
	if _, err := os.Stat(old); err != nil {
		return nil
	}
	if _, err := db.Exec("ATTACH DATABASE ? AS old", old); err != nil {
		return fmt.Errorf("rebuild: attach old index: %w", err)
	}
	defer db.Exec("DETACH old")

	var paths []string
	rows, err := db.Query("SELECT DISTINCT source_path FROM old.chunks")
	if err != nil {
		return fmt.Errorf("rebuild: read old source paths: %w", err)
	}
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			rows.Close()
			return fmt.Errorf("rebuild: scan old source path: %w", err)
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("rebuild: old source paths: %w", err)
	}
	rows.Close()

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			continue // the file still exists, so build has already indexed it
		}
		if _, err := db.Exec("INSERT OR IGNORE INTO chunks("+unionCols+
			") SELECT "+unionCols+" FROM old.chunks WHERE source_path = ?", p); err != nil {
			return fmt.Errorf("rebuild: keep chunks of %s: %w", p, err)
		}
	}
	return nil
}

// Rebuild builds a fresh index beside the old one in dir and swaps it in only when it is not
// smaller than the old index (unless force). It returns the old and new chunk counts and whether
// the swap happened. On a refused swap the candidate is removed and no .bak is made; otherwise
// the old index is kept as index.sqlite.bak.
func Rebuild(dir string, force bool, build func(db *sql.DB) error) (oldN, newN int, swapped bool, err error) {
	old := filepath.Join(dir, "index.sqlite")
	newPath := filepath.Join(dir, "index.new.sqlite")

	oldN = countChunks(old)

	removeDB(newPath)
	db, err := OpenFile(newPath)
	if err != nil {
		return oldN, 0, false, err
	}
	if err := build(db); err != nil {
		db.Close()
		removeDB(newPath)
		return oldN, 0, false, err
	}
	if !force {
		if err := keepVanished(db, dir); err != nil {
			db.Close()
			removeDB(newPath)
			return oldN, 0, false, err
		}
	}
	newN = countChunks(newPath)
	db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
	if err := db.Close(); err != nil {
		removeDB(newPath)
		return oldN, newN, false, err
	}

	if newN < oldN && !force {
		removeDB(newPath)
		return oldN, newN, false, nil
	}

	bak := old + ".bak"
	removeDB(bak)
	if _, err := os.Stat(old); err == nil {
		if err := moveDB(bak, old); err != nil {
			removeDB(newPath)
			return oldN, newN, false, fmt.Errorf("rebuild: %w", err)
		}
	}
	removeDB(old) // stale -wal/-shm of the old index
	if err := moveDB(old, newPath); err != nil {
		return oldN, newN, false, fmt.Errorf("rebuild: %w", err)
	}
	return oldN, newN, true, nil
}

// TryIngestLock takes the non-blocking ingest lock for dir (dir/ingest.lock, an exclusive flock).
// The returned release unlocks and closes the file. It errors with "ingest lock busy" when another
// process holds it. The name is deliberately not TryLock: ticket 31 owns the exported
// store.TryLock(dir, wait) in internal/store/lock.go, and two declarations of that name in one
// package would not compile (PROGRESS.md defect I).
func TryIngestLock(dir string) (release func(), err error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(filepath.Join(dir, "ingest.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		return nil, errors.New("ingest lock busy")
	}
	return func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}, nil
}
