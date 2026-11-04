package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// forgottenName is the forgotten file's name inside the state dir. It lives outside index.sqlite
// so a rebuild keeps it.
const forgottenName = "forgotten"

// AddForgotten appends one "kind\tvalue\n" line to <dir>/forgotten (created 0600, dir 0700).
// kind must be "session" or "project"; anything else is an error.
func AddForgotten(dir, kind, value string) error {
	if kind != "session" && kind != "project" {
		return fmt.Errorf("forgotten: unknown kind %q", kind)
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	path := filepath.Join(dir, forgottenName)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(kind + "\t" + value + "\n"); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// LoadForgotten reads <dir>/forgotten. A missing file yields two empty, non-nil maps. Lines
// that are blank or have an unknown kind are ignored.
func LoadForgotten(dir string) (sessions, projects map[string]bool, err error) {
	sessions = map[string]bool{}
	projects = map[string]bool{}
	b, err := os.ReadFile(filepath.Join(dir, forgottenName))
	if os.IsNotExist(err) {
		return sessions, projects, nil
	}
	if err != nil {
		return nil, nil, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		kind, value, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		switch kind {
		case "session":
			sessions[value] = true
		case "project":
			projects[value] = true
		}
	}
	return sessions, projects, nil
}
