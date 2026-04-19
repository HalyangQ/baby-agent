# ch04 学习问答笔记

本文件仅记录学习过程中的疑问与回答，不记录代码修改内容。

## 本章学习导航（学习前先看）

### 1) 本章需要学习的内容
一句总结：这一章表面上是在“接 MCP 工具”，本质上是在让 Agent 从“自己带几个工具”升级到“接入标准化工具生态”。

详细内容：
- 理解 MCP（Model Context Protocol）到底解决什么问题，以及它和普通 tool calling 的边界在哪里。
- 理解 Agent 为什么需要同时管理本地工具和 MCP 工具，以及这意味着什么样的系统分层。
- 理解 `ch04` 新增的能力不是“多一个文件系统工具”，而是“工具接入方式被协议化、生态化”。
- 理解 MCP client、MCP server、tool schema、tool execution 之间的关系，以及它们在项目中的映射。

### 2) 本章操作内容（动手清单）
一句总结：先跑起来，再观察 Agent 是如何把本地工具和 MCP 工具统一暴露给模型的。

详细内容：
- 运行：
  - `go run ./ch04/main`
- 建议尝试的输入：
  - “请读取 README.md 并总结项目目标”
  - “使用 MCP 工具列出当前目录下的文件”
  - “当前你有哪些可用工具，请按类别解释”
- 观察点：
  - MCP 工具是否成功加载。
  - 模型看到的工具名为什么会带 `babyagent_mcp__...` 前缀。
  - 本地工具和 MCP 工具是否都能走同一套 tool loop。

### 3) 本章学完需要掌握的知识点
一句总结：你学完这章后，应该能把“MCP 是什么”讲成一个系统能力问题，而不只是“又多接了几个工具”。

详细内容：
- 能解释 MCP 为什么要存在，它解决的是“工具接入标准化”问题，而不是单纯的“远程调用”。
- 能说清 MCP 和 OpenAI tool calling 的关系：前者更像工具生态协议，后者更像模型调用接口。
- 能理解本地工具和 MCP 工具在 Agent 内部为什么可以被统一成同一个 `tool.Tool` 接口。
- 能解释为什么 MCP 工具要做命名空间包装，以及这背后的工具冲突问题。
- 能区分“模型调用工具”和“Agent 接入工具生态”是两层不同的系统设计。

### 4) 扩展阅读与参考资料总结
一句总结：这章的扩展阅读重点不是某个具体 MCP server，而是 MCP 作为标准协议的角色、能力边界和 SDK 实现方式。

详细内容：
- MCP 官方文档（概览）
  - 原始链接：https://modelcontextprotocol.io/
  - 作用：建立对 MCP 角色、能力和整体目标的第一印象。
- MCP 规范（Spec）
  - 原始链接：https://github.com/modelcontextprotocol/spec
  - 作用：理解 MCP 在协议层到底规定了什么，比如初始化、工具发现、工具调用和能力协商。
- MCP Go SDK
  - 原始链接：https://github.com/modelcontextprotocol/go-sdk
  - 作用：理解本章 `mcp.go` 里那些连接、会话、列工具、调用工具的代码到底来自什么抽象。

### 5) 扩展阅读需要重点学习什么
一句总结：扩展阅读要重点看的是“MCP 为什么存在”和“它和你已经学过的 tool calling 到底是什么关系”。

详细内容：
- MCP 的协议定位
  - 学什么：MCP 想统一的到底是什么，是 server 侧能力接入、工具发现，还是模型调用协议。
  - 目标：避免把 MCP 简化成“一个远程工具 API”。
- MCP 与 tool calling 的关系
  - 学什么：MCP 和 OpenAI tools/function calling 分别处在哪一层。
  - 目标：建立“模型调用接口”和“工具生态协议”是两层系统设计的认知。
- SDK 如何落地协议
  - 学什么：client session、transport、list tools、call tool 的抽象。
  - 目标：把 README 里的协议流程映射到 `ch04/mcp.go` 的代码实现。

### 6) 本章应掌握的核心知识点
一句总结：`ch04` 学完后，你真正应该掌握的，不是某个 `filesystem` MCP server 怎么连，而是 Agent 怎么把外部工具生态以协议化方式接进来，并理解这件事离工业级治理还差什么。

详细内容：
- **这一章新增的是“工具生态接入能力”，不是简单多了一个工具。**  
  `ch02/ch03` 里你主要在学模型怎么调用工具；到了 `ch04`，重点变成 Agent 怎么把外部工具生态标准化接进来。也就是说，这一章的主线不是 function calling 本身，而是工具来源和工具接入方式被协议化了。
- **MCP 和 tool calling 处在不同层。**  
  tool calling 解决的是“模型怎么调用工具”，MCP 解决的是“Agent 怎么接工具生态”。前者是模型接口层，后者是系统协议层。模型最终看到的仍然是一组 tools，而不是 MCP 本身，这个分层意识非常关键。
- **这个项目里实现的是 MCP 的最小可用 tools 接入子集。**  
  当前 `ch04` 主要做了连接 MCP server、`ListTools`、`CallTool`、把 MCP tools 适配到统一的 `tool.Tool` 接口。也就是说，它已经跑通了 MCP 的 tools 主路径，但 resources、prompts、roots、sampling 这些更完整的 MCP 能力基本都还没碰。
- **MCP tool 进入 Agent 后，要经历两层适配。**  
  第一层是命名空间适配，也就是把原始 server tool 名包装成 `babyagent_mcp__{server}__{tool}` 这种全局唯一名字；第二层是 schema 适配，也就是把 MCP tool 定义包装成模型可见的 tool schema。这两层的目的是让 MCP tool 和本地 tool 能进入同一个 tool loop。
- **包装名不是执行名，Agent 内部同时维护“模型侧视图”和“server 侧原始身份”。**  
  模型看到的是包装名，但真正执行 `CallTool` 时，还是回到 MCP server 原始 tool 名。这说明 `McpTool` 本质上是一个桥接对象：对模型像一个普通 tool，对 server 又保留原始协议语义。
- **当前 schema 传递基本是直通透传，而不是强治理。**  
  `ch04` 里 MCP server 返回的 `InputSchema` 基本直接被当作模型侧 tool parameters 使用，中间没有显式的 schema 治理层。这种方式简单、教学友好，但也意味着兼容性、校验、默认值、降级这些工业问题还没被真正解决。
