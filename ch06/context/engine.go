package context

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/openai/openai-go/v3"

	"babyagent/ch06/memory"
	"babyagent/shared"
)

type messageWrap struct {
	Message shared.OpenAIMessage
	Tokens  int
}

type Engine struct {
	systemPromptTemplate string
	messages             []messageWrap
	policies             []Policy
	onPolicyEvent        func(policyName string, running bool, err error)
	onMemoryEvent        func(event MemoryEvent)
	contextTokens        int
	contextWindow        int

	memory memory.Memory
}

type TokenBudget struct {
	ContextWindow int
}

type Usage struct {
	PromptTokens int
}

type MemoryEvent struct {
	Running bool
	Error   error
	Detail  string
	Report  memory.UpdateReport
}

type TurnDraft struct {
	NewMessages []shared.OpenAIMessage
}

func NewContextEngine(memory memory.Memory, policies []Policy) *Engine {
	return &Engine{
		policies:      policies,
		messages:      make([]messageWrap, 0),
		contextWindow: 200000,
		memory:        memory,
	}
}

func (c *Engine) Init(systemPrompt string, budget TokenBudget) {
	c.systemPromptTemplate = systemPrompt
	if budget.ContextWindow > 0 {
		c.contextWindow = budget.ContextWindow
	}
}

func (c *Engine) BuildRequestMessages() []shared.OpenAIMessage {
	result := make([]shared.OpenAIMessage, 0, len(c.messages)+1)
	if c.systemPromptTemplate != "" {
		result = append(result, openai.SystemMessage(c.BuildSystemPrompt()))
	}
	for i := range c.messages {
		result = append(result, c.messages[i].Message)
	}
	return result
}

func (c *Engine) StartTurn(userMsg shared.OpenAIMessage) TurnDraft {
	return TurnDraft{
		NewMessages: []shared.OpenAIMessage{userMsg},
	}
}

func (c *Engine) CommitTurn(ctx context.Context, draft TurnDraft, usage Usage) error {
	// 根据情况压缩上下文
	for i := range draft.NewMessages {
		msg := draft.NewMessages[i]
		c.messages = append(c.messages, messageWrap{Message: msg, Tokens: CountTokens(msg)})
	}
	c.recountTokens()
	if err := c.applyPolicies(ctx); err != nil {
		return err
	}
	// 更新记忆
	if c.onMemoryEvent != nil {
		c.onMemoryEvent(MemoryEvent{
			Running: true,
			Detail:  fmt.Sprintf("准备处理 %d 条新消息并更新 Global / Workspace Memory", len(draft.NewMessages)),
		})
	}
	report, err := c.memory.Update(ctx, draft.NewMessages)
	if c.onMemoryEvent != nil {
		c.onMemoryEvent(MemoryEvent{
			Running: false,
			Error:   err,
			Detail:  formatMemoryEventDetail(report, err),
			Report:  report,
		})
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Engine) AbortTurn(_ TurnDraft) {
	// no-op: draft is only in-memory and never committed unless CommitTurn is called.
}

func (c *Engine) GetContextUsage() float64 {
	if c.contextWindow <= 0 {
		return 0
	}
	return float64(c.contextTokens) / float64(c.contextWindow)
}

func (c *Engine) recountTokens() {
	totalTokens := 0
	for i := range c.messages {
		totalTokens += c.messages[i].Tokens
	}
	c.contextTokens = totalTokens
}

func (c *Engine) applyPolicies(ctx context.Context) error {
	for _, policy := range c.policies {
		if !policy.ShouldApply(ctx, c) {
			continue
		}
		if c.onPolicyEvent != nil {
			c.onPolicyEvent(policy.Name(), true, nil)
		}
		result, err := policy.Apply(ctx, c)
		if c.onPolicyEvent != nil {
			c.onPolicyEvent(policy.Name(), false, err)
		}
		if err != nil {
			return fmt.Errorf("apply policy %s: %w", policy.Name(), err)
		}
		c.messages = result.Messages
		c.recountTokens()
	}
	return nil
}

func (c *Engine) SetPolicyEventHook(hook func(policyName string, running bool, err error)) {
	c.onPolicyEvent = hook
}

func (c *Engine) SetMemoryEventHook(hook func(event MemoryEvent)) {
	c.onMemoryEvent = hook
}

func (c *Engine) BuildSystemPrompt() string {
	replaceMap := make(map[string]string)
	replaceMap["{runtime}"] = runtime.GOOS
	replaceMap["{workspace_path}"] = shared.GetWorkspaceDir()

	if c.memory != nil {
		replaceMap["{memory}"] = c.memory.String()
	} else {
		replaceMap["{memory}"] = ""
	}

	prompt := c.systemPromptTemplate
	for k, v := range replaceMap {
		prompt = strings.ReplaceAll(prompt, k, v)
	}
	return prompt
}

// Reset 清空所有消息（保留 system prompt）
func (c *Engine) Reset() {
	c.messages = make([]messageWrap, 0)
	c.contextTokens = 0
}

func (c *Engine) MemorySnapshot() memory.MemoryContent {
	if c.memory == nil {
		return memory.MemoryContent{}
	}
	return c.memory.Snapshot()
}

func formatMemoryEventDetail(report memory.UpdateReport, err error) string {
	if err != nil {
		return fmt.Sprintf(
			"处理 %d 条新消息失败：%v",
			report.ProcessedMessages,
			err,
		)
	}

	changedScopes := make([]string, 0, 2)
	if report.GlobalChanged {
		changedScopes = append(changedScopes, "Global")
	}
	if report.WorkspaceChanged {
		changedScopes = append(changedScopes, "Workspace")
	}
	if len(changedScopes) == 0 {
		changedScopes = append(changedScopes, "无实际变更")
	}

	return fmt.Sprintf(
		"处理 %d 条新消息；变更=%s；global %d->%d chars；workspace %d->%d chars",
		report.ProcessedMessages,
		strings.Join(changedScopes, ", "),
		report.GlobalCharsBefore,
		report.GlobalCharsAfter,
		report.WorkspaceCharsBefore,
		report.WorkspaceCharsAfter,
	)
}
