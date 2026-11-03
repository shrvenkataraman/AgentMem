// Package facts extracts deterministic structured facts from normalized messages.
package facts

import (
	"strings"
	"unicode/utf8"

	"agentmem/internal/model"
)

// Command is one shell command seen in the transcript with its outcome.
type Command struct {
	Cmd     string `json:"cmd"`
	Outcome string `json:"outcome"`
	Output  string `json:"output"`
}

// Facts is the structured digest of a transcript.
type Facts struct {
	Task     string    `json:"task"`
	Commands []Command `json:"commands"`
	Files    []string  `json:"files"`
	Open     []string  `json:"open"`
}

// commandTools are the tool names that run a shell command.
var commandTools = map[string]bool{"Bash": true, "bash": true, "terminal": true, "shell": true}

const (
	taskBytes   = 300
	outputBytes = 120
)

// Extract returns the task and shell commands of msgs. Files and Open are empty in this ticket.
func Extract(msgs []model.Message) Facts {
	f := Facts{Commands: []Command{}, Files: []string{}, Open: []string{}}
	results := make(map[string]model.Message)
	haveTask := false
	for _, m := range msgs {
		if m.Role != "tool_result" {
			continue
		}
		if _, seen := results[m.ToolID]; !seen {
			results[m.ToolID] = m
		}
	}
	for _, m := range msgs {
		switch m.Role {
		case "user":
			if !haveTask {
				f.Task = cutBytes(m.Text, taskBytes)
				haveTask = true
			}
		case "tool_call":
			if !commandTools[m.ToolName] {
				continue
			}
			c := Command{Cmd: m.Text, Outcome: "unknown"}
			r, ok := results[m.ToolID]
			if ok {
				c.Outcome = "ok"
				if r.IsError {
					c.Outcome = "fail"
				}
				c.Output = cutBytes(firstLine(r.Text), outputBytes)
			}
			f.Commands = append(f.Commands, c)
		}
	}
	return f
}

// firstLine returns the first non-empty line of s.
func firstLine(s string) string {
	for _, ln := range strings.Split(s, "\n") {
		if strings.TrimSpace(ln) != "" {
			return strings.TrimSpace(ln)
		}
	}
	return ""
}

// cutBytes returns s capped to n bytes, never splitting a rune.
func cutBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	b := []byte(s[:n])
	for len(b) > 0 {
		r, size := utf8.DecodeLastRune(b)
		if r == utf8.RuneError && size <= 1 {
			b = b[:len(b)-1]
			continue
		}
		break
	}
	return string(b)
}
