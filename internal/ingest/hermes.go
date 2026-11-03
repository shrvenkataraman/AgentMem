package ingest

import "agentmem/internal/model"

// ParseHermes reads user/assistant/tool rows of ~/.hermes/state.db style db at path (read-only)
// with messages.id > afterID, ordered by id. It returns the messages and the highest row id seen.
func ParseHermes(path string, afterID int64) (msgs []model.Message, maxID int64, err error) {
	panic("not implemented")
}
