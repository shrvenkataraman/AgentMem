package ingest

import (
	"encoding/json"
	"fmt"
	"strings"
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
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	ToolUseID string          `json:"tool_use_id"`
	Content   json.RawMessage `json:"content"`
	IsError   bool            `json:"is_error"`
}

// toolCallText renders the text of a tool_use block.
func toolCallText(name string, input json.RawMessage) string {
	if len(input) == 0 {
		return name
	}
	var m map[string]any
	if err := json.Unmarshal(input, &m); err != nil {
		return name + " " + string(input)
	}
	if name == "Bash" {
		v, ok := m["command"]
		if !ok {
			return ""
		}
		if s, ok := v.(string); ok {
			return s
		}
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
	b, err := json.Marshal(m)
	if err != nil {
		return name + " " + string(input)
	}
	return name + " " + string(b)
}

// toolResultText renders the text of a tool_result block content.
func toolResultText(content json.RawMessage) string {
	if len(content) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(content, &s); err == nil {
		return s
	}
	var blocks []claudeBlock
	if err := json.Unmarshal(content, &blocks); err != nil {
		return ""
	}
	var parts []string
	for _, b := range blocks {
		if b.Type == "text" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n")
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
		if b.Type != "text" && b.Type != "tool_use" && b.Type != "tool_result" {
			continue
		}
		msg := model.Message{
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
		}
		switch b.Type {
		case "tool_use":
			msg.Role = "tool_call"
			msg.ToolID = b.ID
			msg.ToolName = b.Name
			msg.Text = toolCallText(b.Name, b.Input)
		case "tool_result":
			msg.Role = "tool_result"
			msg.ToolID = b.ToolUseID
			msg.ToolName = ""
			msg.IsError = b.IsError
			msg.Text = toolResultText(b.Content)
		}
		msgs = append(msgs, msg)
	}
	return msgs, false
}
