# AgentMem

AgentMem gives coding agents a searchable memory of their past sessions. It indexes transcripts from Claude Code, pi and Hermes into one local SQLite file. An agent then runs `agentmem ask` to find an earlier decision, command or error message.

- **Local only:** AgentMem has no server and no cloud service.
- **Deterministic:** no language model runs when AgentMem writes the index or builds a digest.
- **One static binary:** the Go build needs no CGO, and the dependencies are vendored.

## Install

Go 1.27 or later is necessary. AgentMem is tested on Linux only.

```bash
git clone https://github.com/shrvenkataraman/AgentMem.git
```

```bash
cd AgentMem && ./scripts/install.sh
```

The script installs `agentmem` to `~/.local/bin`. To use a different folder, give the folder as the first argument.

## Quick start

Index your past sessions. An estimate for a first run is about 1 minute and 130 MB, from one test corpus of 37,000 chunks:

```bash
agentmem ingest
```

Search them:

```bash
agentmem ask "ERR_FOO_BAR"
```

Show one full result, with the chunk id from the search output:

```bash
agentmem show 123
```

## Connect Claude Code

These two commands change your Claude Code configuration. Add `--remove` to either command to undo its change.

Tell the agent to search before it re-derives something:

```bash
agentmem install-rules --file ~/.claude/CLAUDE.md
```

Add the hooks that carry a short digest of the session across a compaction:

```bash
agentmem install-hooks --settings ~/.claude/settings.json
```

The hooks exist for Claude Code only. Sessions from pi and Hermes are indexed and searchable, but they get no digest.

### Check the hooks

The hooks need a live check on your machine:

1. Run `/compact` in a Claude Code session.
2. Look for a new file under `~/.local/share/agentmem/spool/`. The PreCompact hook writes this file.
3. After compaction, look for the digest in the session context. AgentMem then deletes the spool file.

If the SessionStart `compact` hook does not fire, the next prompt delivers the digest through UserPromptSubmit. Spool files expire after 24 hours. AgentMem never delivers a spool file to a different session.

### Roll back

```bash
agentmem install-hooks --settings ~/.claude/settings.json --remove
```

```bash
agentmem install-rules --file ~/.claude/CLAUDE.md --remove
```

```bash
rm -rf ~/.local/share/agentmem
```

The last command deletes the index and all other AgentMem state.

## Commands

| Command | Purpose |
|---|---|
| `ingest` | Index new session data |
| `ask` | Search the index |
| `show` | Print one chunk |
| `stats` | Print index counts as JSON with `--json` |
| `audit` | Check the index for secrets that redaction missed |
| `forget` | Remove a session or project (`--session ID`, `--project PATH`, `--dry-run`) |
| `rebuild` | Build the index again |
| `install-rules`, `install-hooks` | Connect Claude Code |

## Known limits

- The digest covers the whole transcript, not only the part after the last compaction.

## Where data lives

| Item | Default location | Override |
|---|---|---|
| Index and other state | `~/.local/share/agentmem` | `AGENTMEM_STATE_DIR` |
| Claude Code sessions | `~/.claude/projects` | `AGENTMEM_CLAUDE_DIR` |
| pi sessions | `~/.pi/agent/sessions` | `AGENTMEM_PI_DIR` |
| Hermes database | `~/.hermes/state.db` | `AGENTMEM_HERMES_DB` |

## Security and privacy

- **Secrets:** AgentMem redacts secrets before it writes the index, with rules based on gitleaks. Redaction can miss a secret, so run `agentmem audit` and review the result. The audit prints chunk ids and counts, never the text, and `leaks` must be 0.
- **Encryption:** the index is not encrypted. It holds your transcripts, including tool output. Do not share the file.
- **Untrusted content:** past sessions can contain hostile text, for example from a fetched web page. AgentMem labels its output as a historical record, not instructions. A label does not stop every attack, so an agent must treat results as data.

## Development

```bash
GOFLAGS=-mod=vendor go test ./... -count=1
```

```bash
go vet ./...
```

## License

[Apache-2.0](LICENSE)
