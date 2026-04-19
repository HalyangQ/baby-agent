package ch05

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/openai/openai-go/v3"

	"babyagent/ch05/tool"
	"babyagent/shared"

	ctxengine "babyagent/ch05/context"
)

type Agent struct {
	model         string
	client        openai.Client
	contextEngine *ctxengine.Engine
	nativeTools   map[tool.AgentTool]tool.Tool
	mcpClients    map[string]*McpClient
}

func NewAgent(modelConf shared.ModelConfig, systemPrompt string, tools []tool.Tool, mcpClients []*McpClient, contextEngine *ctxengine.Engine) *Agent {
	a := Agent{
		model:         modelConf.Model,
		client:        shared.NewLLMClient(modelConf),
		contextEngine: contextEngine,
		nativeTools:   make(map[tool.AgentTool]tool.Tool),
		mcpClients:    make(map[string]*McpClient),
	}

	a.contextEngine.Init(systemPrompt, ctxengine.TokenBudget{ContextWindow: modelConf.ContextWindow})

	for _, t := range tools {
		a.nativeTools[t.ToolName()] = t
	}
	for _, mcpClient := range mcpClients {
		a.mcpClients[mcpClient.Name()] = mcpClient
	}

	return &a
}

func (a *Agent) execute(ctx context.Context, toolName string, argumentsInJSON string) (string, error) {
	// 判断 native tool
	t, ok := a.nativeTools[toolName]
	if ok {
		return t.Execute(ctx, argumentsInJSON)
	}
	// 判断 MCP Tool
	for _, mcpClient := range a.mcpClients {
		for _, t := range mcpClient.GetTools() {
			if t.ToolName() != toolName {
				continue
			}
			return t.Execute(ctx, argumentsInJSON)
		}
	}
	return "", errors.New("tool not found")
}

func (a *Agent) buildTools() []openai.ChatCompletionToolUnionParam {
	tools := make([]openai.ChatCompletionToolUnionParam, 0)
	// 集成 mcp tools
	for _, t := range a.nativeTools {
		tools = append(tools, t.Info())
	}
	// 集成 mcp tools
	for _, mcpClient := range a.mcpClients {
		for _, t := range mcpClient.GetTools() {
			tools = append(tools, t.Info())
		}
	}
	return tools
}

func (a *Agent) ResetSession() {
	a.contextEngine.Reset()
}

func (a *Agent) DebugContext() string {
	stats := a.contextEngine.DebugStats()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("messages=%d, tokens=%d, context_window=%d, usage=%.2f%%",
		stats.MessageCount,
		stats.ContextTokens,
		stats.ContextWindow,
		stats.ContextUsage*100,
	))
	if len(stats.Policies) > 0 {
		b.WriteString("\n")
		b.WriteString("policies:")
		for _, policy := range stats.Policies {
			b.WriteString("\n")
			b.WriteString(fmt.Sprintf("- %s: threshold=%.2f", policy.Name, policy.Threshold))
			if policy.Description != "" {
				b.WriteString(", ")
				b.WriteString(policy.Description)
			}
		}
	}
	return b.String()
}

// RunStreaming 和 Run 基本逻辑一致，但是使用流式请求，并且通过 channel 实现流式输出
func (a *Agent) RunStreaming(ctx context.Context, query string, viewCh chan MessageVO) error {
	a.contextEngine.SetPolicyEventHook(func(event ctxengine.PolicyEvent) {
		viewCh <- MessageVO{
			Type: MessageTypePolicy,
			Policy: &PolicyVO{
				Name:           event.Name,
				Running:        event.Running,
				Error:          event.Error,
				Summary:        event.Summary,
				BeforeMessages: event.BeforeMessages,
				AfterMessages:  event.AfterMessages,
				BeforeTokens:   event.BeforeTokens,
				AfterTokens:    event.AfterTokens,
			},
		}
	})
	defer a.contextEngine.SetPolicyEventHook(nil)

	draft := a.contextEngine.StartTurn(openai.UserMessage(query))
	defer a.contextEngine.AbortTurn(draft)

	// 为本轮次创建新的消息链。草稿消息在 commit 前不会污染上下文。
	messages := a.contextEngine.BuildRequestMessages()
	messages = append(messages, draft.NewMessages...)
	var usage openai.CompletionUsage
	for {
		params := openai.ChatCompletionNewParams{
			Model:    a.model,
			Messages: messages,
			Tools:    a.buildTools(),
		}

		log.Printf("calling llm model %s...", a.model)
		stream := a.client.Chat.Completions.NewStreaming(ctx, params)
		acc := openai.ChatCompletionAccumulator{}
		for stream.Next() {
			chunk := stream.Current()
			acc.AddChunk(chunk)

			if len(chunk.Choices) > 0 {
				deltaRaw := chunk.Choices[0].Delta
				// 不同厂商会把推理内容放在 reasoning_content、reasoning 或 thinking 字段里。
				delta, err := parseDeltaWithReasoning(deltaRaw.RawJSON())
				if err != nil {
					log.Printf("parse delta failed, raw=%s, err=%v", deltaRaw.RawJSON(), err)
					continue
				}
				if reasoningContent := delta.ReasoningText(); reasoningContent != "" {
					viewCh <- MessageVO{
						Type:             MessageTypeReasoning,
						ReasoningContent: &reasoningContent,
					}
				}
				if delta.Content != "" {
					content := delta.Content
					viewCh <- MessageVO{
						Type:    MessageTypeContent,
						Content: &content,
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			viewCh <- MessageVO{
				Type:    MessageTypeError,
				Content: shared.Ptr(err.Error()),
			}
			return err
		}
		if len(acc.Choices) == 0 {
			log.Printf("no choices returned, resp: %v", acc)
			return nil
		}
		usage = acc.Usage
		message := acc.Choices[0].Message
		// 拼接 assistant message 到整体消息链中
		assistantMsg := message.ToParam()
		messages = append(messages, assistantMsg)
		draft.NewMessages = append(draft.NewMessages, assistantMsg)

		// tool loop 结束，可以返回结果
		if len(message.ToolCalls) == 0 {
			break
		}

		for _, toolCall := range message.ToolCalls {

			viewCh <- MessageVO{
				Type: MessageTypeToolCall,
				ToolCall: &ToolCallVO{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			}

			toolResult, err := a.execute(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				toolResult = err.Error()

				viewCh <- MessageVO{
					Type:    MessageTypeError,
					Content: &toolResult,
				}

			}
			log.Printf("tool call %s, arguments %s, error: %v", toolCall.Function.Name, toolCall.Function.Arguments, err)
			// 返回 tool message 到整体消息链中
			toolMsg := openai.ToolMessage(toolResult, toolCall.ID)
			messages = append(messages, toolMsg)
			draft.NewMessages = append(draft.NewMessages, toolMsg)
		}

	}

	err := a.contextEngine.CommitTurn(ctx, draft, ctxengine.Usage{PromptTokens: int(usage.TotalTokens)})
	return err
}

type deltaWithReasoning struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
	Reasoning        string `json:"reasoning"`
	Thinking         string `json:"thinking"`
}

func parseDeltaWithReasoning(rawJSON string) (deltaWithReasoning, error) {
	delta := deltaWithReasoning{}
	err := json.Unmarshal([]byte(rawJSON), &delta)
	return delta, err
}

func (d deltaWithReasoning) ReasoningText() string {
	switch {
	case d.ReasoningContent != "":
		return d.ReasoningContent
	case d.Reasoning != "":
		return d.Reasoning
	default:
		return d.Thinking
	}
}