- **MCP 引入的是外部协议接入复杂度，而不只是“工具更多”。**  
  一旦接了 MCP，系统就开始面对连接管理、协议兼容、命名冲突、安全边界、调试链路变长、可靠性依赖外部 server 这些问题。也就是说，Agent 从“本地工具执行器”开始变成“外部能力接入器和工具生态管理器”。
- **这个项目当前更像“轻接入层”，不是“强治理层”。**  
  它已经能把 MCP tool 跑通，但还没有权限控制、schema 治理、可靠性策略、结果标准化、审计监控、策略路由这些工业级能力。所以你要能判断：它已经实现了什么，也要知道它还没有实现什么。
- **MCP 不是只有 tools，真实 Agent 系统还需要内容、模板和边界能力。**  
  tools 只解决动作执行；resources 解决上下文内容供给；prompts 解决可复用提示模板；roots 解决操作边界声明；sampling 解决 server 反向借用模型能力。理解这一点，你才不会把 MCP 简化成“远程 function calling”。
- **工程选型时，要区分动作、内容、模板和边界。**  
  做事优先用 tools；提供上下文内容优先用 resources；复用提示结构优先用 prompts；声明允许操作的工作区边界用 roots。不要把一切都硬做成 tool，否则系统一复杂，语义、缓存、治理和上下文管理都会混乱。

## Q&A

## [重点] Q1. MCP 到底解决什么问题？它和前面学过的 tool calling 有什么根本区别？
一句总结：tool calling 解决的是“模型怎么调用工具”，MCP 解决的是“Agent 怎么以标准协议接入外部工具生态”；前者偏模型接口，后者偏系统架构。

详细回答：
- **tool calling 解决的是模型调用接口问题：**  
  在 `ch02/ch03` 里，模型已经会调用工具了，也就是说模型这一侧已经解决了工具 schema 怎么暴露给模型、模型怎么返回 `tool_call`、Agent 怎么执行工具，以及工具结果怎么回填给模型。所以 tool calling 解决的是“模型如何发起工具调用”这个问题。
- **MCP 解决的是工具从哪来、怎么接进来：**  
  前几章没有解决的一个大问题是，这些工具到底从哪来。如果没有 MCP，每接一个新的外部工具体系，Agent 都得自己做一遍连接服务、适配工具列表、转换参数 schema、调用执行接口、处理返回结果的工作。这会导致模型方和工具方之间出现组合爆炸。
- **MCP 想做的是标准化接入：**  
  MCP 的目标是让工具提供方和 Agent/应用接入方之间有一套统一协议。这样工具 server 只要按 MCP 暴露能力，Agent 只要实现 MCP client，就能发现工具、读取 schema、调用工具，而不用为每个外部工具体系单独定制适配。
- **两者处在不同层：**  
  tool calling 更偏模型层或 LLM API 层，模型看到的是一组工具 schema，并决定调用哪个工具；MCP 更偏系统层或工具生态层，Agent 负责去发现、连接、管理外部工具，再把这些工具转成模型能理解的 schema。
- **为什么模型最终看到的仍然只是工具：**  
  模型并不知道某个工具来自本地代码，还是来自一个 MCP server。MCP 的复杂性被藏在 Agent 和 MCP client 这一层，模型只会看到工具名、工具描述和参数 schema。这也是 `ch04` 最值得建立的工程意识：协议接入层和模型调用层是两层不同的系统设计。

## Q2. MCP client、MCP server、tool schema、tool execution 之间是什么关系？
一句总结：server 是能力提供方，client 是协议接入方，schema 是能力描述层，execution 是真实调用层。

详细回答：
- **MCP server：**  
  真正提供能力的一方。它负责暴露有哪些 tools、每个 tool 的描述和参数 schema，并在收到调用时真正执行工具逻辑。
- **MCP client：**  
  接入 server 的一方。在这个项目里就是 Agent 里的那层 MCP 适配代码。它负责连接 server、拉取 tools、调用 tools，并把协议世界翻译给 Agent。
- **tool schema：**  
  是“工具说明书”。它描述工具叫什么、做什么、参数怎么传。MCP server 会先把自己的 tool schema 返回给 MCP client；然后 MCP client 再把它包装成模型可见的 tool schema。
- **tool execution：**  
  是真正执行工具的过程。模型先根据 tool schema 决定要调用哪个工具、传什么参数；Agent 再通过 MCP client 把这次调用转成 `CallTool` 发给 MCP server；最后由 server 真正执行并返回结果。
- **把它们压成一条链路：**  
  可以理解成 `MCP server -> 返回 tool schema -> MCP client/Agent 适配 -> 模型看到 tool schema -> 模型发起 tool call -> MCP client 转成 CallTool -> MCP server 执行 -> 返回结果`。

## Q3. 这个项目里的 MCP server 代码在哪？`connect()` 连接的对端到底是谁？
一句总结：`ch04/mcp.go` 里只有 MCP client；真正的 MCP server 不在仓库里，而是由 `mcp-server.json` 配置的 `npx @modelcontextprotocol/server-filesystem` 在运行时启动出来，并通过 `stdio` 和 Agent 通信。

详细回答：
- **仓库里没有 MCP server 的实现代码：**  
  `ch04/mcp.go` 这里只有 MCP client 逻辑，没有本地实现一个 MCP server。真正的 server 来自根目录的 `mcp-server.json` 配置。
- **默认连接的是外部文件系统 server：**  
  `mcp-server.json` 里配置的是 `@modelcontextprotocol/server-filesystem`。也就是说，默认对接的不是仓库里的 Go 服务，而是一个通过 `npx` 启动的外部 MCP server 进程。
- **`connect()` 的真实含义是“启动并连接一个本地子进程”：**  
  当配置里有 `command` 和 `args` 时，`connect()` 会执行 `exec.Command(...)`，实际拉起一个本地进程，然后通过 `mcp.CommandTransport` 和这个子进程的标准输入/标准输出建立 MCP 会话。所以它不是去连一个你已经写好的 Go server，而是运行时启动外部服务端。
- **这里的 client/server 分工：**  
  `babyagent` 是 MCP client，`npx -y @modelcontextprotocol/server-filesystem ${workspaceFolder}` 启动出来的进程是 MCP server，二者通过 `stdio` 传 JSON-RPC / MCP 协议消息。
