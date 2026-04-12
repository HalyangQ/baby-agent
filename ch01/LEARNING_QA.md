# ch01 学习问答笔记

本文件仅记录学习过程中的疑问与回答，不记录代码修改内容。

## 本章学习导航（学习前先看）

### 1) 本章需要学习的内容
一句总结：理解一次 LLM 调用从“协议层”到“工程层”的完整路径。

详细内容：
- 认识 `chat/completions` 的最小请求与响应结构。
- 理解非流式与流式（SSE）在数据返回方式上的差异。
- 对比 `Raw HTTP` 与 `OpenAI SDK` 两种实现方式的工程取舍。
- 建立上下文与 token usage 的基础认知，为后续 agent 能力（tool、memory、context policy）做准备。

### 2) 本章操作内容（动手清单）
一句总结：用同一问题跑通 4 种调用组合，并观察日志差异。

详细内容：
- 准备环境变量：配置 `OPENAI_BASE_URL`、`OPENAI_API_KEY`、`OPENAI_MODEL`。
- 运行四种模式：
  - SDK + 非流式
  - SDK + 流式
  - Raw HTTP + 非流式
  - Raw HTTP + 流式
- 观察关键现象：
  - 非流式一次性返回完整内容。
  - 流式按 chunk 增量返回，并以 `[DONE]` 结束。
  - usage 在非流式与流式中的出现时机差异。
  - SDK 与 Raw 实现复杂度差异。

### 3) 本章学完需要掌握的知识点
一句总结：你应能解释“协议结构、流式机制、SDK 价值、choice 与 usage 的工程含义”。

详细内容：
- 能说清楚 `messages`、`model`、`stream` 的作用。
- 能解释 SSE 基本格式与流式结束标识。
- 能说明为什么 SDK 更适合工程落地，以及何时需要 Raw HTTP。
- 能解释 `choices` 的设计意义与多候选选择策略。
- 能说明流式 usage 的获取条件：`stream_options.include_usage=true` + 服务端支持。

### 4) 扩展阅读与参考资料总结
一句总结：这组资料的目标是同时补齐“协议原理”和“工程落地”。

详细内容：
- OpenAI Chat Completions API 文档
  - 原始链接：https://developers.openai.com/api/reference/resources/chat/subresources/completions/methods/create
  - 作用：确认请求参数与响应结构的官方标准。
  - 学习价值：后续实现兼容 OpenAI 协议的 Agent 时可直接对照。
- MDN SSE 文档
  - 原始链接：https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events
  - 作用：理解流式返回的底层协议格式（如 `data:` 与结束语义）。
  - 学习价值：帮助你看懂 Raw HTTP 流式解析逻辑。
- `openai-go` SDK 仓库
  - 原始链接：https://github.com/openai/openai-go
  - 作用：理解 SDK 对请求、流式迭代、错误处理的封装方式。
  - 学习价值：从“会用 SDK”升级到“理解 SDK”。
- 其他兼容厂商 API 文档（如 DeepSeek）
  - 原始链接：https://api-docs.deepseek.com
  - 作用：验证 OpenAI 协议兼容的迁移路径。
  - 学习价值：建立多厂商切换意识，降低平台绑定风险。

#### OpenAI API Reference - Create chat completion：额外收获
一句总结：除了基础调用，这篇文档已经覆盖了结构化输出、工具编排、多模态与生产参数等进阶能力。

详细内容：
- `developer` 消息角色
  - 在较新的模型中，`developer` 指令用于承载高优先级行为约束（可理解为新一代系统指令语义）。
- 多模态能力
  - `messages` 可承载文本、图片、音频等输入内容部分。
  - 可请求 `modalities` 输出文本与音频，并配置 `audio` 参数（如格式与声音）。
- 结构化输出
  - `response_format` 支持 `json_schema`，可启用严格 schema 约束。
  - 相比旧的 `json_object`，`json_schema` 是更推荐的工程方案（模型支持时）。
- Tool 调用控制
  - `tool_choice` 支持 `none/auto/required`，也可强制指定工具。
  - 还支持限制可用工具集合（allowed tools）与 custom tool（含 grammar 格式）。
- `n` 与成本关系
  - 一次生成多个 choice 会增加总 token 消耗，计费按全部候选累计。
  - 默认保持 `n=1` 往往更经济，只有在 rerank 场景再提高 `n`。
