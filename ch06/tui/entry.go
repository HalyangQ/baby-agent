package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	reasonStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	toolStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	policyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("228")).Bold(true)
)

// LogEntry 日志条目结构体
type LogEntry struct {
	Title   string         // 标题（如 "你:", "推理:" 等）
	Content string         // 正文内容
	Style   lipgloss.Style // 渲染样式
}

// NewLabel 创建轮次标签
func NewLabel(content string) LogEntry {
	return LogEntry{Title: "", Content: content, Style: labelStyle}
}

// NewContent 创建用户输入
func NewContent(content string) LogEntry {
	return LogEntry{Title: "你", Content: content, Style: contentStyle}
}

// NewAnswer 创建 AI 回答
func NewAnswer(content string) LogEntry {
	return LogEntry{Title: "回答", Content: content, Style: contentStyle}
}

// NewReasoning 创建推理内容
func NewReasoning(content string) LogEntry {
	return LogEntry{Title: "推理", Content: content, Style: reasonStyle}
}

// NewTool 创建工具调用
func NewTool(content string) LogEntry {
	return LogEntry{Title: "工具调用", Content: content, Style: toolStyle}
}

// NewError 创建错误信息
func NewError(content string) LogEntry {
	return LogEntry{Title: "错误", Content: content, Style: errorStyle}
}

// NewPolicyRunning 创建策略运行中状态
func NewPolicyRunning(name string) LogEntry {
	return LogEntry{Title: "上下文策略", Content: fmt.Sprintf("%s (运行中...)", name), Style: policyStyle}
}

// UpdatePolicyCompleted 更新策略 log entry 为完成状态
func (e *LogEntry) UpdatePolicyCompleted(success bool) {
	status := "已完成"
	if !success {
		status = "已失败"
	}
	// 移除 " (运行中...)" 后缀并替换为完成状态
	e.Content = strings.Replace(e.Content, " (运行中...)", "", 1)
	e.Content = fmt.Sprintf("%s (%s)", e.Content, status)
}

// NewMemoryRunning 创建记忆更新运行中状态
func NewMemoryRunning(detail string) LogEntry {
	content := "(运行中...)"
	if detail != "" {
		content = detail + " (运行中...)"
	}
	return LogEntry{Title: "记忆更新", Content: content, Style: policyStyle}
}

// UpdateMemoryCompleted 更新记忆 log entry 为完成状态
func (e *LogEntry) UpdateMemoryCompleted(success bool, detail string) {
	status := "已完成"
	if !success {
		status = "已失败"
	}
	// 移除 "(运行中...)" 后缀并替换为完成状态
	e.Content = strings.Replace(e.Content, "(运行中...)", "", 1)
	e.Content = strings.TrimSpace(e.Content)
	if detail != "" {
		e.Content = detail
	}
	e.Content = fmt.Sprintf("%s (%s)", e.Content, status)
}

// NewMemorySnapshot 创建记忆快照
func NewMemorySnapshot(content string) LogEntry {
	return LogEntry{Title: "记忆快照", Content: content, Style: policyStyle}
}

// NewMemoryStatus 创建记忆状态
func NewMemoryStatus(content string) LogEntry {
	return LogEntry{Title: "记忆状态", Content: content, Style: policyStyle}
}

// NewBorder 创建分隔线
func NewBorder() LogEntry {
	return LogEntry{Title: "", Content: strings.Repeat("─", 48), Style: borderStyle}
}

// AppendContent 追加内容
func (e *LogEntry) AppendContent(chunk string) {
	e.Content += chunk
}

func (e *LogEntry) Render() string {
	if e.Title == "" {
		return e.Style.Render(e.Content)
	}
	return e.Style.Render(e.Title + ": " + e.Content)
}
