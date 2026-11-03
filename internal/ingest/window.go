package ingest

import "agentmem/internal/model"

// Window splits msgs into chunks (windows of about 2560 bytes at paragraph boundaries) and caps
// tool_result text. ContentHash is hex sha256 of the chunk text.
func Window(msgs []model.Message) []model.Chunk {
	panic("not implemented")
}