- usage 细分诊断
  - 除总量外，还可看到 `prompt_tokens_details` / `completion_tokens_details` 等更细指标。
  - 这些细分有助于判断缓存命中、推理 token 占比、音频 token 等成本结构。
- 生产参数 `service_tier`
  - 可指定请求处理层级（如 `auto/default/flex/...`）。
  - 响应可能返回实际采用的 tier，未必完全等于请求值。
- 安全与缓存标识演进
  - `user` 参数在向 `safety_identifier` 与 `prompt_cache_key` 的组合演进。
  - 这反映生产系统对“滥用检测 + 缓存优化”的双重要求。
- `store` 参数
  - 可控制请求输出是否用于 distillation/evals 相关产品能力。
- 流式安全开关
  - `stream_options.include_obfuscation` 用于流量混淆以缓解侧信道风险，可按网络信任边界取舍。

#### OpenAI API Reference - Create chat completion：参数速查与主要用途
一句总结：参数很多，但可以按“生成控制、工具编排、结构化输出、生产治理”四层来理解。

参数归属（四类）：
- 生成控制
  - `temperature`、`top_p`、`max_completion_tokens`、`n`、`frequency_penalty`、`presence_penalty`、`seed`、`stop`、`verbosity`、`stream`、`stream_options`
- 工具编排
  - `tools`、`tool_choice`、`parallel_tool_calls`、`web_search_options`（以及旧字段 `functions`、`function_call`）
- 结构化输出
  - `response_format`（`text` / `json_schema` / `json_object`）、`prediction`
- 生产治理
  - `service_tier`、`prompt_cache_key`、`prompt_cache_retention`、`safety_identifier`、`metadata`、`store`（以及旧字段 `user`）

高频参数及用途（优先掌握）：
- `messages`：输入上下文与消息历史（几乎所有行为都由它驱动）。
- `model`：决定能力、成本、延迟和可用特性。
- `temperature` / `top_p`：控制回答随机性，通常优先只调一个。
- `max_completion_tokens`：限制输出上限，防止成本和长度失控。
- `stream`：开启流式输出，提升首字响应体验。
- `stream_options.include_usage`：流式场景返回最终 token usage。
- `tools`：声明模型可调用的工具能力。
- `tool_choice`：控制工具调用策略（禁用、自动、强制）。
- `response_format`：约束输出结构；工程场景优先 `json_schema`。
- `n`：一次返回多个候选，适合 rerank，但会提高成本。
- `service_tier`：控制服务层级（延迟/价格策略）。
- `prompt_cache_key`：提升提示缓存命中率，降低重复请求成本。
- `safety_identifier`：稳定匿名用户标识，用于安全风控。

## Q1. 相比 Raw HTTP，使用 SDK 的核心价值是什么？
一句总结：SDK 把底层协议复杂度封装掉了，同时保留了多厂商兼容能力。

详细回答：
- 抽象更高，代码更短：直接调用 `client.Chat.Completions.New(...)` / `NewStreaming(...)`，不用手写 URL、Header、请求体和 SSE 读循环。
- 类型更安全：请求参数和响应字段由 Go 类型约束，减少手写 JSON 结构出错概率。
- 流式接口更易用：用 `stream.Next()` / `stream.Current()` 迭代，不用自己处理 `data: ...` 和 `[DONE]`。
- 可移植性仍保留：通过 `option.WithBaseURL(modelConf.BaseURL)` 仍可切到兼容 OpenAI 协议的其他服务商。

## Q2. 非流式和流式在 usage 获取上有什么差异？
一句总结：流式也能拿到 usage，但需要显式开启并依赖服务端支持。

详细回答：
- 流式场景需要在请求里设置 `stream_options.include_usage=true`。
- usage 通常只会在最后一个 chunk 给出，中间 chunk 一般没有 usage。
- 若是第三方兼容 OpenAI 的服务，可能不支持该字段，就会拿不到 usage。
- 因此可理解为：非流式默认更稳定拿 usage；流式能拿，但有条件。

## Q3. 为什么流式响应中依然保留 `choices` 结构？
一句总结：`choices` 是协议层的“候选回答列表”，流式和非流式共用这套结构。

