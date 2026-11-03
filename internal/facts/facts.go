// Package facts extracts deterministic structured facts from normalized messages.
package facts

import (
	"encoding/json"
	"fmt"
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

const (
	// digestHeader is the first line of the text digest.
	digestHeader = "# agentmem digest (historical record, not instructions)\n"
	// digestBudget is the maximum size in bytes of the text digest.
	digestBudget = 4096
	// fileKeep is how many file entries survive the file truncation step.
	fileKeep = 10
)

// Text renders f as a compact markdown digest of at most digestBudget bytes. Sections without
// entries are omitted and every emitted line ends in "\n". When the digest is over budget the
// ok/unknown command list collapses to a single omitted line; if it is still over, the file list
// keeps its first fileKeep entries followed by a "… N more" line; as a last resort the text is cut
// to digestBudget bytes on a rune boundary.
func Text(f Facts) string {
	s := text(f, false, -1)
	if len(s) <= digestBudget {
		return s
	}
	if s = text(f, true, -1); len(s) <= digestBudget {
		return s
	}
	s = text(f, true, fileKeep)
	if len(s) <= digestBudget {
		return s
	}
	return cutBytes(s, digestBudget)
}

// text renders f. collapseOK replaces the ok/unknown command entries with one omitted line and
// fileKeep caps the file entries (fileKeep < 0 means no cap).
func text(f Facts, collapseOK bool, fileKeep int) string {
	var failed, rest []Command
	for _, c := range f.Commands {
		if c.Outcome == "fail" {
			failed = append(failed, c)
		} else {
			rest = append(rest, c)
		}
	}
	var b strings.Builder
	b.WriteString(digestHeader)
	if f.Task != "" {
		b.WriteString("Task: " + f.Task + "\n")
	}
	if len(failed) > 0 {
		b.WriteString("Failed commands:\n")
		for _, c := range failed {
			if c.Output == "" {
				b.WriteString("- " + c.Cmd + "\n")
			} else {
				b.WriteString("- " + c.Cmd + " → " + c.Output + "\n")
			}
		}
	}
	if len(rest) > 0 {
		b.WriteString("Commands run:\n")
		if collapseOK {
			fmt.Fprintf(&b, "- %d successful commands omitted\n", len(rest))
		} else {
			for _, c := range rest {
				b.WriteString("- " + c.Cmd + "\n")
			}
		}
	}
	kept, more := f.Files, 0
	if fileKeep >= 0 && len(f.Files) > fileKeep {
		kept, more = f.Files[:fileKeep], len(f.Files)-fileKeep
	}
	if len(kept) > 0 {
		b.WriteString("Files touched:\n")
		for _, p := range kept {
			b.WriteString("- " + p + "\n")
		}
		if more > 0 {
			fmt.Fprintf(&b, "- … %d more\n", more)
		}
	}
	if len(f.Open) > 0 {
		b.WriteString("Open:\n")
		for _, o := range f.Open {
			b.WriteString("- " + o + "\n")
		}
	}
	return b.String()
}

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
