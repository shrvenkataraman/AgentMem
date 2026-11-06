package ingest

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"

	_ "modernc.org/sqlite"

	"agentmem/internal/model"
)

// ParseHermes reads user/assistant/tool rows of ~/.hermes/state.db style db at path (read-only)
// with messages.id > afterID, ordered by id. It returns the messages and the highest row id seen.
func ParseHermes(path string, afterID int64) (msgs []model.Message, maxID int64, err error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT m.id, m.session_id, m.role, m.content, s.source, s.cwd, m.timestamp, m._compressed_summary, m.tool_call_id, m.tool_calls, m.tool_name
FROM messages m JOIN sessions s ON s.id=m.session_id WHERE m.id > ? ORDER BY m.id`, afterID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id         int64
			sessionID  string
			role       string
			content    sql.NullString
			source     string
			cwd        sql.NullString
			timestamp  float64
			compressed int
			toolCallID sql.NullString
			toolCalls  sql.NullString
			toolName   sql.NullString
		)
		if err := rows.Scan(&id, &sessionID, &role, &content, &source, &cwd, &timestamp, &compressed,
			&toolCallID, &toolCalls, &toolName); err != nil {
			return nil, 0, err
		}
		// maxID counts every row seen, including the ones skipped below.
		if id > maxID {
			maxID = id
		}
		if role != "user" && role != "assistant" && role != "tool" {
			continue
		}
		if compressed != 0 {
			continue
		}
		base := model.Message{
			Agent:       "hermes",
			SourceStore: "hermes/" + source,
			SessionID:   sessionID,
			SourcePath:  path,
			TS:          int64(math.Round(timestamp * 1000)),
			Project:     cwd.String,
		}
		if role == "tool" {
			if !content.Valid || content.String == "" {
				continue
			}
			m := base
			m.SourceRef = fmt.Sprintf("%d#0", id)
			m.Seq = int(id) * 100
			m.Role = "tool_result"
			m.ToolID = toolCallID.String
			m.ToolName = toolName.String
			m.Text = content.String
			msgs = append(msgs, m)
			continue
		}
		if content.Valid && content.String != "" {
			m := base
			m.SourceRef = fmt.Sprintf("%d#0", id)
			m.Seq = int(id) * 100
			m.Role = role
			m.Text = content.String
			msgs = append(msgs, m)
		}
		if role != "assistant" || !toolCalls.Valid || toolCalls.String == "" {
			continue
		}
		var calls []hermesToolCall
		if err := json.Unmarshal([]byte(toolCalls.String), &calls); err != nil {
			return nil, 0, err
		}
		for k, c := range calls {
			idx := 1 + k
			m := base
			m.SourceRef = fmt.Sprintf("%d#%d", id, idx)
			m.Seq = int(id)*100 + idx
			m.Role = "tool_call"
			m.ToolID = c.ID
			m.ToolName = c.Function.Name
			m.Text = hermesCallText(c)
			msgs = append(msgs, m)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return msgs, maxID, nil
}

type hermesToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// hermesCallText renders a tool call: terminal shows its command, everything else "<name> <arguments>".
func hermesCallText(c hermesToolCall) string {
	if c.Function.Name == "terminal" {
		var a struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(c.Function.Arguments), &a); err == nil && a.Command != "" {
			return a.Command
		}
	}
	return c.Function.Name + " " + c.Function.Arguments
}
