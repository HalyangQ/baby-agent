package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

type BashTool struct {
	eventHook EventHook
}

func NewBashTool() *BashTool {
	return &BashTool{}
}

type BashToolParam struct {
	Command string `json:"command"`
}

func (t *BashTool) ToolName() AgentTool {
	return AgentToolBash
}

func (t *BashTool) RuntimeDescription() string {
	return fmt.Sprintf("bash tool: regular shell runtime=%s", runtime.GOOS)
}

func (t *BashTool) SetEventHook(hook EventHook) {
	t.eventHook = hook
}

func (t *BashTool) emit(stage, message string) {
	if t.eventHook == nil {
		return
	}
	t.eventHook(ToolEvent{
		ToolName: string(t.ToolName()),
		Stage:    stage,
		Message:  message,
	})
}

func (t *BashTool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        string(AgentToolBash),
		Description: openai.String("execute bash command"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "the bash command to execute",
				},
			},
			"required": []string{"command"},
		},
	})
}

func (t *BashTool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	p := BashToolParam{}
	err := json.Unmarshal([]byte(argumentsInJSON), &p)
	if err != nil {
		return "", err
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows: use cmd.exe to interpret the command line
		cmd = exec.CommandContext(ctx, "cmd", "/C", p.Command)
	} else {
		// Linux/macOS: use POSIX sh (more universal than assuming bash exists)
		cmd = exec.CommandContext(ctx, "sh", "-c", p.Command)
	}

	start := time.Now()
	t.emit("execute_start", fmt.Sprintf("running on host shell: %s", p.Command))
	output, err := cmd.CombinedOutput()
	t.emit("execute_done", fmt.Sprintf("host shell finished in %s, output_bytes=%d, err=%v", time.Since(start).Round(time.Millisecond), len(output), err))
	if err != nil {
		return "", err
	}
	return string(output), nil
}
