package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir, "-c", "user.email=a@b", "-c", "user.name=n"}, args...)...)
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestT12_ProjectNormalization(t *testing.T) {
	root, _ := filepath.EvalSymlinks(t.TempDir())
	repo := filepath.Join(root, "repo")
	os.MkdirAll(filepath.Join(repo, "sub"), 0o755)
	gitIn(t, repo, "init", "-q")
	gitIn(t, repo, "commit", "-q", "--allow-empty", "-m", "x")
	wt := filepath.Join(root, "wt")
	gitIn(t, repo, "worktree", "add", "-q", wt)
	st := t.TempDir()
	for _, in := range []string{repo, filepath.Join(repo, "sub"), wt} {
		eq(t, "repo path "+in, strings.TrimSpace(ok(t, st, nil, "project", in)), repo)
	}
	eq(t, "non-git path", strings.TrimSpace(ok(t, st, nil, "project", "/home/u/Work/proj/")), "/home/u/Work/proj")
	eq(t, "empty", strings.TrimSpace(ok(t, st, nil, "project", "")), "")
}