- **`RefreshTools()` 做的是工具发现：**  
  建好 session 之后，`RefreshTools()` 会调用 `ListTools`，也就是向那个外部 MCP server 询问“你有哪些工具、它们的 description 和 input schema 是什么”，然后再把这些工具包装成项目自己的 `tool.Tool`。

## Q4. `command` 和 `stdio` 分别是什么？除了这种方式，MCP 还会怎么通信？
一句总结：`command` 是“怎么启动本地 MCP server 进程”的配置，`stdio` 是“通过标准输入输出和这个进程通信”的传输方式；除了 `stdio`，MCP 还常见 `HTTP/SSE` 这种远程传输方式。

详细回答：
- **`command` 是启动 server 的命令配置：**  
  在这个项目里，`command` 不是协议概念，而是“怎么把对端 MCP server 拉起来”的本地命令。比如 `npx` 配合 `@modelcontextprotocol/server-filesystem`，意思就是用这个命令去启动一个外部 MCP server 进程。
- **`stdio` 是标准输入输出通信：**  
  `stdio` 指的是 standard input / standard output，也就是标准输入和标准输出。当 MCP 走 `stdio` 时，client 会启动一个本地子进程，然后往它的 stdin 写请求，再从它的 stdout 读响应。所以它适合本机进程间通信、本地工具 server，以及不需要额外开网络端口的场景。
- **这和 HTTP 不同：**  
  如果走 `stdio`，client 和 server 更像两个本地进程直接通过管道通信；如果走 `HTTP/SSE`，client 则是通过网络请求去连接一个远程 MCP server。这时不需要自己拉起本地进程，而是连一个已经在运行的远端服务。
- **这个项目已经体现了两种配置思路：**  
  `shared/mcp.go` 里，带 `command` 的配置会被当作 `stdio` 传输；带 `url` 的配置会被当作远程 HTTP 类传输。也就是说，MCP 协议和传输方式是分开的：协议层可以保持一致，但底层既可以走本地进程通信，也可以走网络通信。
- **更适合你的理解方式：**  
  `command` 回答的是“这个 server 怎么启动”，`stdio` 回答的是“启动后怎么和它说话”。除了本地 `stdio`，MCP 还常见 `HTTP/SSE` 这类远程通信方式。这样理解，就比较容易把“协议层”和“传输层”分开了。

## Q5. 如果是一个远端的 MCP server，是不是就不用 `command` 了？
一句总结：远端 MCP server 通常不需要 `command`，因为它不由你本地启动；`command` 只用于本地拉起 server 进程的场景。

详细回答：
- **本地 MCP server 需要回答两件事：**  
  第一是这个 server 怎么启动，第二是启动后怎么和它通信。所以本地场景下常见的是 `command` / `args` 配合 `stdio`。
- **远端 MCP server 不需要你负责启动：**  
  如果 server 已经运行在别的机器或别的进程里，Agent 的职责就不是把它拉起来，而是直接去连接它。这时更关键的是它的地址是什么，以及用什么远程 transport 去连它。
- **远端场景更像“连地址”，不是“起进程”：**  
  所以远端配置通常会更接近 `url` 加上 `HTTP/SSE` 这类远程传输方式，而不是 `command`。因为 `command` 本质上只回答“怎么在本地启动一个 server 进程”。
- **更适合你的理解方式：**  
  `command` 只在“我要本地拉起一个 server”时才有意义；如果 server 已经在远端运行好了，Agent 只要知道去哪里连它，就不需要 `command`。

## Q6. `npx` 是什么？它在这里起什么作用？
一句总结：`npx` 是 npm 生态里用来直接运行包内命令的工具，这里它的作用是临时启动 `@modelcontextprotocol/server-filesystem` 这个 MCP server。

详细回答：
- **`npx` 不是协议概念，而是 Node/npm 工具：**  
  它可以理解成“直接运行一个 npm 包里的可执行程序”。在这个项目里，`npx` 的作用不是长期安装依赖，而是临时把某个 npm 包下载并运行起来。
- **这里的具体含义：**  
  `npx -y @modelcontextprotocol/server-filesystem ${workspaceFolder}` 的意思是找到 `@modelcontextprotocol/server-filesystem` 这个包，如果本地没有就临时下载，然后直接运行它暴露出来的命令。
- **为什么这里用它：**  
  因为 `@modelcontextprotocol/server-filesystem` 是一个 npm 包，而项目只是想把它当作 MCP server 启动起来使用，不想把整个 Node 工程接进来，所以用 `npx` 是最方便的方式。

## Q7. 这里起的是一个代理 server 吗？它是不是替我去 npm 包里拉取 tool 再加载进来？
一句总结：不是 `npx` 代理去拉工具，而是 `npx` 把一个现成的 MCP server 跑起来；这个 server 自己就内置了文件系统 tools，Agent 只是去发现并调用它们。

详细回答：
- **`npx` 只是启动器：**  
  它负责把 `@modelcontextprotocol/server-filesystem` 这个 npm 包下载并运行起来，但它本身不是 MCP 工具代理层。
- **真正的 MCP server 是那个 npm 包：**  
  运行起来之后，`@modelcontextprotocol/server-filesystem` 本身就是一个文件系统 MCP server。它自己实现了一组文件系统相关工具，而不是运行后再去别的地方动态拉一批工具回来。
- **Agent 做的是工具发现，不是工具下发：**  
  Agent 连上这个 server 后，会调用 `ListTools` 去问“你有哪些工具”。server 再把自己本来就实现好的工具列表返回给 Agent。之后 Agent 再把这些工具包装成模型可见的 tool schema。
- **更准确的流程：**  
  先由 `npx` 启动 npm 包实现的 MCP server；再由 Agent 通过 MCP session 连接这个 server；然后由 server 返回自己已有的工具定义；最后 Agent 再发现、包装并调用这些工具。

## Q8. `@modelcontextprotocol/server-filesystem` 是谁实现的？
一句总结：`@modelcontextprotocol/server-filesystem` 不是这个项目实现的，而是 MCP 官方生态里的一个现成文件系统 server；这个项目只是把它当作外部 MCP server 连进来。

详细回答：
- **它不属于这个仓库的代码：**  
  `ch04/mcp.go` 里只有 MCP client，没有本地实现一个文件系统 MCP server。`@modelcontextprotocol/server-filesystem` 来自项目外部，不是这个仓库作者自己写的。
