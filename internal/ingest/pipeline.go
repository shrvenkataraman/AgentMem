package ingest

import (
	"agentmem/internal/model"
	"agentmem/internal/redact"
)

// Process runs the fixed ingest pipeline on parsed messages:
// DropSelfPoison -> redact.Text on every Text -> NormalizeProject on every Project -> Window.
func Process(msgs []model.Message, known map[string]bool) []model.Chunk {
	if known == nil {
		known = map[string]bool{}
	}
	msgs = DropSelfPoison(msgs, known)
	proj := map[string]string{}
	for i := range msgs {
		msgs[i].Text = redact.Text(msgs[i].Text)
		p := msgs[i].Project
		if _, ok := proj[p]; !ok {
			proj[p] = NormalizeProject(p)
		}
		msgs[i].Project = proj[p]
	}
	return Window(msgs)
}
