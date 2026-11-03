// Package search answers queries over the chunks index.
package search

import "agentmem/internal/model"

// Hit is one candidate chunk. Coverage is set by Rerank, SeenIn by Collapse.
type Hit struct {
	ID int64
	model.Chunk
	Coverage float64
	SeenIn   int
}

// Cols is the SELECT column list scanned into a Hit, in this order (alias chunks as c).
const Cols = "c.id,c.session_id,c.source_store,c.agent,c.project,c.source_path,c.source_ref,c.seq,c.window_idx,c.ts,c.role,c.text,c.content_hash,c.is_error,c.tool_id,c.tool_name"