- **它属于 MCP 官方 servers 生态：**  
  这个包属于 `modelcontextprotocol/servers` 这条官方生态线，经常被列为 official MCP server 的一部分。从公开资料看，它是 MCP 官方提供的文件系统 server 实现。
- **它在这个项目里的角色：**  
  你的 `baby-agent` 只是通过 `npx` 把它启动起来，再把它当作一个外部 MCP server 去连接、发现工具、调用工具。也就是说，项目自己实现的是 MCP client，不是这个 filesystem server。
- **更适合的理解方式：**  
  你可以把角色分成三层：第一层是 MCP 协议和官方生态；第二层是 `@modelcontextprotocol/server-filesystem` 这种官方 server；第三层是你的 `baby-agent`，它只是 client，负责接入这个 server 提供出来的工具。

## [重点] Q9. 为什么这个外部 server 提供出来的 tool，不能直接把原始名字暴露给模型，而要包装成 `babyagent_mcp__filesystem__xxx`？
一句总结：`babyagent_mcp__filesystem__xxx` 这种包装名，本质上是在给跨 server 的工具做全局唯一命名和来源标识，避免冲突，并让模型侧、执行侧、调试侧都能稳定识别工具。

详细回答：
- **避免名字冲突：**  
  不同 MCP server 可能都提供叫 `read_file`、`search`、`list` 这种工具。如果直接把原始名字暴露给模型，工具集合一合并就会撞名。模型看到两个同名工具时，系统就很难稳定区分，调用结果也容易混乱。
- **把工具来源编码进名字：**  
  `babyagent_mcp__filesystem__read_file` 这种名字里，至少包含这是 `babyagent` 接进来的工具、它来自 `mcp`、它属于 `filesystem` 这个 server、原始工具名是 `read_file` 这几层信息。这样模型侧和系统侧都能明确知道这不是本地原生 tool，也不是别的 MCP server 的 tool。
- **统一本地 tool 和 MCP tool 的管理方式：**  
  在项目内部，最终都要转成统一的 `tool.Tool` 抽象给模型看。既然都进了同一个工具池，就需要一套统一、全局唯一的命名规则。这个包装名其实就是“模型可见名字”，而 MCP server 自己的原始名字仍然可以在内部保留，用于真正调用对端。
- **降低后续扩展时的心智负担：**  
  现在只有一个 `filesystem` server，看起来直接叫 `read_file` 也能跑。但如果后面再接 `github`、`database`、`slack` 这类 server，很快就会出现很多语义相近甚至同名的工具。早点做 namespacing，比后面补救稳定得多。
- **方便调试和审计：**  
  当你看到一条 tool call 是 `babyagent_mcp__filesystem__read_file`，你立刻就知道它来自哪个接入层、哪个 server。如果只看到 `read_file`，你还得额外查映射关系。
- **更准确的理解方式：**  
  不是外部 server 的原始名字不能用，而是原始名字更适合 server 内部语义，包装后的名字才适合进入 Agent 的全局工具空间，暴露给模型使用。

## [重点] Q10. 为什么本地启动的 MCP 工具，最后还要被重新包装成模型可见的 tool schema？
一句总结：本地启动的 MCP 工具之所以还要重新包装成模型可见的 tool schema，是因为模型不理解 MCP 协议；Agent 必须把 MCP 工具翻译成统一的模型工具视图，才能让本地 tool 和 MCP tool 走同一套调用链。

详细回答：
- **MCP tool 和模型 tool 不是同一个抽象：**  
  MCP server 返回的是它自己的工具定义，里面带工具名、描述、输入 schema，以及后续如何通过 MCP `CallTool` 去执行。但模型并不理解 MCP 协议，也不会自己发 MCP 请求。模型只理解“这里有一组可调用工具，它们叫什么、参数是什么”。
- **Agent 要充当协议翻译层：**  
  所以 Agent 需要把 MCP server 返回的工具列表，重新包装成模型能理解的 tool schema，再交给模型。模型只负责决定要不要调用这个工具、传什么参数；真正执行时，还是 Agent 拿着这次调用去转成 MCP `CallTool` 请求发给对端 server。
- **这样才能把不同来源的工具统一进同一个 tool loop：**  
  现在项目里既有本地 tool，也有 MCP tool。如果不做重新包装，模型面对的就会是两套完全不同的工具体系。重新包装之后，不管工具来自本地实现还是 MCP server，模型看到的都是统一的一组 tool schema，tool loop 也就可以复用。
- **包装时还能补上 Agent 自己需要的治理信息：**  
  这一步不只是格式转换，还顺便做了几件工程上必须的事，比如做全局唯一命名、记录来源 server、保留原始 tool 名和模型可见名之间的映射，以及把执行入口统一收敛到 Agent 内部。
- **更准确的理解方式：**  
  不是 MCP server 直接把 tool 给模型，而是 Agent 先拿到 MCP tool，再做一层“模型视图包装”，最后才展示给模型。所以这一步本质上是一个 adapter / facade。

## Q11. MCP tool 被包装成统一名字后，真正执行时又是怎么映射回原始 server tool 的？
一句总结：包装名只用于模型侧的全局唯一标识，真正发给 MCP server 执行的仍然是 `McpTool` 内部保存的原始 tool 名。

详细回答：
- **工具发现时，先保存原始 MCP tool 名：**  
  在 `RefreshTools()` 里，每个 MCP tool 都会被包装成一个 `McpTool`，其中 `toolName: mcpTool.Name`。这里保存的就是 server 原始名字，例如 `read_file`。
- **暴露给模型时，再生成包装名：**  
  `ToolName()` 返回的是 `babyagent_mcp__{server}__{tool}`，也就是模型看到的是包装后的全局名字，不是原始 `read_file`。
- **真正执行时，回到原始名字：**  
  `Execute()` 里调用的是 `t.client.callTool(ctx, t.toolName, argumentsInJSON)`。这里传进去的是 `t.toolName`，也就是最早保存下来的原始 MCP tool 名，而不是 `ToolName()` 返回的包装名。
- **最终发给 MCP server 的仍然是原始名字：**  
  在 `callTool()` 里，真正发起 MCP 调用时是 `session.CallTool(...)`，其中 `Name: toolName`。所以发给 MCP server 的仍然是原始 server tool 名。
- **更准确的理解方式：**  
  这相当于 `McpTool` 内部同时维护了两套名字：一套给模型看，解决全局命名空间问题；一套给 server 看，保证 MCP 调用仍然走原始协议语义。

