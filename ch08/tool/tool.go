package tool

import (
	"context"

	"github.com/openai/openai-go/v3"
)

type AgentTool = string

const (
	AgentToolRead        AgentTool = "read"
	AgentToolWrite       AgentTool = "write"
	AgentToolEdit        AgentTool = "edit"
	AgentToolBash        AgentTool = "bash"
	AgentToolLoadStorage AgentTool = "load_storage"
)

type Tool interface {
	ToolName() AgentTool
	Info() openai.ChatCompletionToolUnionParam
	Execute(ctx context.Context, argumentsInJSON string) (string, error)
}

type ToolEvent struct {
	ToolName string
	Stage    string
	Message  string
}

type EventHook func(ToolEvent)

type ObservableTool interface {
	SetEventHook(EventHook)
}

type DescribableTool interface {
	RuntimeDescription() string
}
