// Package model holds the types shared by every stage of the ingest pipeline.
package model

// Message is one normalized unit of transcript content (one text block, one tool call, or one tool result).
type Message struct {
	Agent       string `json:"agent"`        // claude | pi | hermes
	SourceStore string `json:"source_store"` // claude | pi | hermes
	SessionID   string `json:"session_id"`
	SourcePath  string `json:"source_path"`
	SourceRef   string `json:"source_ref"` // "<line or row id>#<block index>"
	Seq         int    `json:"seq"`        // line or row id * 100 + block index
	TS          int64  `json:"ts"`         // epoch milliseconds
	Role        string `json:"role"`       // user | assistant | tool_call | tool_result
	Project     string `json:"project"`    // raw cwd as recorded by the harness
	Text        string `json:"text"`
	ToolID      string `json:"tool_id,omitempty"`
	ToolName    string `json:"tool_name,omitempty"`
	IsError     bool   `json:"is_error,omitempty"`
	Sidechain   bool   `json:"sidechain,omitempty"`
}

// Chunk is a Message after windowing: one stored row.
type Chunk struct {
	Message
	WindowIdx   int    `json:"window_idx"`
	ContentHash string `json:"content_hash"`
}
