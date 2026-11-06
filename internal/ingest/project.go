package ingest

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// NormalizeProject maps a recorded cwd to its repository root key (see ticket "project").
// A cwd inside a git worktree maps to the repository's main working directory.
func NormalizeProject(raw string) string {
	if raw == "" {
		return ""
	}
	out, err := exec.Command("git", "-C", raw, "rev-parse", "--path-format=absolute", "--git-common-dir").Output()
	if err == nil {
		common := strings.TrimSpace(string(out))
		if strings.HasSuffix(common, "/.git") {
			return filepath.Dir(common)
		}
	}
	return filepath.Clean(raw)
}
