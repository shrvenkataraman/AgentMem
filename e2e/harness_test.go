// Package e2e drives the real agentmem binary against fixtures in a temp state dir.
package e2e

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	return runEnv(t, state, nil, args...)
}

// runEnv is run with extra KEY=VALUE env entries (later entries win). The three store roots
// default to nonexistent paths so a test can never read the real ~/.claude, ~/.pi or ~/.hermes.
func runEnv(t *testing.T, state string, env []string, args ...string) (string, string, int) {
	return runFull(t, state, env, "", args...)
}

// runStdin is run with the given stdin text.
func runStdin(t *testing.T, state, stdin string, args ...string) (string, string, int) {
	return runFull(t, state, nil, stdin, args...)
}

func runFull(t *testing.T, state string, env []string, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin(t), args...)
	cmd.Env = append(os.Environ(), "AGENTMEM_STATE_DIR="+state,
		"AGENTMEM_CLAUDE_DIR="+state+"/none-claude", "AGENTMEM_PI_DIR="+state+"/none-pi",
		"AGENTMEM_HERMES_DB="+state+"/none-hermes.db", "AGENTMEM_SESSION=",
		"CLAUDE_CODE_SESSION_ID=", "AGENTMEM_AGENT=", "AGENTMEM_CWD=/nonexistent/cwd")
	cmd.Env = append(cmd.Env, env...)
	cmd.Stdin = strings.NewReader(stdin)
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
		dec := json.NewDecoder(strings.NewReader(ln))
		dec.UseNumber()
		if err := dec.Decode(&m); err != nil {
			t.Fatalf("bad json line %q: %v", ln, err)
		}
		out = append(out, m)
	}
	return out
}

// corpus returns env entries pointing the three store roots at testdata/corpus.
func corpus() []string {
	return []string{"AGENTMEM_CLAUDE_DIR=" + fixture("corpus/claude"), "AGENTMEM_PI_DIR=" + fixture("corpus/pi"),
		"AGENTMEM_HERMES_DB=" + fixture("corpus/hermes/state.db")}
}

// eq fails unless fmt.Sprint(got) == fmt.Sprint(want). JSON numbers decode as json.Number, so
// integers compare exactly.
func eq(t *testing.T, name string, got, want any) {
	t.Helper()
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("%s: got %v, want %v", name, got, want)
	}
}

func has(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("output lacks %q:\n%s", sub, s)
	}
}

func lacks(t *testing.T, s, sub string) {
	t.Helper()
	if strings.Contains(s, sub) {
		t.Fatalf("output must not contain %q:\n%s", sub, s)
	}
}

// ok runs the binary and fails unless it exits 0; returns stdout.
func ok(t *testing.T, state string, env []string, args ...string) string {
	t.Helper()
	so, se, code := runEnv(t, state, env, args...)
	if code != 0 {
		t.Fatalf("agentmem %v exited %d\nstderr: %s", args, code, se)
	}
	return so
}

// parsed runs `agentmem parse` on a testdata file and returns the JSON objects.
func parsed(t *testing.T, store, stage, rel string) []map[string]any {
	t.Helper()
	return jsonLines(t, ok(t, t.TempDir(), nil, "parse", "--store", store, "--stage", stage, fixture(rel)))
}

// openIdx opens the index in state read-write for direct SQL assertions.
func openIdx(t *testing.T, state string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+state+"/index.sqlite")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// count runs a single-int SELECT.
func count(t *testing.T, db *sql.DB, q string, args ...any) int {
	t.Helper()
	var n int
	if err := db.QueryRow(q, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// corpusState returns a fresh state dir after `agentmem ingest` over testdata/corpus.
func corpusState(t *testing.T) string {
	t.Helper()
	st := t.TempDir()
	ok(t, st, corpus(), "ingest")
	return st
}

// statsOf returns `agentmem stats --json` as a map.
func statsOf(t *testing.T, state string, env []string) map[string]any {
	t.Helper()
	return jsonLines(t, ok(t, state, env, "stats", "--json"))[0]
}

// asks runs `ask --json` (extra flags first, then the query) and returns results.
func asks(t *testing.T, state string, env []string, args ...string) []map[string]any {
	t.Helper()
	m := jsonLines(t, ok(t, state, env, append([]string{"ask", "--json"}, args...)...))[0]
	r, _ := m["results"].([]any)
	var out []map[string]any
	for _, x := range r {
		out = append(out, x.(map[string]any))
	}
	return out
}

// copyTree copies directory src to dst.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		rel, _ := filepath.Rel(src, p)
		if info.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		b, _ := os.ReadFile(p)
		return os.WriteFile(filepath.Join(dst, rel), b, 0o644)
	})
}

