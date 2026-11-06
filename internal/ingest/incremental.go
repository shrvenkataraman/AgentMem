package ingest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"agentmem/internal/model"
	"agentmem/internal/store"
)

// poisonStore is the cursor store name under which the self-poison pairing keys are carried
// across runs.
const poisonStore = "poison"

// LoadKnown returns the poison pairing keys persisted for key. An absent or malformed cursor
// yields an empty map.
func LoadKnown(db *sql.DB, key string) map[string]bool {
	known := map[string]bool{}
	c, ok := store.GetCursor(db, poisonStore, key)
	if !ok {
		return known
	}
	var keys []string
	if err := json.Unmarshal([]byte(c), &keys); err != nil {
		return known
	}
	for _, k := range keys {
		known[k] = true
	}
	return known
}

// SaveKnown persists the poison pairing keys for key as a JSON array of sorted keys.
func SaveKnown(db *sql.DB, key string, known map[string]bool) error {
	keys := make([]string, 0, len(known))
	for k := range known {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b, err := json.Marshal(keys)
	if err != nil {
		return err
	}
	return store.SetCursor(db, poisonStore, key, string(b))
}

// parseCursor splits "<byteOffset>:<nextLineNo>". An invalid value falls back to the default "0:1".
func parseCursor(s string) (int64, int) {
	off, no := int64(0), 1
	f := strings.SplitN(s, ":", 2)
	if len(f) != 2 {
		return off, no
	}
	o, err := strconv.ParseInt(f[0], 10, 64)
	if err != nil {
		return off, no
	}
	n, err := strconv.Atoi(f[1])
	if err != nil || n < 1 {
		return off, no
	}
	return o, n
}

// IngestJSONL ingests only the lines of path added since the last run. The cursor
// "<byteOffset>:<nextLineNo>" is kept under (storeName, path); a file shorter than the cursor
// offset is treated as truncated and re-read from the start. It returns the number of newly
// inserted chunks and the number of skipped lines.
func IngestJSONL(db *sql.DB, storeName, path string) (inserted, skipped int, err error) {
	offset, nextLine := int64(0), 1
	if c, ok := store.GetCursor(db, storeName, path); ok {
		offset, nextLine = parseCursor(c)
	}
	fi, err := os.Stat(path)
	if err != nil {
		return 0, 0, err
	}
	if fi.Size() < offset {
		offset, nextLine = 0, 1
	}
	lines, newOff, err := ReadLines(path, offset, nextLine)
	if err != nil {
		return 0, 0, err
	}

	var msgs []model.Message
	switch storeName {
	case "claude":
		for _, l := range lines {
			m, s := ParseClaudeLine(path, l)
			skipped += boolInt(s)
			msgs = append(msgs, m...)
		}
	case "pi":
		p := &PiParser{Path: path}
		if offset > 0 {
			first, _, err := ReadLines(path, 0, 1)
			if err != nil {
				return 0, 0, err
			}
			if len(first) > 0 {
				p.ParseLine(first[0])
			}
		}
		for _, l := range lines {
			m, s := p.ParseLine(l)
			skipped += boolInt(s)
			msgs = append(msgs, m...)
		}
	default:
		return 0, 0, fmt.Errorf("unknown store %q", storeName)
	}

	known := LoadKnown(db, path)
	inserted, err = store.Upsert(db, Process(msgs, known))
	if err != nil {
		return 0, skipped, err
	}
	if len(known) > 0 {
		if err := SaveKnown(db, path, known); err != nil {
			return inserted, skipped, err
		}
	}
	if err := store.SetCursor(db, storeName, path, fmt.Sprintf("%d:%d", newOff, nextLine+len(lines))); err != nil {
		return inserted, skipped, err
	}
	if err := store.AddSkipped(db, storeName, path, skipped); err != nil {
		return inserted, skipped, err
	}
	return inserted, skipped, nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
