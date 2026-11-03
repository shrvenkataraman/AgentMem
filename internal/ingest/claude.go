package ingest

import (
	"encoding/json"
	"fmt"
	"time"

	"agentmem/internal/model"
)

// claudeLine is the JSONL envelope written by Claude Code.
type claudeLine struct {
	Type        string `json:"type"`
	SessionID   string `json:"sessionId"`
	Timestamp   string `json:"timestamp"`
	Cwd         string `json:"cwd"`
	IsSidechain bool   `json:"isSidechain"`
	Message     struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

// claudeBlock is one entry of a content list (or the only entry of a string content).
type claudeBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ParseClaudeLine converts one Claude Code JSONL line to messages. skipped is true when the
// line is unparseable or not a user/assistant line.
func ParseClaudeLine(path string, l Line) (msgs []model.Message, skipped bool) {
	var cl claudeLine
	if err := json.Unmarshal(l.Data, &cl); err != nil {
		return nil, true
	}
	if cl.Type != "user" && cl.Type != "assistant" {
		return nil, true
	}

	var blocks []claudeBlock
	if len(cl.Message.Content) > 0 {
		var s string
		if err := json.Unmarshal(cl.Message.Content, &s); err == nil {
			blocks = []claudeBlock{{Type: "text", Text: s}}
		} else if err := json.Unmarshal(cl.Message.Content, &blocks); err != nil {
			return nil, true
		}
	}

	ts, _ := time.Parse(time.RFC3339Nano, cl.Timestamp)
	for idx, b := range blocks {
		if b.Type != "text" {
			continue
		}
		msgs = append(msgs, model.Message{
			Agent:       "claude",
			SourceStore: "claude",
			SessionID:   cl.SessionID,
			SourcePath:  path,
			SourceRef:   fmt.Sprintf("%d#%d", l.No, idx),
			Seq:         l.No*100 + idx,
			TS:          ts.UnixMilli(),
			Role:        cl.Message.Role,
			Project:     cl.Cwd,
			Text:        b.Text,
			Sidechain:   cl.IsSidechain,
		})
	}
	return msgs, false
}
