package ch06

const (
	MessageTypeReasoning = "reasoning"
	MessageTypeContent   = "content"
	MessageTypeToolCall  = "tool_call"
	MessageTypeError     = "error"
	MessageTypePolicy    = "policy"
	MessageTypeMemory    = "memory"
)

// MessageVO 用于流式展示当前模型流式输出或者状态
type MessageVO struct {
	Type string `json:"type"`

	ReasoningContent *string     `json:"reasoning_content,omitempty"`
	Content          *string     `json:"content,omitempty"`
	ToolCall         *ToolCallVO `json:"tool,omitempty"`
	Policy           *PolicyVO   `json:"policy,omitempty"`
	Memory           *MemoryVO   `json:"memory,omitempty"`
}

// PolicyVO 策略执行状态
type PolicyVO struct {
	Name    string `json:"name"`    // 策略名称
	Running bool   `json:"running"` // 是否正在执行
	Error   error  `json:"error"`
}

// MemoryVO 记忆更新状态
type MemoryVO struct {
	Running bool   `json:"running"` // 是否正在执行
	Error   error  `json:"error"`
	Detail  string `json:"detail,omitempty"`
}

type ToolCallVO struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
