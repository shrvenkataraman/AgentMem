package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestT32_LagSeconds(t *testing.T) {
	env, root := tmpCorpus(t)
	mt := time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC)
	filepath.Walk(root, func(p string, i os.FileInfo, err error) error {
		if err == nil && !i.IsDir() {
			os.Chtimes(p, mt, mt)
		}
		return nil
	})
	st := t.TempDir()
	ok(t, st, env, "ingest")
	lag := statsOf(t, st, env)["lag_seconds"].(map[string]any)
	eq(t, "claude", lag["claude"], 486000)
	eq(t, "pi", lag["pi"], 655198)
}

func TestT32_LagNeverNegative(t *testing.T) {
	env, root := tmpCorpus(t)
	mt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	filepath.Walk(root, func(p string, i os.FileInfo, err error) error {
		if err == nil && !i.IsDir() {
			os.Chtimes(p, mt, mt)
		}
		return nil
	})
	st := t.TempDir()
	ok(t, st, env, "ingest")
	eq(t, "claude", statsOf(t, st, env)["lag_seconds"].(map[string]any)["claude"], 0)
}
