package ingest

import "agentmem/internal/model"

// PiParser carries the session header (id, cwd) seen on the first line of a pi JSONL file.
type PiParser struct {
	Path      string
	SessionID string
	Cwd       string
}

// ParseLine converts one pi JSONL line to messages. The "session" header line sets SessionID
// and Cwd and yields no messages.
func (p *PiParser) ParseLine(l Line) (msgs []model.Message, skipped bool) {
	panic("not implemented")
}
