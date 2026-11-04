package ingest

import (
	"agentmem/internal/model"
	"agentmem/internal/redact"
	"agentmem/internal/store"
)

// Process runs the fixed ingest pipeline on parsed messages:
// drop forgotten -> DropSelfPoison -> redact.Text on every Text -> NormalizeProject on every
// Project -> Window.
func Process(msgs []model.Message, known map[string]bool) []model.Chunk {
	if known == nil {
		known = map[string]bool{}
	}
	sess, projs, _ := store.LoadForgotten(store.StateDir())
	msgs = DropSelfPoison(msgs, known)
	proj := map[string]string{}
	out := msgs[:0]
	for i := range msgs {
		msgs[i].Text = redact.Text(msgs[i].Text)
		p := msgs[i].Project
		if _, ok := proj[p]; !ok {
			proj[p] = NormalizeProject(p)
		}
		msgs[i].Project = proj[p]
		if sess[msgs[i].SessionID] || projs[msgs[i].Project] {
			continue
		}
		out = append(out, msgs[i])
	}
	return Window(out)
}
