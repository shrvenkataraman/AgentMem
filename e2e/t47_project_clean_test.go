package e2e

import (
	"strings"
	"testing"
)

func TestT47_NonGitPathIsCleaned(t *testing.T) {
	st := t.TempDir()
	for in, want := range map[string]string{
		"/a/b/../c":       "/a/c",
		"/a//b/./c/":      "/a/b/c",
		"/home/u/proj///": "/home/u/proj",
	} {
		eq(t, "project "+in, strings.TrimSpace(ok(t, st, nil, "project", in)), want)
	}
}
