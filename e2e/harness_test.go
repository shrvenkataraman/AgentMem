// Package e2e drives the real agentmem binary against fixtures in a temp state dir.
package e2e

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	buildOnce sync.Once
	binPath   string
	buildErr  error
)

// bin builds the agentmem binary once per test run and returns its path.
func bin(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, _ := os.MkdirTemp("", "agentmem-bin")
		binPath = filepath.Join(dir, "agentmem")
		cmd := exec.Command("go", "build", "-o", binPath, "./cmd/agentmem")
		cmd.Dir = ".."
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0", "GOFLAGS=-mod=vendor")
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildErr = &buildError{string(out)}
		}
	})
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	return binPath
}

type buildError struct{ out string }

func (e *buildError) Error() string { return "go build failed:\n" + e.out }

// run executes the binary with AGENTMEM_STATE_DIR=state and returns stdout, stderr and exit code.
func run(t *testing.T, state string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin(t), args...)
	cmd.Env = append(os.Environ(), "AGENTMEM_STATE_DIR="+state)
	var so, se bytes.Buffer
	cmd.Stdout, cmd.Stderr = &so, &se
	err := cmd.Run()
	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return so.String(), se.String(), code
}

// fixture returns the path of a file under ../testdata.
func fixture(rel string) string {
	p, _ := filepath.Abs(filepath.Join("..", "testdata", rel))
	return p
}

// jsonLines decodes newline-delimited JSON objects.
func jsonLines(t *testing.T, s string) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, ln := range strings.Split(strings.TrimSpace(s), "\n") {
		if ln == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(ln), &m); err != nil {
			t.Fatalf("bad json line %q: %v", ln, err)
		}
		out = append(out, m)
	}
	return out
}
