package ingest

import (
	"path/filepath"
	"strings"

	"agentmem/internal/model"
)

// shellTools are the tool names whose Text is a shell command; only these can carry an agentmem
// invocation. Non-shell tools (Read, Write, ...) are never dropped.
var shellTools = map[string]bool{"Bash": true, "bash": true, "terminal": true, "shell": true}

// DropSelfPoison removes tool_call messages that invoke the agentmem binary together with the
// tool_result messages paired to them by (SessionID, ToolID). known holds the pairing keys
// (SessionID+"\x00"+ToolID) of already-dropped calls: it is read and updated in place, so a
// result that arrives in a later batch than its call is still dropped.
func DropSelfPoison(msgs []model.Message, known map[string]bool) []model.Message {
	// dropped holds the keys of the calls dropped during this call; it is always consulted so the
	// pairing works even when known is nil. known, when non-nil, is additionally updated in place
	// so the pairing persists across batches.
	dropped := make(map[string]bool)
	out := make([]model.Message, 0, len(msgs))
	for _, m := range msgs {
		key := m.SessionID + "\x00" + m.ToolID
		switch m.Role {
		case "tool_call":
			if shellTools[m.ToolName] && isAgentmemInvocation(m.Text) {
				dropped[key] = true
				if known != nil {
					known[key] = true
				}
				continue
			}
		case "tool_result":
			if dropped[key] || (known != nil && known[key]) {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

// isAgentmemInvocation reports whether cmd runs the agentmem binary as a program. The command is
// split into segments on &&, ||, ;, | and newline; in each segment leading env assignments (words
// containing = that do not start with /) are skipped and the first remaining word is the program.
func isAgentmemInvocation(cmd string) bool {
	for _, seg := range strings.FieldsFunc(cmd, func(r rune) bool {
		return r == '&' || r == '|' || r == ';' || r == '\n'
	}) {
		fields := strings.Fields(seg)
		i := 0
		for i < len(fields) && strings.Contains(fields[i], "=") && !strings.HasPrefix(fields[i], "/") {
			i++
		}
		if i < len(fields) && filepath.Base(fields[i]) == "agentmem" {
			return true
		}
	}
	return false
}
