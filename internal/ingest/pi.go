package ingest

import (
	"encoding/json"
	"fmt"
	"time"

	"agentmem/internal/model"
)

// PiParser carries the session header (id, cwd) seen on the first line of a pi JSONL file.
type PiParser struct {
	Path      string
	SessionID string
	Cwd       string
}

// piLine is one pi JSONL line. Content is raw because a message line may carry a plain string
// (system lines) or a block list (user/assistant lines).
type piLine struct {
	Type      string `json:"type"`
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Cwd       string `json:"cwd"`
	Message   struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// piBlock is one entry of a message line's content list.
type piBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ParseLine converts one pi JSONL line to messages. The "session" header line sets SessionID
// and Cwd and yields no messages.
func (p *PiParser) ParseLine(l Line) (msgs []model.Message, skipped bool) {
	var ln piLine
	if err := json.Unmarshal(l.Data, &ln); err != nil {
		return nil, true
	}
	switch ln.Type {
	case "session":
		p.SessionID = ln.ID
		p.Cwd = ln.Cwd
		return nil, false
	case "message":
		if ln.Message.Role != "user" && ln.Message.Role != "assistant" {
			return nil, false
		}
		var blocks []piBlock
		if err := json.Unmarshal(ln.Message.Content, &blocks); err != nil {
			return nil, false
		}
		var ts int64
		if t, err := time.Parse(time.RFC3339Nano, ln.Timestamp); err == nil {
			ts = t.UnixMilli()
		}
		for idx, b := range blocks {
			if b.Type != "text" {
				continue
			}
			msgs = append(msgs, model.Message{
				Agent:       "pi",
				SourceStore: "pi",
				SessionID:   p.SessionID,
				SourcePath:  p.Path,
				SourceRef:   fmt.Sprintf("%d#%d", l.No, idx),
				Seq:         l.No*100 + idx,
				TS:          ts,
				Role:        ln.Message.Role,
				Project:     p.Cwd,
				Text:        b.Text,
			})
		}
		return msgs, false
	default:
		return nil, false
	}
}