## Q12. MCP tool 被包装成模型可见 schema 之后，执行时又是怎么映射回原始 MCP tool 并真正发起 `CallTool` 的？
一句总结：MCP tool 被包装成模型可见 schema 后，并不是靠后续字符串反查来映射回原始 tool，而是 `McpTool` 在创建时就同时保存了模型侧包装视图和 server 侧原始信息；执行时 `Execute()` 直接拿内部保存的原始 `toolName` 去发起 `session.CallTool(...)`。

详细回答：
- **第一步，`ListTools` 时拿到原始 MCP tool：**  
  `RefreshTools()` 会先调用 `e.session.ListTools(...)`，拿到 `mcpToolResult.Tools`。每个 `mcp.Tool` 里面本来就有原始 `Name`、`Description` 和 `InputSchema`。
- **第二步，把原始信息包进 `McpTool`：**  
  在循环里创建 `McpTool` 时，会保存 `toolName: mcpTool.Name`、`mcpTool: mcpTool`、`client: e`。这里最关键的是，`McpTool` 不只是生成一个展示名字，而是把原始 tool 名、原始 schema 和 client/session 都保存下来了，所以它本身就是一张映射表。
- **第三步，给模型暴露的是包装后的 schema：**  
  `Info()` 里构造 OpenAI tool schema 时，`Name` 用的是 `t.ToolName()`，也就是 `babyagent_mcp__...`；`Description` 用的是原始 `mcpTool.Description`；`Parameters` 用的是原始 `mcpTool.InputSchema`。所以模型看到的是“包装后的名字 + 原始 schema 内容”。
- **第四步，模型实际选中的是这个 `tool.Tool` 对象：**  
  Agent 在给模型暴露工具时，不只是给了一个名字字符串，而是内部维护了一组 `tool.Tool` 对象。模型最后触发某个 tool call 时，Agent 实际上会找到对应的那个 `McpTool` 实例去执行。
- **第五步，执行时直接回到原始名字：**  
  `McpTool.Execute()` 里没有用包装名，而是调用 `t.client.callTool(ctx, t.toolName, argumentsInJSON)`。这里的 `t.toolName` 是最开始保存的原始 MCP tool 名。
- **第六步，真正发起 MCP `CallTool`：**  
  `callTool()` 里调用 `e.session.CallTool(ctx, &mcp.CallToolParams{ Name: toolName, Arguments: json.RawMessage(argumentsInJSON) })`。这里的 `Name` 就是原始 server tool 名，所以对 MCP server 来说，它根本不知道模型看到过 `babyagent_mcp__filesystem__read_file` 这种包装名。

## [重点] Q13. MCP tool 的输入 schema 是怎么从 server 传到模型侧的，过程中有没有损失或约束变化？
一句总结：在 `ch04` 里，MCP tool 的输入 schema 基本是从 server 原样透传到模型侧的，中间没有明显的格式裁剪；但它依赖 schema 本身能兼容 OpenAI-style tool parameters，而且当前实现没有额外的 schema 校验、降级和治理。

这章里基本是直通传递，中间几乎没有做 schema 级转换，但仍然存在一些潜在约束和风险。

先看代码路径。

- MCP server 先返回原始 input schema：
  在 ch04/mcp.go 的 `RefreshTools()` 里，`ListTools()` 拿到的是 `mcpToolResult.Tools`。每个 `mcp.Tool` 里都有自己的 InputSchema。
- 包装成模型可见 tool 时，直接把原始 schema 塞给 OpenAI tool schema：
  在 `McpTool.Info()` 里：

  ```
  return openai.ChatCompletionFunctionTool(shared2.FunctionDefinitionParam{
        Description: openai.String(t.mcpTool.Description),
        Name:        t.ToolName(),
        Parameters:  t.mcpTool.InputSchema.(map[string]any),
    })
  ```

  这里的 Parameters 就是直接用 `t.mcpTool.InputSchema`，没有做字段重写、裁剪或归一化。 所以按当前实现，schema 传递路径基本是 `MCP server -> McpTool -> OpenAI tool parameters`:

1. MCP server 返回 InputSchema
2. McpTool 保存它
3. Info() 直接把它作为模型侧 Parameters
4. 模型按这个 schema 生成 arguments
5. 执行时原样把 argumentsInJSON 传回 CallTool

从这个项目代码看，没有显式的 schema 损失转换层。
但这不代表“完全没有约束变化风险”，因为还是有几类隐含前提。

- 第一类，类型断言约束：
  代码直接写了：`t.mcpTool.InputSchema.(map[string]any)`

  这意味着它假设 MCP tool 的 InputSchema 一定已经是一个 Go 的 map[string]any，也就是 JSON object 形式。
  如果某个 server 返回的 schema 不是这种结构，当前代码会直接 panic，而不是优雅降级。
- 第二类，模型工具 schema 和 MCP schema 的兼容前提：
  这里隐含假设是：MCP server 给出的 schema，和 OpenAI function calling 所期望的 parameters 结构足够接近。
  对常见 JSON Schema 子集来说通常没问题，但如果某个 MCP server 用了模型侧不稳定支持的复杂 schema 特性，模型虽然看到了 schema，实际遵守程度可能会下降。
- 第三类，当前实现没有做 schema 治理：
  它没有做：
  - 字段裁剪
  - 默认值补充
  - 枚举压缩
  - 不兼容关键字降级
  - strict 模式控制

  也就是说，它走的是“server 给什么，就尽量原样给模型看”的路线。这很简单，也很教学化，但工业场景未必够稳。
- 第四类，执行侧其实也没有二次参数校验：
  模型生成的 argumentsInJSON 在 Execute() 里直接传到：

  session.CallTool(... Arguments: json.RawMessage(argumentsInJSON))

  这说明 Agent 这一层没有再按 schema 自己做一遍强校验。
  所以真正的参数合法性，更多还是依赖：
  - 模型是否按 schema 生成
  - MCP server 自己是否严格校验

所以更准确地说：

- 从代码实现上看： schema 基本是原样从 MCP server 透传到模型侧的。
- 从系统语义上看： 中间没有主动“损失转换”，但存在兼容性前提和缺少治理层的问题。
## [重点] Q14. MCP 在这个项目里到底引入了哪些新风险和复杂度，而不只是“多了一种工具来源”？
一句总结：MCP 在这个项目里引入的真正变化，是把 Agent 从“本地工具执行器”推进成“外部协议接入器和工具生态管理器”；新增的复杂度主要来自连接管理、协议兼容、安全边界、调试链路、治理能力和可靠性。

