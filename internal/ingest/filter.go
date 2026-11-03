package ingest

import "agentmem/internal/model"

// DropSelfPoison removes tool_call messages that invoke the agentmem binary together with the
// tool_result messages paired to them by (SessionID, ToolID). known holds the pairing keys
// (SessionID+"\x00"+ToolID) of already-dropped calls: it is read and updated in place, so a
// result that arrives in a later batch than its call is still dropped.
func DropSelfPoison(msgs []model.Message, known map[string]bool) []model.Message {
	panic("not implemented")
}