详细回答：
- `choices` 表示一次请求可能返回多个候选结果（通常由 `n` 控制，默认 `1`）。
- 非流式中，每个 choice 是完整 `message`。
- 流式中，每个 choice 是增量 `delta`，按 chunk 逐步拼起来。
- 保留 `choices` 有三个意义：协议一致性、支持多候选并行流、以及为 tool call/finish reason 等字段扩展留空间。

## Q4. 多个 choice 的设计目的是什么？工程上如何选出最优答案？
一句总结：是，多个 choice 常用于候选竞争与重排，核心是“先过滤，再打分”。

详细回答：
- 常见用途：
  - 采样多样性：同题多解，得到不同风格/思路。
  - 质量提升：先生成多份，再评审重排，常比单次更稳。
  - 下游策略：按场景路由不同风格答案（如简洁版/详细版）。
- 选择策略（实用顺序）：
  - 先做硬过滤：剔除格式不合法、缺字段、超长度、违规、未答题候选。
  - 再做规则打分：按正确性、完整性、可执行性、简洁度、语气匹配加权求分。
  - 复杂任务可加“评审模型”做 rerank；成本敏感时先规则筛再评审 top K。
  - 高风险场景加一致性检查，排除冲突和逻辑漏洞。

## Q5. `messages` 是否需要传入全部历史聊天记录，作为上下文？
一句总结：通常需要传入“当前回答所需的历史上下文”，但工程上不会无限累加全量历史。

详细回答：
- Chat Completions 接口本身是无状态的，服务端不会自动记住上一次请求内容。
- 因此你需要把当前轮必要的信息放进 `messages`，包括系统/开发者指令、相关历史对话、工具结果。
- 生产实践里一般采用“上下文窗口管理”：只保留近邻对话、对远历史做摘要、或分层记忆存储，避免 token 爆炸。

## Q6. `temperature` 如何理解？“输出随机性”在神经网络里如何生效？
一句总结：`temperature` 不改模型权重，只在推理时改变下一 token 的采样分布形状。

详细回答：
- 模型每一步先输出 logits（每个候选 token 的分数）。
- 解码时先做缩放：`logits / T`，再 softmax 得到概率分布。
- `T < 1`：分布更尖锐，头部 token 更容易被选中，输出更稳定。
- `T > 1`：分布更平坦，尾部 token 更容易被采样，输出更发散。
- 所以这是“推理时的采样策略参数”，不是训练阶段参数。

## Q7. 推理链路中“解码阶段”到底是什么？
一句总结：解码阶段就是把模型分数转成可输出 token，并循环生成整段回答的过程。

详细回答：
- 一次生成可分为两段：
  - 前向计算：输入 tokens 经过 Transformer，得到下一 token 的 logits。
  - 解码阶段：对 logits 做温度/过滤/采样，选出下一个 token。
- 单步解码流程：
  - 计算 logits
  - 应用 `temperature`（以及可选 `top_p` 等）
  - softmax 得到概率
  - 采样或贪心选 token
  - 回填上下文，进入下一步
- 非流式与流式的区别只在“返回时机”：
  - 非流式：全部解码完成后一次性返回。
  - 流式：每解出一部分就增量返回 chunk。

## Q8. logits 和 token 的关系是什么？能举个例子吗？
一句总结：token 是候选输出单位，logits 是模型给每个候选 token 的原始分数。

详细回答：
- 在每一步生成时，模型会对词表中的每个 token 给出一个分数，这组分数就是 logits。
- 这些 logits 经过 softmax 转成概率分布后，系统再按策略（采样或贪心）选出下一个 token。
- 因此关系是：`token` 是候选项，`logits` 是候选项对应的打分。

示例（简化）：
- 候选 token：`["北京", "上海", "广州", "深圳"]`
- logits：`[3.2, 2.1, 0.5, -0.3]`
- softmax 后概率（约）：`[0.67, 0.24, 0.07, 0.02]`
- 解码结果：最可能选中 `"北京"` 作为下一 token。

## Q9. `max_completion_tokens` 在推理链路中如何实现“限制”？
一句总结：它通过解码循环中的“已生成 token 计数器”做硬截止，而不是事后裁剪文本。