详细回答：
- **连接和生命周期复杂度上升：**  
  本地 tool 是进程内直接调用，MCP tool 则要先连 server。这样就多了连接建立、重连、session 存活、server 是否在线、本地子进程是否正常退出这些问题。工具不再只是一个函数，而是一个外部依赖。
- **协议兼容风险出现了：**  
  现在中间多了一层 MCP 协议和 SDK。你要假设 server 的 `ListTools` 返回结构符合预期、`InputSchema` 能兼容模型侧 tool schema、`CallTool` 的返回内容能被当前 Agent 正确消费。这意味着问题不再只是“模型调没调对”，还可能是“协议对没对齐”。
- **工具来源变多后，命名和治理变复杂：**  
  一旦接多个 MCP server，就会出现工具重名、同类工具行为不一致、不同 server 质量参差不齐、某些工具 description 写得很弱的问题。所以系统必须开始做命名空间、来源标识、冲突治理，而不是简单把工具都塞给模型。
- **安全边界扩大了：**  
  本地 tool 至少是你仓库里自己写的，MCP tool 则可能来自第三方 server。风险包括 server 本身能力过强、tool description 误导模型、返回结果不可信，以及本地 `command` 启动的外部进程本身就是额外攻击面。所以 MCP 引入的是“外部工具信任问题”，不只是“多一个工具”。
- **调试链路变长了：**  
  以前出问题，你主要查 prompt、tool schema 和 tool execution。现在还得查 MCP client 是否连上、server 是否正确暴露工具、schema 是不是透传成功、tool call 有没有正确映射回原始 server tool、server 返回有没有按预期解析。整条链路更长，定位成本更高。
- **参数和返回值治理变弱了：**  
  当前 `ch04` 基本是把 MCP schema 直通给模型，再把模型 arguments 直通回 server。这样实现很轻，但意味着 Agent 层几乎不做二次校验，schema 不兼容时容易直接出错，返回结果格式不统一时上层也难以稳定消费。也就是说，Agent 现在更像“协议转发器”，还不是“强治理层”。
- **可靠性问题开始变真实：**  
  本地 tool 一般是同步、确定、可控的。MCP tool 可能受外部 server 状态影响，比如 server 启动失败、远端 HTTP/SSE 不稳定、tool list 拉取失败、tool call 超时。这会让 Agent 从“单机逻辑问题”变成“分布式依赖问题”。
- **模型可用性不再只取决于模型本身：**  
  以前回答不好，主要怀疑模型或 prompt。现在模型表现还受到工具生态质量影响：schema 写得差、server 返回不稳定、description 不清晰，都可能让模型表现下降。也就是说，MCP 把“模型质量问题”扩展成了“模型 + 工具生态质量问题”。

## [重点] Q15. 这个项目里 MCP 目前更像“轻接入层”还是“强治理层”，以及离工业化还差什么？
一句总结：这个项目里的 MCP 目前更像轻接入层，擅长把外部工具生态接进 Agent，但还没有建立工业级所需的权限控制、schema 治理、可靠性、结果标准化、审计监控和策略路由能力。

详细回答：
- **为什么说它更像轻接入层：**  
  `ch04` 现在做的核心事情，是把 MCP server 接进来、拉取工具列表、包装成模型可见 schema、再把调用转回 `CallTool`。这说明它已经有了“协议接入”能力，但主要职责还是建连接、拉工具、做名字包装、转发调用和回传结果。这套能力更接近 adapter / bridge，而不是 policy / governance。
- **为什么它还不是强治理层：**  
  它基本没有做那些工业系统很在意的控制动作，比如强校验、强权限控制、强观测和审计、强可靠性机制、强结果治理。也就是说，它现在更像是“先把工具接进来并跑通”，而不是“把工具接进来后进行严格治理”。
- **没有强校验：**  
  schema 基本直通，没有额外的兼容性检查、字段降级、默认值补全或严格参数校验。
- **没有强权限控制：**  
  它没有按 server、tool、参数范围做权限分级，也没有人工确认、白名单、租户隔离这类治理能力。
- **没有强观测和审计：**  
  现在能看见工具列表，也能跑调用，但还没有结构化审计日志、按工具统计失败率、调用耗时、参数记录、结果摘要这些工业级观测能力。
- **没有强可靠性机制：**  
  对连接失败、server 崩溃、schema 不兼容、tool 超时等问题，还没有重试、熔断、降级、回退策略。
- **没有强结果治理：**  
  tool result 基本是直通文本，没有统一结果模型、错误分层、敏感信息过滤、结果摘要或标准化。
- **离工业化还差的六层能力：**  
  权限与安全层、schema 治理层、可靠性层、结果标准化层、观测与审计层、策略与路由层。权限与安全层要解决 server、tool、参数和用户身份的授权；schema 治理层要做兼容性检查、字段裁剪、默认值和版本治理；可靠性层要处理连接失败、重试、熔断、降级；结果标准化层要统一成功/失败、错误类型、摘要和原始 payload；观测与审计层要支持记录、统计、回放和排障；策略与路由层要决定哪些工具能暴露给模型、冲突时怎么选、哪些 query 优先走本地 tool。
- **更工程化的总结：**  
  `ch04` 已经把 MCP 接入这件事跑通了，但现在主要是 `transport + adapter` 层，距离工业级 `policy + governance + reliability` 还差完整的一层系统能力。

## [重点] Q16. `policy / governance` 一般会怎么做？
一句总结：`policy / governance` 的核心不是“把工具接进来”，而是让工具的暴露、调用、参数、结果、故障和审计都处在可控边界内。

详细回答：
- **policy 和 governance 的分工：**  
  `policy` 回答的是什么情况下允许做什么，更像规则和决策逻辑；`governance` 回答的是这些规则怎么被持续执行、审计、演进和兜底，更像制度化的控制体系。放到 MCP / tool ecosystem 里，这两层通常是一起出现的。
- **工具暴露策略：**  
  一般不会把所有工具都直接暴露给模型，而是先做筛选。常见做法是按场景、用户身份、租户、环境、任务类型决定哪些 tools 可以进入当前会话。比如开发环境允许 `filesystem`，生产环境只允许只读工具。
