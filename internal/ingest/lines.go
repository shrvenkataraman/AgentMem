// Package ingest turns harness transcript stores into model.Message values.
package ingest

import (
	"bytes"
	"os"
)

// Line is one complete JSONL line. No is 1-based.
type Line struct {
	No   int
	Data []byte
}

// ReadLines returns the complete (newline-terminated) lines of path that start at byte offset
// and the offset just after the last complete line. A torn final line is not returned and the
// offset does not advance past it. fromLine is the 1-based number of the first line at offset.
func ReadLines(path string, offset int64, fromLine int) ([]Line, int64, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, offset, err
	}
	if offset > int64(len(b)) {
		offset = int64(len(b))
	}
	var out []Line
	no := fromLine
	for rest := b[offset:]; ; {
		i := bytes.IndexByte(rest, '\n')
		if i < 0 {
			break
		}
		out = append(out, Line{No: no, Data: rest[:i]})
		no++
		offset += int64(i + 1)
		rest = rest[i+1:]
	}
	return out, offset, nil
}