详细回答：
- 推理阶段是逐 token 解码，服务端会持续累计已生成 token 数。
- 每生成一个 token，计数器递增；达到 `max_completion_tokens` 后立即停止生成。
- 这种停止通常对应 `finish_reason = "length"`。
- 该参数限制的是输出 completion 长度，不是输入 prompt 长度。

## Q10. 流式返回是“每生成一个 token 就立即返回”吗？
一句总结：本质是增量返回，但实际常按小批次 chunk 推送，不一定严格 1 token 1 包。

详细回答：
- 解码循环产生新 token 后，服务端会组装为 SSE 增量事件并尽快 flush。
- 为了网络与系统开销优化，服务端/网关常做短暂缓冲，合并几个 token 一起发。
- 所以客户端看到的是“近实时连续输出”，粒度可能是单 token，也可能是多 token。
- 结束时会发送完成信号（如 `[DONE]`），可选再附 usage 收尾 chunk。

## Q11. 模型训练时是否会学习“结束”，避免一直文本接龙？
一句总结：会学习自然收尾模式，但工程上仍需硬性停止条件兜底。

详细回答：
- 训练数据让模型学到“何时该结束”的统计模式，推理时会出现自然停止倾向。
- 线上系统不会只依赖该能力，还会配置 `max_completion_tokens`、`stop` 序列、超时与安全策略。
- 因此稳定结束来自“双保险”：模型内生收尾能力 + 工程外部控制策略。

## Q12. `prompt_cache_key` 怎么用？多轮对话还生效吗？平台如何隔离不同接入方？
一句总结：`prompt_cache_key` 是缓存分组线索，适合同场景复用；多轮可命中稳定前缀；缓存隔离按租户边界完成。

详细回答：
- 它的作用
  - `prompt_cache_key` 不是缓存内容本身，而是帮助平台把“同场景、同模板版本”的请求归到同一缓存分组。
  - 目标是提高 prompt caching 命中率，从而优化延迟和成本。
- 如何设计与复用
  - 同产品、同功能、同提示词版本可复用同一个 key。
  - 推荐命名：`product:feature:prompt_version:segment`。
  - 示例：`support:refund:v2:zh-CN`、`agent:qa:v3:enterprise-cn`。
  - 模板改动后应升级版本，避免错误命中。
- 平台如何区分不同接入方
  - 一般先按账号/项目（租户）隔离，再在租户内结合 key 进行缓存匹配。
  - 不同接入方即使 key 同名，也不会互相命中缓存。
- 多轮对话是否生效
  - 会生效，但核心命中对象是“稳定前缀”（系统提示、固定模板、公共知识段）。
  - 首轮通常命中最明显；轮次越深上下文越分叉，命中收益通常下降。
  - 实操建议：把稳定内容前置、对长历史做摘要、保持 key 与模板版本同步。

## Q13. SSE 标准里的 `data`、`event`、`id` 分别做什么？`id` 是谁生成的？
一句总结：`data` 承载消息内容，`event` 指定事件类型，`id` 用于断线续传；`id` 由服务端生成。

详细回答：
- `data:`
  - 事件的消息体内容。
  - 可出现多行 `data`，客户端会按规范拼接成一次事件 payload。
- `event:`
  - 事件类型名。
  - 客户端可按类型监听（如 `addEventListener("ping", ...)`）；缺省时按默认 `message` 处理。
- `id:`
  - 事件标识，用于断线重连时的续传定位。
  - 客户端会记录最后收到的 `id`，重连时通过 `Last-Event-ID` 发送给服务端。
- `id` 由谁生成
  - 通常由 SSE 服务端生成，格式可自定义（递增序号、时间戳、UUID 等）。
  - 客户端不负责生成，只负责保存并在重连时回传。

## Q14. 如何简述 SSE 协议？它在 LLM 流式输出里怎么使用？
一句总结：SSE 是基于 HTTP 的服务端单向事件流，LLM 用它把解码增量实时推送给客户端。

详细回答：
- SSE 协议特性
  - 客户端发起一次请求后，连接保持打开，服务端持续推送事件。
  - 事件字段常见有 `data`、`event`、`id`、`retry`，空行表示单个事件结束。
- 在 LLM 流式输出中的工作方式
  - 客户端请求时开启流式（如 `stream: true`）。
  - 服务端边解码边发送 `data` 增量事件，客户端实时拼接显示（打字机效果）。
  - 生成完成后发送结束信号（常见 `data: [DONE]`）。
  - 若开启 usage 回传（如 `include_usage`），通常会在结束前追加 usage 事件。