- **权限分级：**  
  不只是“能不能用某个 tool”，还包括“能用到什么程度”。例如只读和可写的区别、只能访问某些路径、只能操作某个 workspace、只能查数据不能改数据。这通常会落到 tool 级、参数级、资源级授权。
- **参数策略与输入校验：**  
  很多风险不是来自 tool 名，而是来自参数。工业系统通常会在模型生成参数后再过一层 policy 检查，比如路径是否越界、SQL 是否包含危险操作、命令是否命中黑名单、查询范围是否过大。这层不能只信模型，也不能只信 server schema。
- **确认与审批机制：**  
  高风险动作不会让模型直接执行，而是要求用户确认，或者进入审批流。常见触发条件包括写文件、删除数据、发消息/发邮件、调用外部有成本 API、访问敏感资源。
- **结果治理：**  
  tool 返回结果后，不是全部原样透给模型或用户。通常会做敏感信息脱敏、结果截断、错误分类、摘要提取和结构化标准化。这既是安全问题，也是可用性问题。
- **可靠性策略：**  
  governance 不只是权限，还包括故障处理规则。比如哪些 server 可以自动重试，哪些错误要熔断，哪些 tool 超时后要降级，当 MCP server 不可用时是否 fallback 到本地 tool。
- **审计与可追溯：**  
  强治理一定会记录暴露了哪些工具给模型、模型选了哪个工具、传了什么参数、tool 返回了什么结果、为什么被允许或被拒绝、谁确认了高风险操作。没有这层，出了问题很难排查，也没法做事后复盘。
- **策略配置与版本化：**  
  policy 不能散落在代码里。工业系统通常会把它们配置化、版本化，例如不同环境不同策略、不同租户不同工具白名单、新策略灰度发布、策略变更可回滚。
- **评估与持续收紧：**  
  governance 不是一次写完就结束。通常会根据日志和事故不断调整，比如哪些 tool 经常被误用、哪些 description 容易误导模型、哪些参数边界需要收紧、哪些确认流程太重或太轻。
- **如果压缩成一条工程主线：**  
  可以理解为先决定模型看见什么工具，再决定模型在什么条件下能调用，再决定调用前要不要拦截或确认，再决定结果怎么处理和记录，最后用审计和评估反过来更新规则。

## [重点] Q17. 对照 MCP 官方文档，这个项目实际只实现了 MCP 能力里的哪一部分，哪些能力根本还没碰？
一句总结：这个项目当前只实现了 MCP 官方能力里的 tools 主路径，而且还是偏“静态、轻接入”的那一段；MCP 更完整的 client/server feature 集合，大部分都还没碰。

详细回答：
- **先看 MCP 官方能力地图：**  
  MCP 规范把能力大致分成几层：
  - Base Protocol（lifecycle、transports、authorization、utilities），
  - Client Features（roots、sampling、elicitation），
  - Server Features（prompts、resources、tools）。
    结合 ch04 代码，当前项目真正做到的主要是下面这些。
- **这个项目已经实现的部分：**  
  - 第一是 transport 接入，也就是 `ch04` 支持的本地 `command + stdio` 和远端 `url + streamable transport` 这两类接入方式；这对应了 MCP 的 transport 层能力，但这里只是“能连上”，不是完整 transport 治理。
  - 第二是 session 建立，代码通过 `client.Connect(...)` 建立 MCP 会话，并用 `Ping` 做存活检查，这里大概率也依赖 SDK 帮你托底了 initialize 握手；
  - 第三是真正的 tools 主路径，也就是 `tools/list`、`tools/call`，再加上把 MCP tools 包装成模型可见 schema，并映射回原始 MCP tool 去执行。这和官方 architecture 文档里 “Tool Discovery” / “Tool Execution” 那条主线是对上的。
  - 所以如果更严格一点说，`ch04` 当前实现的是 transport 接入、MCP session 建立、`tools/list`、`tools/call`，以及把 MCP tools 适配进本项目自己的 tool loop。
- **Resources 完全没碰：**  
  官方 resources 这一套包括 `resources/list`、`resources/read`、`resources/templates/list`、`resources/subscribe` 和 `notifications/resources/list_changed`。`ch04` 没有做资源发现、资源读取、资源订阅，也没有把 resource 暴露成上下文入口。
- **Prompts 完全没碰：**  
  官方 prompts 这一套包括 `prompts/list`、`prompts/get`、prompt arguments，以及 prompt list changed notifications。`ch04` 没有从 MCP server 拉 prompt template，也没有把 MCP prompt 融进本地系统 prompt 或会话构造。
- **Roots 没碰：**  
  roots 是客户端能力，允许 client 向 server 暴露“你可以在哪些文件系统根目录内操作”。`ch04` 虽然在配置里把 `${workspaceFolder}` 传给了 filesystem server，但这不是规范意义上的 roots capability 实现，它只是把路径当启动参数传给 server，而不是走 MCP `roots/list` 协议。
- **Sampling 和 Elicitation 没碰：**  
  官方 sampling 允许 server 反过来向 client 请求模型生成，也就是 `sampling/createMessage`。`ch04` 完全没有实现这条链路。elicitation 也是 client feature，允许 server 向 client 请求额外信息，当前项目也没有实现。
- **Authorization 基本没碰：**  
  spec 把 authorization 单列成 base protocol 一部分，但 `ch04` 的 `shared/mcp.go` 只是支持静态 `headers` / `url` 配置，没有真正的 OAuth、token negotiation 或 auth lifecycle 管理。所以这里最多算“预留了简单 header 注入”，不是完整授权能力。
- **动态变更、分页和 utilities 也基本没做：**  
  官方 tools/resources/prompts 都有 `listChanged` 这类能力，server 可以主动通知客户端能力集变了，`ch04` 当前没有处理这类通知，也没有自动 refresh 机制。官方 `tools/list`、`resources/list`、`prompts/list` 都支持 cursor 分页，`ch04` 也没有处理分页游标。spec 里额外 utilities 还包括 progress tracking、cancellation、error reporting 和 logging；`ch04` 里只有本地 Go `context.CancelFunc` 控制 Agent 自己的运行，但这不等于完整实现了 MCP 协议级 cancellation、progress 或 logging。
- **如果压缩成一句工程判断：**  
  `ch04` 只实现了 MCP 的“最小可用 tools 接入子集”，还没有进入 resources、prompts、roots、sampling、notifications、authorization、utilities 这些更完整的协议能力。

