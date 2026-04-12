# ch02 学习问答笔记

本文件仅记录学习过程中的疑问与回答，不记录代码修改内容。

## 本章学习导航（学习前先看）

### 1) 本章需要学习的内容
一句总结：把 LLM 从“会回答”升级为“会调用工具执行动作”的最小 Agent。

详细内容：
- 理解 Function Calling 的协议结构：`tools` 声明、`tool_calls` 返回、`tool` 消息回填。
- 掌握最小 Agent Loop：模型推理 -> 调工具 -> 观察结果 -> 再推理 -> 最终回答。
- 理解本地工具封装方式：`read`、`write`、`edit`、`bash` 的统一接口实现。

### 2) 本章操作内容（动手清单）
一句总结：跑通一个完整 tool loop，并观察每轮消息链如何变化。

详细内容：
- 执行基础任务：
  - `go run ./ch02/main -q "请读取 README.md 并总结项目目标"`
- 执行写文件任务：
  - `go run ./ch02/main -q "在 ch02 目录下创建一个 TODO.md，内容为 1. 研究 agent"`
- 观察点：
  - 模型是否先返回 `tool_calls` 再给最终答案。
  - 工具调用失败时，模型是否会基于错误继续尝试。
  - 多轮工具调用时，结果如何回填到 `messages`。

### 3) 本章学完需要掌握的知识点
一句总结：你应能解释“为什么 Agent 能动手做事”，以及“tool loop 如何闭环”。

详细内容：
- 能解释 `Tool.Info()` 的函数签名声明作用（名称、描述、参数 schema）。
- 能描述 `Run()` 中 tool loop 的退出条件（无 `tool_calls` 即结束）。
- 能说清 `assistant/tool/user` 三类消息在上下文中的协作方式。
- 能识别工具执行风险点（尤其 `bash`）并理解后续章节为何要引入安全策略。

### 4) 扩展阅读与参考资料总结
一句总结：扩展阅读重点是“协议正确性、SDK 实现细节、安全边界”。

详细内容：
- OpenAI Function Calling 官方文档
  - 原始链接：https://platform.openai.com/docs/guides/function-calling
  - 作用：确认工具调用协议与 JSON Schema 约束细节。
- OpenAI Go SDK 仓库
  - 原始链接：https://github.com/openai/openai-go
  - 作用：理解 Go SDK 如何组织工具类型、请求参数与响应解析。

## Q&A（学习中持续追加）

（待补充）
