package ingest

import "agentmem/internal/model"

// ParseClaudeLine converts one Claude Code JSONL line to messages. skipped is true when the
// line is unparseable or not a user/assistant line.
func ParseClaudeLine(path string, l Line) (msgs []model.Message, skipped bool) {
	panic("not implemented")
}