来源：
- MCP Specification: https://modelcontextprotocol.io/specification/2025-06-18
- MCP Architecture: https://modelcontextprotocol.io/docs/learn/architecture
- Tools: https://modelcontextprotocol.io/specification/2025-06-18/server/tools
- Resources: https://modelcontextprotocol.io/specification/2025-06-18/server/resources
- Prompts: https://modelcontextprotocol.io/specification/2025-06-18/server/prompts
- Roots: https://modelcontextprotocol.io/specification/2025-06-18/client/roots
- Sampling: https://modelcontextprotocol.io/specification/2025-06-18/client/sampling

## [重点] Q18. `resources / prompts / roots / sampling` 这几类能力，在真实 Agent 系统里分别解决什么问题，为什么 `tools` 不是 MCP 的全部？
一句总结：`tools` 只覆盖“动作执行”，而真实 Agent 系统还需要上下文资源、提示模板、边界声明和 server 反向调用模型能力，所以 `resources / prompts / roots / sampling` 这些能力共同补齐了 MCP 作为完整连接协议的角色。

详细回答：
- **`tools` 解决的是“动作能力”：**  
  它最像函数调用。模型发现可用工具，选择一个工具，带参数调用，然后拿到结果。所以 tools 适合描述查天气、读文件、查数据库、发请求、执行命令这类“执行一个动作”的事情。
- **`resources` 解决的是“上下文内容供给”：**  
  很多时候 Agent 不是要执行动作，而是要读取某些内容作为上下文。比如某个文件内容、某个文档片段、某个数据库视图、某个配置资源。这类东西更像“可读上下文对象”，而不是函数调用。如果全都硬塞成 tools，会把“读资源”和“执行动作”混在一起，语义不清，也不利于缓存、订阅和上下文管理。所以 resources 解决的是“把外部内容标准化暴露给 Agent 作为上下文来源”。
- **`prompts` 解决的是“可复用提示模板”：**  
  真实系统里，很多能力不一定要写成 tool，也不一定只是裸文本 prompt。例如代码评审模板、PR 总结模板、数据分析模板、客服回复模板，这些更像“参数化的消息模板”，而不是动作，也不是静态资源。prompts 的作用就是让 server 暴露一组可复用的 prompt 模板，client 可以列出、选择、填参数，再组装进会话。所以 prompts 解决的是“提示模板资产化、参数化和复用”。
- **`roots` 解决的是“操作边界声明”：**  
  tools 和 resources 都会碰到一个问题：server 到底可以在哪些目录、哪些文件范围里操作？roots 的意义就在这里。它不是动作能力，也不是上下文内容，而是 client 主动告诉 server 你被允许看的根目录有哪些、你应该把操作边界限制在哪里。所以 roots 解决的是“授权边界和工作空间边界表达”，这对文件系统类、代码类 Agent 很关键。
- **`sampling` 解决的是“server 反过来调用模型”：**  
  这是 MCP 里很容易被忽略的一类能力。tools、resources、prompts 大多是 server 把东西提供给 client；sampling 则允许 server 反过来向 client 请求一次 LLM 生成。也就是说，某个 MCP server 不只是被动提供工具，它还可以在内部工作流里说“帮我调一次模型，我需要一次生成、总结、分类或判断”。这让 server 自己也能变成一个更复杂的 agentic 组件，而不是一个纯工具箱。所以 sampling 解决的是“server 侧复用 client 的模型能力”。
- **把这几类能力压成一张图：**  
  可以理解成 `tools` 让模型去做动作，`resources` 让模型拿到上下文内容，`prompts` 让系统复用参数化提示模板，`roots` 让 server 知道可操作边界，`sampling` 让 server 反向请求模型能力。
- **为什么 `tools` 不是 MCP 的全部：**  
  因为真实 Agent 系统至少有五个维度：要做什么动作、要读什么上下文、要用什么模板组织交互、能在哪些边界内操作，以及 server 自己要不要反向借用模型能力。只做 tools，最多解决第一个问题；而 MCP 想标准化的，其实是整个“模型应用和外部能力/上下文系统的连接面”，不是单一的函数调用。

## Q19. 在真实 Agent 系统里，什么时候更适合用 `tools`，什么时候更适合用 `resources / prompts`，而不是把一切都做成 `tool`？
一句总结：做事用 `tools`，读内容用 `resources`，复用提示结构用 `prompts`；不要把一切都做成 tool，因为真实 Agent 系统不只有动作，还有上下文、模板和边界。

详细回答：
- **适合用 `tools` 的场景：**  
  当你要表达的是“做一件事”时，用 `tools`。典型特征是有明确输入参数、会触发执行、可能有副作用、返回的是一次调用结果。比如查数据库、发请求、写文件、执行搜索、发送消息。能被理解成一次函数调用的，优先是 tool。
- **适合用 `resources` 的场景：**  
  当你要表达的是“提供一段可读取的上下文内容”时，用 `resources`。它更像“数据源”或“上下文对象”，而不是动作。比如某个文档内容、某个配置文件、某个知识库条目、某个数据库只读视图。重点是让模型读内容，而不是执行动作时，用 resource。
- **适合用 `prompts` 的场景：**  
  当你要表达的是“可复用的提示模板”时，用 `prompts`。它适合沉淀一类固定工作流的交互骨架，而不是一个外部动作。比如 PR review 模板、Incident 总结模板、客服回复模板、数据分析报告模板。重点是复用一套参数化提示结构时，用 prompt。
- **不要把一切都做成 `tool` 的原因：**  
  因为 `tool` 的语义太强，天然带“调用”和“执行”味道。如果把读文档、拿模板、声明边界都硬塞成 tool，会让动作和内容混在一起，上下文管理变差，缓存、订阅、复用都不自然，模型也更难判断“该读内容”还是“该执行动作”。
- **一个简单判断法：**  
  先问四个问题：这是要让模型做事吗？是，就偏 `tool`；这是要让模型读内容吗？是，就偏 `resource`；这是要让系统复用提示模板吗？是，就偏 `prompt`；这是在声明可操作边界吗？是，就偏 `roots`。
- **工程上常见误区：**  
  很多系统早期会把一切都做成 tool，因为最省事。但系统一复杂，就会发现读文件不等于执行动作，模板不等于工具，边界不等于参数，server 反向调模型也不等于 tool。所以 MCP 才把这些能力拆开。