## Q15. openai-go SDK 是如何封装“重试”和“错误处理”的？（含代码示例）
一句总结：SDK 内置自动重试与错误类型化返回；若要“回调式”处理，可通过 middleware 自定义拦截。

详细回答：
- 自动重试（内置）
  - 默认最多重试 2 次，可用 `option.WithMaxRetries(n)` 覆盖。
  - 默认重试范围：连接错误、`408`、`409`、`429`、`>=500`。
  - 退避逻辑：优先遵循 `Retry-After-Ms/Retry-After`，否则指数退避 + 抖动。
- 错误处理（类型化返回）
  - API 非 2xx 时，SDK 返回 `*openai.Error`（包含状态码、请求/响应对象、错误体）。
  - 推荐用 `errors.As(err, &apierr)` 分流处理。
  - 非 API 错误（如网络层）会原样返回（如 `*url.Error` / `*net.OpError`）。
- 回调式错误处理（工程做法）
  - SDK 本身不是“注册错误回调”模式。
  - 可通过 `option.WithMiddleware(...)` 在请求前后拦截，统一记录错误与耗时。

示例代码（重试 + 错误处理 + middleware）：
```go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func errHook(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
	start := time.Now()
	resp, err := next(req)
	if err != nil {
		log.Printf("[transport_err] method=%s url=%s err=%v", req.Method, req.URL.String(), err)
		return resp, err
	}
	if resp.StatusCode >= 400 {
		log.Printf("[http_err] status=%d cost=%s", resp.StatusCode, time.Since(start))
	}
	return resp, nil
}

func main() {
	client := openai.NewClient(
		option.WithAPIKey("sk-xxx"),
		option.WithMaxRetries(3),
		option.WithRequestTimeout(20*time.Second),
		option.WithMiddleware(errHook),
	)

	_, err := client.Chat.Completions.New(context.Background(), openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5_2,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage("hello"),
		},
	},
		option.WithMaxRetries(5),
	)
	if err != nil {
		var apierr *openai.Error
		if errors.As(err, &apierr) {
			log.Printf("api status=%d", apierr.StatusCode)
			log.Printf("req dump:\n%s", apierr.DumpRequest(true))
			log.Printf("resp dump:\n%s", apierr.DumpResponse(true))
			return
		}
		log.Printf("non-api err: %v", err)
	}
}
```

## Q16. 流式 SSE 中途失败怎么处理？是否意味着整次请求失败？
一句总结：默认按“本次流请求失败或不完整”处理，能否断点恢复取决于服务端是否支持续传协议。

详细回答：
- SSE 连接中断后，客户端通常会收到流错误并结束读取流程。
- 在多数 LLM 流式接口中，这次请求应视为失败或不完整结束。
- 工程上最常见策略是“整次重试”：
  - 重新发起请求；
  - 客户端做必要的去重/拼接保护（避免重复展示）。
- 只有服务端支持基于 `id` / `Last-Event-ID` 的可重放续传时，才可能从断点恢复。
- 因此默认心智模型是：中断即失败；续传属于可选增强能力而非默认保证。

## Q17. SSE 底层是 TCP 吗？“跑在 HTTP 上”具体是什么意思？
一句总结：SSE 的传输链路通常是 `SSE -> HTTP -> TCP`，其本质是一个长生命周期的 HTTP 响应流。

详细回答：
- 底层协议关系
  - SSE 不是新的传输层协议，它是 HTTP 响应体的事件流格式。
  - HTTP（1.1/2）通常运行在 TCP 上；若是 HTTPS，则在 HTTP 与 TCP 间还有 TLS。
- “跑在 HTTP 上”的含义
  - 客户端发起普通 HTTP 请求（常见 `Accept: text/event-stream`）。
  - 服务端返回 `Content-Type: text/event-stream`，并保持连接不断开。
  - 服务端持续写入事件块，客户端持续读取并处理。
- 是否是应用层会话
  - SSE 本身不强制要求服务端维护业务会话状态。
  - 工程上可以无状态推送，也可以结合业务上下文做有状态推送，这取决于服务实现而非 SSE 协议本身。
