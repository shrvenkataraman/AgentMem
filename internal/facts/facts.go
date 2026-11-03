// Package facts extracts deterministic structured facts from normalized messages.
package facts

import (
	"encoding/json"
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

// fileTools are the lower-cased tool names that touch a file path.
var fileTools = map[string]bool{"read": true, "edit": true, "write": true, "multiedit": true, "notebookedit": true}

const (
	taskBytes   = 300
	outputBytes = 120
	openCount   = 3
	openBytes   = 200
)

// Extract returns the task, shell commands, touched files and trailing open messages of msgs.
func Extract(msgs []model.Message) Facts {
	f := Facts{Commands: []Command{}, Files: []string{}, Open: []string{}}
	seenFiles := make(map[string]bool)
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
			if commandTools[m.ToolName] {
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
			if fileTools[strings.ToLower(m.ToolName)] {
				if p := filePath(m.Text); p != "" {
					if _, seen := seenFiles[p]; !seen {
						seenFiles[p] = true
						f.Files = append(f.Files, p)
					}
				}
			}
		}
	}
	var opens []string
	for _, m := range msgs {
		if m.Role != "user" {
			continue
		}
		opens = append(opens, cutBytes(m.Text, openBytes))
		if len(opens) > openCount {
			opens = opens[len(opens)-openCount:]
		}
	}
	if opens != nil {
		f.Open = opens
	}
	return f
}

// filePath returns the file path of a file tool call text: the JSON after the first space, field
// "file_path", else "path". It returns "" when there is no space, the JSON is invalid, or the path
// is empty.
func filePath(text string) string {
	i := strings.IndexByte(text, ' ')
	if i < 0 {
		return ""
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(text[i+1:]), &v); err != nil {
		return ""
	}
	for _, k := range []string{"file_path", "path"} {
		if s, ok := v[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
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
