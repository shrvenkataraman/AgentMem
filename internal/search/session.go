package search

import (
	"database/sql"
	"os"

	"agentmem/internal/ingest"
)

// CurrentSession returns the session id `ask` should treat as the one it is running in, or "" when
// it cannot tell. Order: the --session flag value, AGENTMEM_SESSION, CLAUDE_CODE_SESSION_ID, else the
// newest chunk of the project the current working directory normalizes to.
func CurrentSession(db *sql.DB, flag string) string {
	if flag != "" {
		return flag
	}
	if s := os.Getenv("AGENTMEM_SESSION"); s != "" {
		return s
	}
	if s := os.Getenv("CLAUDE_CODE_SESSION_ID"); s != "" {
		return s
	}
	raw := os.Getenv("AGENTMEM_CWD")
	if raw == "" {
		raw, _ = os.Getwd()
	}
	project := ingest.NormalizeProject(raw)
	var id string
	if err := db.QueryRow("SELECT session_id FROM chunks WHERE project = ? ORDER BY ts DESC LIMIT 1",
		project).Scan(&id); err != nil {
		return ""
	}
	return id
}
