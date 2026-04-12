package ch02

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"babyagent/ch02/tool"
	"babyagent/shared"
)

type Agent struct {
	systemPrompt string
	model        string
	maxSteps     int
	client       openai.Client
	messages     []openai.ChatCompletionMessageParamUnion
	tools        map[tool.AgentTool]tool.Tool
}

const defaultMaxSteps = 12

type toolExecutionResult struct {
	ToolName string `json:"tool_name"`
	OK       bool   `json:"ok"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
}

func NewAgent(modelConf shared.ModelConfig, systemPrompt string, tools []tool.Tool) *Agent {
	a := Agent{
		systemPrompt: systemPrompt,
		model:        modelConf.Model,
		maxSteps:     defaultMaxSteps,
		client:       openai.NewClient(option.WithBaseURL(modelConf.BaseURL), option.WithAPIKey(modelConf.ApiKey)),
		tools:        make(map[tool.AgentTool]tool.Tool),
		messages:     make([]openai.ChatCompletionMessageParamUnion, 0),
	}
	for _, t := range tools {
		a.tools[t.ToolName()] = t
	}
	a.messages = append(a.messages, openai.SystemMessage(systemPrompt))
	return &a
}

func (a *Agent) SetMaxSteps(steps int) error {
	if steps <= 0 {
		return fmt.Errorf("max steps must be positive, got %d", steps)
	}
	a.maxSteps = steps
	return nil
}

func (a *Agent) formatToolResult(toolName string, output string, execErr error) string {
	payload := toolExecutionResult{
		ToolName: toolName,
		OK:       execErr == nil,
		Output:   output,
	}
	if execErr != nil {
		payload.Error = execErr.Error()
	}
	b, err := json.Marshal(payload)
	if err != nil {
		// Fallback to plain text when serialization fails.
		if execErr != nil {
			return fmt.Sprintf("tool=%s ok=false output=%q error=%q", toolName, output, execErr.Error())
		}
		return fmt.Sprintf("tool=%s ok=true output=%q", toolName, output)
	}
	return string(b)
}

func (a *Agent) execute(ctx context.Context, toolName string, argumentsInJSON string) (string, error) {
	t, ok := a.tools[tool.AgentTool(toolName)]
	if !ok {
		return "", errors.New("tool not found")
	}
	return t.Execute(ctx, argumentsInJSON)
}

// Run 提供对于单次用户请求 query 的 tool loop，返回本轮结果的输出。Run 会保持当前对话历史，不同主题的对话轮次应该初始化多个 Agent 实例运行。
func (a *Agent) Run(ctx context.Context, query string) (string, error) {
	a.messages = append(a.messages, openai.UserMessage(query))

	for i := 0; i < a.maxSteps; i++ {
		params := openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: a.messages,
			Tools:    make([]openai.ChatCompletionToolUnionParam, 0),
		}

		for _, t := range a.tools {
			params.Tools = append(params.Tools, t.Info())
		}

		log.Printf("round %d, calling llm model %s, messages: %s", i, a.model, shared.JsonString(a.messages))
		resp, err := a.client.Chat.Completions.New(ctx, params)
		if err != nil {
			return "", fmt.Errorf("round %d llm request failed: %w", i, err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("round %d no choices returned", i)
		}
		message := resp.Choices[0].Message
		log.Printf("round %d, llm response: %v", i, shared.JsonString(message))
		// 拼接 assistant message 到整体消息链中
		a.messages = append(a.messages, message.ToParam())

		// tool loop 结束，可以返回结果
		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		for _, toolCall := range message.ToolCalls {
			toolResult, err := a.execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			log.Printf("round %d, tool call %s, arguments %s, error: %v", i, toolCall.Function.Name, toolCall.Function.Arguments, err)
			// 返回 tool message 到整体消息链中
			a.messages = append(a.messages, openai.ToolMessage(a.formatToolResult(toolCall.Function.Name, toolResult, err), toolCall.ID))
		}
	}
	return "", fmt.Errorf("max tool loop steps reached (%d)", a.maxSteps)
}