// texts returns the "text" field of each decoded object.
func texts(m []map[string]any) []string {
	var o []string
	for _, x := range m {
		o = append(o, x["text"].(string))
	}
	return o
}

// sessions returns the distinct "session" (ask --json) or "session_id" (raw) values, sorted.
func sessionsOf(m []map[string]any) []string {
	seen := map[string]bool{}
	var o []string
	for _, x := range m {
		s, _ := x["session"].(string)
		if s == "" {
			s, _ = x["session_id"].(string)
		}
		if !seen[s] {
			seen[s] = true
			o = append(o, s)
		}
	}
	sort.Strings(o)
	return o
}

// rawLines runs `ask --raw` and returns the non-empty stdout lines (id, session, coverage TSV).
func rawLines(t *testing.T, state string, env []string, args ...string) []string {
	t.Helper()
	out := ok(t, state, env, append([]string{"ask", "--raw"}, args...)...)
	var o []string
	for _, l := range strings.Split(out, "\n") {
		if l != "" {
			o = append(o, l)
		}
	}
	return o
}

// rawField returns tab-separated field i of each raw line.
func rawField(lines []string, i int) []string {
	var o []string
	for _, l := range lines {
		o = append(o, strings.Split(l, "\t")[i])
	}
	return o
}

// fixLines returns the lines of a testdata file, each with its trailing newline.
func fixLines(t *testing.T, rel string) []string {
	t.Helper()
	b, err := os.ReadFile(fixture(rel))
	if err != nil {
		t.Fatal(err)
	}
	var o []string
	for _, l := range strings.SplitAfter(string(b), "\n") {
		if l != "" {
			o = append(o, l)
		}
	}
	return o
}

// writeFile writes s to path, replacing it.
func writeFile(t *testing.T, path, s string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
		t.Fatal(err)
	}
}

// appendFile appends s to path.
func appendFile(t *testing.T, path, s string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
}

// cursorOf returns cursors.cursor for (store, key), or "" when absent.
func cursorOf(t *testing.T, db *sql.DB, store, key string) string {
	t.Helper()
	var c string
	db.QueryRow("SELECT cursor FROM cursors WHERE store=? AND key=?", store, key).Scan(&c)
	return c
}

// tmpCorpus copies testdata/corpus/{claude,pi} into a temp dir and returns env entries pointing at
// the copies (hermes stays the fixture) plus the temp root, for tests that mutate sources.
func tmpCorpus(t *testing.T) ([]string, string) {
	t.Helper()
	d := t.TempDir()
	copyTree(t, fixture("corpus/claude"), d+"/claude")
	copyTree(t, fixture("corpus/pi"), d+"/pi")
	return []string{"AGENTMEM_CLAUDE_DIR=" + d + "/claude", "AGENTMEM_PI_DIR=" + d + "/pi",
		"AGENTMEM_HERMES_DB=" + fixture("corpus/hermes/state.db")}, d
}

// fm builds a normalized message object (the JSON shape `parse` prints) for digest stdin.
func fm(seq int, role, text, toolID, toolName string, isErr bool) map[string]any {
	m := map[string]any{"agent": "claude", "source_store": "claude", "session_id": "s", "source_path": "/f",
		"source_ref": fmt.Sprintf("%d#0", seq), "seq": seq * 100, "ts": 1790848800000 + seq, "role": role, "project": "/p", "text": text}
	if toolID != "" {
		m["tool_id"], m["tool_name"] = toolID, toolName
	}
	if isErr {
		m["is_error"] = true
	}
	return m
}

// jl encodes objects as JSON lines.
func jl(ms ...map[string]any) string {
	var sb strings.Builder
	for _, m := range ms {
		b, _ := json.Marshal(m)
		sb.Write(b)
		sb.WriteByte('\n')
	}
	return sb.String()
}
