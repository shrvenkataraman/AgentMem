package e2e

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestT59_ConvertsGitleaksToml(t *testing.T) {
	dir := t.TempDir()
	in, out := filepath.Join(dir, "g.toml"), filepath.Join(dir, "g.json")
	writeFile(t, in, `title = "x"
[[rules]]
id = "slack-bot-token"
regex = '''(xoxb-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*)'''
keywords = ["XOXB"]

[[rules]]
id = "pkcs12-file"
path = '''(?i)\.p12$'''

[[rules]]
id = "generic-api-key"
regex = '''(?i)key\s*[:=]\s*([a-z0-9]{16,})'''
secretGroup = 1
entropy = 3.5
`)
	cmd := exec.Command("python3", "-I", "../scripts/gitleaks_to_json.py", in, out)
	so, err := cmd.Output()
	if err != nil {
		t.Fatalf("script failed: %v", err)
	}
	eq(t, "stdout", strings.TrimSpace(string(so)), "kept=2 dropped=1")

	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var rules []struct {
		ID          string   `json:"id"`
		Regex       string   `json:"regex"`
		SecretGroup int      `json:"secretGroup"`
		Entropy     float64  `json:"entropy"`
		Keywords    []string `json:"keywords"`
	}
	if err := json.Unmarshal(b, &rules); err != nil {
		t.Fatal(err)
	}
	eq(t, "count", len(rules), 2)
	eq(t, "id0", rules[0].ID, "slack-bot-token")
	eq(t, "kw0", strings.Join(rules[0].Keywords, ","), "xoxb")
	eq(t, "group0", rules[0].SecretGroup, 0)
	eq(t, "entropy0", rules[0].Entropy, 0.0)
	eq(t, "id1", rules[1].ID, "generic-api-key")
	eq(t, "group1", rules[1].SecretGroup, 1)
	eq(t, "entropy1", rules[1].Entropy, 3.5)

	if code := exec.Command("python3", "-I", "../scripts/gitleaks_to_json.py", in).Run(); code == nil {
		t.Fatal("one argument should exit non-zero")
	}
}
