package main

import (
	"database/sql"
	"fmt"
	"os"
	"sort"

	"agentmem/internal/model"
)

// sessionMsgs is one session's messages, ordered by seq, plus the fields needed to label it.
type sessionMsgs struct {
	ID      string
	Agent   string
	Project string
	First   int64 // earliest ts in milliseconds
	Msgs    []model.Message
}

// loadSessions reads the indexed chunks of sessions whose messages have ts >= since (0 = no lower
// bound), restricted to the given session id when non-empty. Chunks are read in
// (session_id, seq, window_idx) order and consecutive chunks sharing a (session_id, seq) are joined
// into one message (texts concatenated, no separator). The result is ordered by first ts, then id.
func loadSessions(db *sql.DB, since int64, session string) []sessionMsgs {
	q := `SELECT session_id, source_store, agent, project, seq, ts, role, text, tool_id, tool_name, is_error
		FROM chunks WHERE ts >= ?`
	args := []any{since}
	if session != "" {
		q += ` AND session_id = ?`
		args = append(args, session)
	}
	q += ` ORDER BY session_id, seq, window_idx`
	rows, err := db.Query(q, args...)
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentmem handoff:", err)
		return nil
	}
	defer rows.Close()
	var ids []string
	groups := make(map[string]*sessionMsgs)
	for rows.Next() {
		var (
			id, store, agent, project, role, text, toolID, toolName string
			seq                                                     int
			ts                                                      int64
			isErr                                                   int
		)
		if err := rows.Scan(&id, &store, &agent, &project, &seq, &ts, &role, &text,
			&toolID, &toolName, &isErr); err != nil {
			fmt.Fprintln(os.Stderr, "agentmem handoff:", err)
			return nil
		}
		s := groups[id]
		if s == nil {
			s = &sessionMsgs{ID: id, Agent: agent, Project: project, First: ts}
			groups[id] = s
			ids = append(ids, id)
		}
		if ts < s.First {
			s.First = ts
		}
		if n := len(s.Msgs); n > 0 && s.Msgs[n-1].Seq == seq {
			s.Msgs[n-1].Text += text
			continue
		}
		s.Msgs = append(s.Msgs, model.Message{
			Agent: agent, SourceStore: store, SessionID: id, Seq: seq, TS: ts, Role: role,
			Project: project, Text: text, ToolID: toolID, ToolName: toolName, IsError: isErr != 0,
		})
	}
	if err := rows.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "agentmem handoff:", err)
		return nil
	}
	out := make([]sessionMsgs, 0, len(ids))
	for _, id := range ids {
		out = append(out, *groups[id])
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].First != out[j].First {
			return out[i].First < out[j].First
		}
		return out[i].ID < out[j].ID
	})
	return out
}
