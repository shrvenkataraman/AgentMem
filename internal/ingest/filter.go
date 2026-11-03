package ingest

import "agentmem/internal/model"

// DropSelfPoison removes tool_call messages that invoke the agentmem binary together with the
// tool_result messages paired to them by ToolID.
func DropSelfPoison(msgs []model.Message) []model.Message {
	panic("not implemented")
}
