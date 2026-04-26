# 第七章学习笔记

## 学习导航

### 1. 本章需要学习的内容

- **Agentic RAG 的核心变化：**
  本章不是只讲“把文档检索出来再回答”，而是讲让 Agent 自主决定是否检索、检索什么、检索多少、是否需要重排，以及如何把检索结果作为工具结果交给模型继续推理。你要重点区分传统 RAG 和 Agentic RAG：前者通常是固定流程，后者把检索动作变成 Agent 可决策的工具行为。

- **RAG 管道的基本组件：**
  这一章的主线是 `chunking -> embedding -> vector store -> search -> rerank -> tool`。你要理解每一步解决的问题：chunking 解决文档太长和语义稀释；embedding 把文本变成向量；vector store 负责近似相似搜索；rerank 用更精细的模型重新排序候选；tool 把检索能力暴露给 Agent。

- **代码仓库索引系统：**
  README 里把本章定位成一个代码仓库索引系统。你要观察它如何遍历文件、过滤可索引文件、按 chunk 切分、并发调用 embedding、批量写入 pgvector，并通过文件修改时间实现增量更新。

- **语义搜索工具如何接入 Agent：**
  `semantic_search` 是本章最接近 Agent 能力的入口。它把“用户问题 -> query embedding -> vector search -> rerank -> 格式化结果”封装成一个工具。模型是否调用这个工具、传什么 query、传多少 `top_k`，才是 Agentic RAG 和普通 RAG 的关键差异。

- **RAG 和上下文工程、记忆系统的关系：**
  `ch05` 解决有限上下文窗口里如何保留、压缩、卸载；`ch06` 解决跨会话长期信息如何沉淀；`ch07` 开始解决外部知识如何按需召回。你要把三章串起来看：context 是当前窗口治理，memory 是长期状态治理，RAG 是外部知识召回。

### 2. 本章操作内容

- **准备 PostgreSQL + pgvector：**
  本章需要 PostgreSQL 安装 `pgvector` 扩展。README 指向官方文档：
  - https://github.com/pgvector/pgvector

- **配置 Embedding 和 Rerank 服务：**
  本章使用兼容 OpenAI API 风格的 embedding 和 rerank HTTP 服务。重点看 `HTTPEmbeddingConfig` 和 `HTTPRerankConfig`，确认 `BaseURL`、`APIKey`、`Model`、`Dimensions` 是否和你本地服务匹配。

- **执行索引流程：**
  README 当前主要给出了库级用法，而不是完整 `main` 入口。学习时可以先沿着 `Indexer` 的调用链理解流程：创建 `PGVectorStore`，创建 `HTTPEmbeddingService`，创建 `Indexer`，调用 `Index()` 或 `IndexConcurrent()`。

- **测试语义搜索工具：**
  索引完成后，`SemanticSearchTool.Execute()` 会先对 query 做 embedding，再从向量库召回 `top_k * 2` 个候选，然后用 rerank 重排，最后格式化成模型可读的文本结果。

- **运行前置注意：**
  当前本地 `go test ./ch07/...` 被 Go 工具链拦住，报错是 `go.mod` 中 `go 1.25.0` 不被当前 Go 版本解析。需要使用 Go 1.25+ 或调整本地 Go 版本后再验证。

- **实现一致性注意：**
  当前 `ch07/rag` 目录里的 Go 文件声明为 `package shared`，但其他文件按 `babyagent/ch07/rag` 导入，并在代码中使用 `shared.*` 标识。这说明代码很可能处在章节草稿状态，后续运行前需要检查包名或导入别名是否一致。

### 3. 本章学完需要掌握的知识点

- **能解释 Agentic RAG 和传统 RAG 的差异。**
  传统 RAG 多数是固定检索链路；Agentic RAG 把检索作为工具交给模型，让模型根据任务状态决定是否检索、怎么检索、是否再次检索。

- **能解释为什么 chunking 是 RAG 的前提。**
  文档切分不是简单优化，而是为了避免 embedding 输入超长、整文档语义稀释、召回结果无法放入上下文、定位不精确、增量更新成本过高。

- **能解释 embedding 和 rerank 的分工。**
  embedding 适合快速召回，先从大规模向量库里找到可能相关的候选；rerank 适合精排，用 query 和候选内容做更细粒度相关性判断。两者不是替代关系，而是召回和排序的两阶段协作。

- **能解释 pgvector 在本章中的角色。**
  pgvector 不是模型能力，而是向量存储和相似度检索基础设施。它负责保存 chunk embedding，并用 `<=>` 等向量距离算子完成近似或精确搜索。

- **能解释代码索引为什么需要增量更新。**
  代码仓库会不断变化，向量索引只是某个时间点的快照。如果没有修改时间判断、删除旧索引和重新索引机制，RAG 会检索到过期代码，进而误导 Agent。

- **能解释为什么 Coding Agent 不一定优先用 embedding。**
  README 里这个问题很关键。代码检索有大量精确标识符、文件系统永远最新、grep/read/find 成本低且可解释，所以很多 Coding Agent 更偏向工具化的精确搜索，而不是默认构建向量索引。

### 4. 扩展阅读与参考资料总结

- **pgvector GitHub：**
  原始链接：
  - https://github.com/pgvector/pgvector

  这份资料要重点看 pgvector 支持的向量类型、距离算子、IVFFlat/HNSW 索引、索引参数和查询写法。它能帮助你理解本章 `embedding <=> ?` 这类 SQL 背后的含义。

- **LangChain RAG Tutorial：**
  原始链接：
  - https://python.langchain.com/docs/tutorials/rag/

  这份资料适合对照本章理解标准 RAG pipeline，包括加载文档、切分、embedding、vector store、retriever、把检索结果注入 prompt。你要重点看成熟框架如何拆分 loader、splitter、retriever、chain。

- **The Retrieval-Augmented Generation (RAG) Pattern：**
  原始链接：
  - https://arxiv.org/abs/2005.11401

  这是 RAG 原始论文，重点不是工程代码，而是理解 RAG 的基本思想：生成模型不只依赖参数内知识，而是在生成时结合外部检索到的文档。它能帮助你把本章工程实现和研究概念连起来。

- **Vector Database Comparison：**
  原始链接：
  - https://www.pinecone.io/learn/vector-database/

  这份资料适合理解为什么会有专门的向量数据库，以及不同向量数据库在存储、索引、过滤、扩展性、运维成本上的差异。你要重点看“为什么不是普通数据库直接存数组”。

- **Faiss: A library for efficient similarity search：**
  原始链接：
  - https://github.com/facebookresearch/faiss

  Faiss 适合理解向量索引算法本身，尤其是 IVF、PQ、HNSW 这类近似最近邻搜索。它和 pgvector 不是同一类产品形态，但能帮助你理解向量检索为什么需要索引，以及速度和召回率之间的权衡。

### 5. 扩展阅读需要重点学习什么

- **重点一：RAG 不是一个单点能力，而是一条工程链路。**
  你需要关注每个环节如何影响最终效果。chunk 切得不好，embedding 再强也可能召回错；召回数量不合理，rerank 也没有足够候选；注入太多结果，上下文会被噪声污染。

- **重点二：检索质量需要评测，不应该只靠直觉。**
  README 最后强调“策略必须先评测再上线”。你要特别记住：chunk size、top_k、embedding model、rerank model、索引参数都不是凭感觉选的，而应该通过可复现实验比较效果和成本。

- **重点三：向量检索解决的是语义相似，不等于事实正确。**
  RAG 可以减少幻觉，但不能消灭幻觉。检索结果可能过时、召回不完整、排序错误，模型也可能误读结果。因此 RAG 系统仍然需要来源标注、置信度、过滤、上下文治理和失败处理。

- **重点四：代码场景下要重新评估 RAG 的收益。**
  对自然语言知识库，embedding 很有价值；对快速变化的代码仓库，grep/read/find 往往更可靠。你要学会判断任务形态，而不是把 embedding 当成所有检索问题的默认答案。

### 6. 本章应掌握的核心知识点

- **Agentic RAG 的本质是把检索变成模型可调度的行动。**
  检索不再只是用户问题到回答之间的固定中间层，而是 Agent 可以根据当前任务自主选择的一种工具。

- **RAG 的质量主要由检索链路决定，不只由大模型决定。**
  很多 RAG 问题表面看是模型答错，实际是切分、召回、重排、过滤、注入任一环节出了问题。

- **向量相似度是近似语义匹配，不是确定性查找。**
  向量检索适合找“意思相近”的内容，但不适合替代所有精确检索。代码标识符、配置键、错误码、函数名这类内容，精确搜索通常更稳。

- **RAG 和 memory 的边界要分清。**
  RAG 召回的是外部知识库里的文档片段，memory 保存的是系统从历史交互中沉淀出的长期状态。它们都能进入上下文，但来源、生命周期、可信度和治理方式不同。

- **本章仍是教学版 RAG，而不是完整工业 RAG 平台。**
  它已经覆盖 chunking、embedding、pgvector、rerank、semantic search tool、索引增量更新，但还没有完整评测集、混合搜索、权限控制、索引新鲜度监控、召回解释、查询改写、多轮检索规划等工业能力。

### 7. 建议代码阅读顺序

- **先看 [README.md](/Users/bytedance/vibe-coding/baby-agent/ch07/README.md)：**
  先把 Agentic RAG 的整体链路读清楚。重点关注传统 RAG 和 Agentic RAG 的差异、chunking 的必要性、embedding 与 rerank 的分工、以及最后“Coding Agent 为什么不用 Embedding，而用 Grep”的讨论。

- **再看 [type.go](/Users/bytedance/vibe-coding/baby-agent/ch07/rag/type.go)：**
  这里定义了 RAG 系统的核心抽象：`Vector`、`Chunk`、`Meta`、`VectorPoint`、`VectorStore`、`EmbeddingService`、`RerankService`、`ChunkerService`。先读接口能避免一上来陷入实现细节。

- **然后看 [chunker.go](/Users/bytedance/vibe-coding/baby-agent/ch07/rag/chunker.go)：**
  重点观察 `LineChunker` 和 `ParagraphChunker` 的边界处理。你要问自己：什么情况下会切断语义？为什么要保留 `StartPos` / `EndPos`？为什么 chunk 的元数据和内容同样重要？

- **再看 [embedding.go](/Users/bytedance/vibe-coding/baby-agent/ch07/rag/embedding.go) 和 [rerank.go](/Users/bytedance/vibe-coding/baby-agent/ch07/rag/rerank.go)：**
  这两个文件分别对应召回和精排。重点看 HTTP 请求格式、模型配置、维度配置、返回结果解析，以及失败时系统会如何表现。

- **接着看 [pgvector.go](/Users/bytedance/vibe-coding/baby-agent/ch07/db/pgvector.go)：**
  这里是向量存储落地。重点观察表结构、`Embedding` 字段、`InsertBatch`、`Search`、`DeleteByDocument`、`GetDocumentIndexedTime`，以及 SQL 里 `1 - (embedding <=> ?)` 和 `ORDER BY embedding <=> ?` 的含义。

- **然后看 [file_walker.go](/Users/bytedance/vibe-coding/baby-agent/ch07/index/file_walker.go)：**
  这里决定哪些文件会进入索引。重点看排除目录和扩展名白名单。这个文件看似简单，但它决定了知识库的边界。

- **再看 [indexer.go](/Users/bytedance/vibe-coding/baby-agent/ch07/index/indexer.go)：**
  这是构建向量索引的主流程。重点观察 `Index()`、`IndexConcurrent()`、`indexFile()`、`embedChunks()`，特别是增量更新逻辑和并发 embedding 的错误处理。

- **最后看 [semantic_search.go](/Users/bytedance/vibe-coding/baby-agent/ch07/tool/semantic_search.go)：**
  这是 Agentic RAG 的工具入口。重点观察 tool schema 怎么描述 `query` 和 `top_k`，以及 `Execute()` 如何完成 query embedding、向量搜索、rerank 和结果格式化。

- **学习时优先盯这 6 个观察点：**
  你要持续问自己：Agent 为什么要自己决定是否检索；chunk 粒度如何影响召回质量；embedding 和 rerank 为什么要分两阶段；向量索引为什么会过时；RAG 结果如何进入上下文而不污染上下文；代码检索什么时候应该用 grep 而不是 embedding。

## Q&A

### Q1. `ch07/rag` 里的接口关系是什么？它们分别说明了哪些设计边界？

- **一句总结：**
  `ch07/rag` 里的接口比结构体更能说明 RAG 的设计边界：`ChunkerService` 负责切分，`EmbeddingService` 负责向量化，`VectorStore` 负责存储和召回，`RerankService` 负责精排；它们把 RAG 拆成可替换的四段能力。

- **详细回答：**
  这些接口不是 struct，但它们更能说明这一章的系统设计。结构体回答的是“数据长什么样”，接口回答的是“能力边界在哪里”。在 RAG 系统里，真正重要的是把切分、向量化、存储召回、重排序拆开，这样每一层都可以独立替换。

```go
// 文档 -> 文档块
type ChunkerService interface {
    Chunk(documentID, content string) []Chunk
}

// 文本 -> 向量
type EmbeddingService interface {
    Embed(ctx context.Context, chunk string) (Vector, error)
}

// query + 候选 chunks -> 重排序后的 chunks
type RerankService interface {
    Rerank(ctx context.Context, query string, candidates []Chunk) ([]Chunk, error)
}

// 向量点存储、搜索、删除和索引状态管理
type VectorStore interface {
    InsertBatch(ctx context.Context, vps []VectorPoint) error
    Search(ctx context.Context, queryVector Vector, limit int) ([]VectorPointResult, error)
    DeleteByDocument(ctx context.Context, documentID string) error
    GetDocumentIndexedTime(ctx context.Context, documentID string) (time.Time, error)
    Clear(ctx context.Context) error
    Close() error
}
```

  - **`ChunkerService` 的边界是“文档如何变成检索单元”。**
    它不关心 embedding，也不关心数据库。它只负责把一个完整文档拆成多个 `Chunk`。当前有 `LineChunker` 和 `ParagraphChunker` 两种实现。后面如果要做 AST chunking，只需要新增一个实现，而不需要改 embedding 或 vector store。

  - **`EmbeddingService` 的边界是“文本如何变成向量”。**
    它不关心文本来自哪个文件，也不关心向量存到哪里。它只负责把一段文本转换成 `Vector`。当前实现是 `HTTPEmbeddingService`，后面可以替换成 OpenAI embedding、本地 embedding 模型、批量 embedding 服务，调用方都不需要知道细节。

  - **`VectorStore` 的边界是“向量如何被保存和召回”。**
    它不负责生成向量，也不负责判断 chunk 怎么切。它只负责把 `VectorPoint` 写进去，再根据 query vector 搜出 `VectorPointResult`。当前实现是 pgvector，但后面可以换成 Milvus、Pinecone、Faiss、Chroma，只要接口不变，上层索引器和工具就不需要大改。

  - **`RerankService` 的边界是“召回结果如何重新排序”。**
    它接收 query 和候选 `Chunk`，返回排序后的 `Chunk`。这说明 rerank 是召回之后的精排阶段，而不是替代 embedding 或 vector search。当前实现是 HTTP rerank 服务，后面也可以换成本地 cross-encoder、规则排序，或者直接置空跳过 rerank。

  这四个接口串起来，就是本章 RAG 的主链路。索引阶段是把文档变成可搜索的向量点，查询阶段是把用户问题变成 query vector，再从向量库里召回并重排相关 chunk。

```text
[索引阶段]
原始文档 -> ChunkerService -> []Chunk -> EmbeddingService -> []VectorPoint = Chunk + Vector -> VectorStore.InsertBatch() -> 向量数据库

[查询阶段]
用户 query -> EmbeddingService.Embed(query) -> queryVector -> VectorStore.Search(queryVector, limit) -> []VectorPointResult -> 提取 []Chunk -> RerankService.Rerank(query, chunks) -> 排序后的相关 Chunk
```

  这个设计的好处是每一层都有明确职责。chunking 策略不好，就换 chunker；embedding 模型效果不好，就换 embedding service；pgvector 性能不够，就换 vector store；召回相关性不稳定，就加 rerank 或换 rerank 模型。RAG 系统的复杂度不会挤在一个大对象里。

  所以更准确地说，`ch07/rag` 的接口设计是在告诉你：RAG 不是一个函数调用，而是一条可拆、可替换、可调参的工程管道。理解这些接口，比只记住 `Chunk`、`VectorPoint` 这些结构体更重要。

### Q2. Embedding 到底是什么，为什么能表示语义相似？

- **一句总结：**
  Embedding 是把文本映射成一串数字向量的表示方法；模型在训练中学会让语义相近的文本在向量空间里距离更近，所以我们可以用向量相似度来做语义检索。

- **详细回答：**
  你可以先把 embedding 理解成：

```text
文本 -> 模型 -> 向量
```

  比如：

```text
"如何处理错误" -> [0.12, -0.03, 0.88, ...]
"error handling in Go" -> [0.10, -0.01, 0.84, ...]
"今天吃什么" -> [-0.45, 0.72, 0.02, ...]
```

  前两个句子虽然字面不同，一个中文一个英文，但语义相近，所以它们的向量会更接近。第三个句子语义无关，所以向量距离会更远。

  LLM 或 embedding 模型本质上会把 token 变成内部隐藏状态。模型训练过程中，会逐渐学到词、短语、句子在语义上的关系。Embedding 模型做的事情，可以理解为把一段文本压缩成一个固定长度的语义坐标：

```text
文本含义 -> 高维空间中的一个点
```

  这个高维空间可能是 512 维、1536 维或其他维度。每一维不是人类可直接解释的“是否是代码”“是否是食物”，而是模型训练后形成的分布式语义特征。单独看某一维通常没意义，但整体向量之间的距离有意义。

  为什么语义相近会距离更近？因为 embedding 模型通常会经过对比学习、匹配学习或类似目标训练。训练时会给模型很多正负样本：

```text
正样本：
query: "如何处理错误"
doc:   "Go 中 error handling 的最佳实践"

负样本：
query: "如何处理错误"
doc:   "今天晚饭吃什么"
```

  训练目标大概是：让 query 和相关 doc 的向量更接近，让 query 和无关 doc 的向量更远。经过大量训练后，模型就形成了一个语义空间。在这个空间里，同义表达更近，相关主题更近，不同语言但语义相近也可能更近，无关主题更远。

  所以 RAG 才能做：

```text
用户 query -> query embedding
文档 chunk -> chunk embedding
计算 query 和 chunk 的距离
取距离最近的 chunk
```

  这里的“距离”常见有三种：`cosine similarity` 看两个向量方向是否接近，`dot product` 看向量点积大小，`euclidean distance` 看两个点的直线距离。`ch07` 用的是 pgvector 的 cosine distance：

```sql
embedding <=> ?
```

  然后代码里算：

```sql
1 - (embedding <=> ?) as score
```

  意思是距离越小，相似度分数越高。

  一个极简二维例子是：

```text
"删除用户"        -> [0.90, 0.10]
"移除账号"        -> [0.88, 0.12]
"创建订单"        -> [0.10, 0.85]
"今天下雨"        -> [-0.20, 0.40]
```

  那么“删除用户”和“移除账号”方向接近，语义相近；“删除用户”和“创建订单”方向差很多，语义不同。真实 embedding 不是二维，而是几百维或几千维，但核心直觉类似：向量方向接近，大概率表示语义接近。

  Embedding 不是关键词匹配，这是很重要的点。关键词搜索依赖字面重合：如果 query 是“删除用户”，doc 是“移除账号”，没有同样的词，关键词搜索可能搜不到。Embedding 搜的是语义接近，所以“删除用户”“移除账号”“remove account”“delete user”可能都会比较近。

  但 embedding 不是事实判断。它只能说明“语义相似”，不能说明“事实正确”。例如：

```text
query: "如何删除用户"
doc1: "删除用户 API 是 DELETE /users/:id"
doc2: "创建用户 API 是 POST /users"
doc3: "删除订单 API 是 DELETE /orders/:id"
```

  embedding 可能觉得 `doc3` 也有点相关，因为它包含“删除”和“API”概念，但它不是正确答案。所以 embedding 召回通常只是第一步，还需要 rerank、metadata filter、source check、LLM 阅读原文，必要时还需要工具验证。

  代码场景里 embedding 还有额外局限。代码里很多语义是精确符号，不是自然语言近似。比如你要找 `deleteUser`，`grep` 可以精确找到所有定义和调用；embedding 可能会把 `removeAccount`、`deleteOrder`、`createUser`、`userRepository` 也召回。它们语义相关，但不一定是你要的精确目标。

  所以代码场景里要分清：自然语言解释、文档、注释用 embedding 有价值；函数名、调用点、变量名、配置键通常用 `grep / LSP` 更可靠。

  和 `ch07` 的关系是，embedding 出现在两处。索引阶段是：

```text
chunk.Content -> EmbeddingService.Embed() -> Vector
Vector + Chunk -> VectorStore.InsertBatch()
```

  查询阶段是：

```text
query -> EmbeddingService.Embed() -> queryVector
queryVector -> VectorStore.Search()
```

  也就是说，文档 chunk 和用户 query 都必须映射到同一个向量空间，才能比较距离。如果文档用一个 embedding 模型，query 用另一个 embedding 模型，向量空间就不一致，相似度就没有意义。所以同一个向量库里的文档向量和查询向量必须来自同一种 embedding 模型或兼容模型。

### Q3. Chunk size 应该怎么选？为什么 chunk 太大或太小都会出问题？

- **一句总结：**
  Chunk size 没有固定最优值，它是在语义完整性、检索精度、上下文噪声、召回成本之间做权衡；太小会丢上下文，太大会稀释语义并污染 prompt。

- **详细回答：**
  RAG 里 chunk size 决定的是每一个可检索文本单元有多大。比如 `ch07` 代码里 `maxLines = 100`、`maxChars = 2000`，就表示一个 chunk 最多大约 100 行或 2000 字符。

  不能直接整篇文档做 embedding，是因为整篇文档通常太长，而且主题太多。Embedding 模型有输入长度限制；一篇文档里如果包含很多主题，向量会变成“平均语义”；用户只问一个细节，但召回回来的是一大段无关内容；放进 LLM 上下文时会浪费 token；结果定位也不精确，不知道答案具体来自哪里。所以需要切 chunk。

  chunk 太小会有几个问题。假设你把代码按 3 行一块切：

```go
func DeleteUser(id string) error {
    return repo.Delete(id)
}
```

  可能被切成：

```text
Chunk 1:
func DeleteUser(id string) error {

Chunk 2:
    return repo.Delete(id)

Chunk 3:
}
```

  这样每个 chunk 都太碎，`return repo.Delete(id)` 单独看不知道属于哪个函数。Embedding 只能看到很短的片段，模型很难知道它和“删除用户”有关。一个完整答案可能分散在多个 chunk 里，`top_k` 不够时就漏信息。模型拿到碎片后还要自己拼上下文，容易误解。并且 chunk 太小会导致向量数量暴涨，存储、检索、rerank 成本都会上升。

  chunk 太大也有问题。假设你把一个 1000 行文件切成一个 chunk，一个 chunk 里可能同时有用户登录、订单创建、错误处理、日志、测试等很多内容。Embedding 只能给它一个向量，这个向量会变成混合语义。用户问“错误处理”，这个大 chunk 可能只是“有点相关”，但不够精准。

  chunk 太大还会带来召回噪声和上下文压力。命中的 chunk 里大部分内容可能和问题无关，LLM 会被迫在一堆噪声里找答案。`top_k = 5` 时，如果每个 chunk 都很大，很快就把 prompt 塞满。用户问某个小函数，结果返回整个文件，定位也不精确。rerank 通常要读 query 和 candidate 文本，candidate 越大，成本越高、延迟越高。

  核心权衡可以这样记：

```text
chunk 太小：
召回精确度可能高，但语义上下文不足，容易漏信息。

chunk 太大：
上下文完整性更好，但语义稀释、噪声增多、成本上升。
```

  不同内容应该用不同策略。代码文件最好按函数、类、方法、接口这种语义单元切，行切分只是简单版本。Markdown 和文档更适合按标题、段落、列表切。配置文件适合按对象、key group、resource block 切。日志适合按时间窗口、请求 ID、trace ID、错误块切。论文或长文可以按章节、段落，再加摘要块。

  很多 RAG 还会做 overlap，比如：

```text
chunk size = 1000 tokens
overlap = 100 tokens
```

  overlap 的好处是避免关键信息刚好被切在边界，保留前后文，提高召回连续片段的概率。代价是存储变多、embedding 成本变高、重复内容可能进入上下文，搜索结果也可能出现多个高度相似 chunk。所以 overlap 不是越大越好，它主要是为了缓解边界问题。

  chunk size 不能单独调，它必须和 `top_k` 一起看：

```text
小 chunk + 小 top_k：容易漏上下文
小 chunk + 大 top_k：召回更多碎片，但上下文需要拼接
大 chunk + 小 top_k：覆盖面大，但噪声多
大 chunk + 大 top_k：最容易塞爆上下文
```

  所以 chunk size、`top_k`、rerank、context window 是一组参数。

  `ch07` 当前有两个 chunker：`LineChunker` 按行数和字符数切，`ParagraphChunker` 按段落切。`LineChunker` 默认 `maxLines = 100`、`maxChars = 2000`。这是一种教学版策略：实现简单、可观察、适合代码文件初步索引。但它不保证语义边界。比如一个函数可能被切断，或者多个不相关函数被塞进同一个 chunk。更好的代码 RAG 通常会用 AST chunking、function-level chunking、class-level chunking、symbol-aware chunking。

  选择初始值可以从经验范围开始：

```text
短问答 / 精确定位：300-800 tokens
普通文档问答：500-1200 tokens
需要上下文的技术文档：1000-2000 tokens
代码：按函数/类优先；如果只能按行，先用 50-150 行
```

  但这只是起点，不是最终答案。真正的选择方式不是拍脑袋，而是做评测：准备一批问题，标注正确答案来源，尝试不同 chunk size、overlap、top_k，比较 `recall@k`、`MRR`、`NDCG`、答案正确率、上下文 token、延迟和成本。

  如果 chunk size 改了以后，正确 chunk 更容易被召回，`top_k` 里噪声更少，LLM 答案更稳定，token 成本没有明显上升，那这个策略才算更好。所以 chunk size 的本质不是“选一个数字”，而是围绕文档类型和任务目标调检索系统。对于 `ch07`，你应该先理解默认策略只是教学起点，后面真正值得做的是语义边界切分和可复现评测。

### Q4. 向量数据库中存储的不只是向量，还有原始文本吗？

- **一句总结：**
  是的，`ch07` 的向量数据库里不只存向量，也存原始 chunk 文本和来源元数据；向量负责相似度搜索，原文负责被模型读取和回答。

- **详细回答：**
  在 `ch07` 里，向量数据库存的不是裸向量，而是“向量 + 原始文本 + 来源元数据”。具体结构在 [pgvector.go](/Users/bytedance/vibe-coding/baby-agent/ch07/db/pgvector.go)：

```go
type DocumentChunk struct {
    ID         uint
    Content    string
    DocumentID string
    StartPos   int
    EndPos     int
    Embedding  string
    CreatedAt  time.Time
}
```

  这些字段分别表示：

```text
Content     原始 chunk 文本
DocumentID  文档 ID，通常是文件路径
StartPos    chunk 起始位置，代码里通常是起始行号
EndPos      chunk 结束位置
Embedding   chunk 的向量
CreatedAt   写入索引的时间
```

  这样设计的原因很直接：向量只适合做相似度搜索，但它本身不能给模型读。模型最后需要阅读的是原始文本，而不是一串浮点数。向量搜索命中之后，系统必须能把对应 chunk 的原文和来源位置取出来，再格式化成工具结果交给 LLM。

```text
用户 query
-> query embedding
-> 向量库相似度搜索
-> 命中 document_chunks 里的某些行
-> 取出 Content / DocumentID / StartPos / EndPos
-> 格式化成工具结果
-> LLM 阅读这些原始文本后回答
```

  如果只存向量，不存原文，系统只能知道“某个向量和 query 很相似”，但无法把有意义的内容返回给模型。除非它额外根据 `document_id + start/end` 回源文件系统或文档服务读取原文。

  `ch07` 选择把原始 chunk 也存进 pgvector，是教学版和中小规模系统里很常见的做法。这样实现简单，搜索命中后不用再回源文件，结果可以直接格式化给 Agent，并且 `DocumentID / StartPos / EndPos` 还能告诉用户内容来自哪里。

  但工业系统里也可能采用另一种设计：向量库里只存 `vector`、`document_id`、`chunk_id`、metadata，命中后再去对象存储、文件系统、文档服务里拉原文。这样可以减少向量库体积，也方便权限控制、原文更新和多版本管理。

  所以更准确地说，向量数据库不一定必须存原文，但 RAG 系统一定需要在搜索命中后拿到原文。`ch07` 的选择是把原文和向量放在同一张表里，这是最容易理解和调试的实现方式。

### Q5. pgvector 里 SQL 的 `embedding <=> vector` 是什么意思？只能计算余弦距离吗？

- **一句总结：**
  `embedding <=> vector` 是 pgvector 的余弦距离运算，表示比较数据库里某条 embedding 向量和查询向量之间的距离；pgvector 不只支持余弦距离，也支持 L2、inner product、L1 等距离或相似度形式。

- **详细回答：**
  在 [pgvector.go](/Users/bytedance/vibe-coding/baby-agent/ch07/db/pgvector.go) 里有这段 SQL：

```sql
SELECT id, content, document_id, start_pos, end_pos,
       1 - (embedding <=> ?) as score
FROM document_chunks
ORDER BY embedding <=> ?
LIMIT ?
```

  这里的核心是：

```sql
embedding <=> ?
```

  其中 `embedding` 是数据库表里的向量字段，`?` 是查询向量 `queryVector`，`<=>` 是 pgvector 提供的 cosine distance 运算符。也就是说，它在比较数据库中每个 chunk 的 embedding 和用户 query 的 embedding。

  `ORDER BY embedding <=> ?` 表示按“离 query 向量最近”排序。因为 distance 越小，表示两个向量越接近，语义越相似，所以最相似的 chunk 会排在最前面。

  pgvector 的 `<=>` 返回的是距离，不是相似度。距离越小越好，但业务层通常更习惯看 score 越大越好，所以代码里写了：

```sql
1 - (embedding <=> ?) as score
```

  也就是把距离转成更直观的分数：

```text
distance = 0.10 -> score = 0.90
distance = 0.35 -> score = 0.65
distance = 0.80 -> score = 0.20
```

  cosine distance 关注的是两个向量方向是否接近，而不是长度是否接近。可以粗略理解成：两个向量方向越接近，语义越相似，cosine distance 越小；两个向量方向越远，语义越不同，cosine distance 越大。

  假设 query 是“如何更新 memory”，它的 query vector 和数据库里的几个 chunk 比较：

```text
chunk A: "MemoryUpdater 如何更新长期记忆"
distance = 0.08
score = 0.92

chunk B: "System prompt 如何注入 memory"
distance = 0.24
score = 0.76

chunk C: "Bubble Tea TUI 如何渲染消息"
distance = 0.71
score = 0.29
```

  SQL 排序后会是 A、B、C，因为 A 距离 query 最近。

  需要注意，`<=>` 不是普通 SQL 里的“小于等于”，也不是判断两个字段是否相等。它是 pgvector 定义的向量距离运算符。

  pgvector 也不是只能计算余弦距离。常见运算符包括：

```text
<->   L2 distance / Euclidean distance，欧氏距离
<#>   negative inner product，负内积
<=>   cosine distance，余弦距离
<+>   L1 distance / taxicab distance，曼哈顿距离
```

  如果换成欧氏距离，可以写：

```sql
ORDER BY embedding <-> ?
```

  如果换成 inner product，可以写：

```sql
ORDER BY embedding <#> ?
```

  但 `<#>` 返回的是 negative inner product，因为 PostgreSQL 索引默认按升序扫描，所以数值越小代表原始 inner product 越大。

  对应索引也要匹配距离类型。`ch07` 当前建的是：

```sql
USING ivfflat (embedding vector_cosine_ops)
```

  这是 cosine ops。如果想用 L2，需要建：

```sql
USING ivfflat (embedding vector_l2_ops)
```

  如果想用 inner product，需要建：

```sql
USING ivfflat (embedding vector_ip_ops)
```

  所以不是只改 SQL 运算符就完了，索引 operator class 也要对应。

  怎么选距离类型，通常看 embedding 模型建议和评测结果。一般来说，cosine distance 最常用于文本 embedding，关注向量方向，弱化长度影响；inner product 常用于模型训练时就是按点积优化的 embedding，或者向量已归一化时；L2 distance 适合某些数值向量、图像向量，或模型文档明确推荐 L2 时；L1 distance 较少用于文本 embedding。

  对 `ch07` 这种文本 RAG 来说，cosine 是合理默认选择。但真正上线时，还是应该看 embedding 模型官方建议，并用离线评测比较不同距离类型的检索质量。

### Q6. RAG 里为什么需要 Query Rewrite / Query Planning？它和直接 embedding 用户原始问题有什么区别？

- **一句总结：**
  Query Rewrite / Query Planning 是 RAG 进入“可用系统”的关键步骤：它解决的不是“怎么搜”，而是“用户这句话到底应该被转换成什么检索任务”。

- **详细回答：**
  在最简单的 RAG 里，流程通常是用户 query 直接 embedding，然后去向量库里做 top_k 检索，再把结果塞给模型生成答案。这个流程能跑，但它有一个很大的隐含假设：用户原始 query 本身就是一个好的检索 query。现实里这个假设经常不成立。

  比如用户问“为什么这里流式返回没有显示 tool result？”。如果直接拿这句话做 embedding，系统可能搜到 streaming、tool result、TUI 相关内容，但它不一定知道应该进一步拆成几个更具体的问题：流式事件在哪里被消费，tool call 和 tool result 分别在哪个结构里表示，TUI 展示层有没有刻意隐藏 tool result，Agent 是否把 tool result 只回填到上下文。Query Planning 要做的就是把一个用户问题拆成更适合检索、阅读和调用工具的子任务。

  Query Rewrite 更偏“改写检索词”。它会把用户原始问题改写成更适合召回的形式。比如用户问“这个 search 怎么实现的？”，直接 embedding 这句话太模糊，因为“这个”依赖上下文，“search”也可能指很多东西。改写后可能变成“ch07 中 semantic_search 工具如何调用 embedding、vector store 和 rerank 完成检索”，也可能生成更适合代码检索的关键词：`semantic_search`、`EmbeddingProvider`、`VectorStore`、`Reranker`。

  Query Rewrite 常见做法包括补全上下文、扩展同义词、生成关键词、把口语问题改写成更接近文档或代码表达的检索句。它的目标不是直接回答问题，而是提高召回质量。

  Query Planning 比 Query Rewrite 更进一步。它不是只改写一句 query，而是决定这个问题应该怎么查、查几次、查哪些数据源、每次查什么、查完怎么合并。比如用户问“ch07 的 RAG 和 ch06 的 Memory 有什么区别？”，一个更可靠的计划不是直接 embedding 原问题，而是先检索 ch07 中 RAG 的核心结构和流程，再检索 ch06 中 Memory 的核心结构和流程，最后对比两者的数据来源、生命周期、注入方式和更新方式。

  从 ch07 当前实现看，流程更接近“用户说搜什么，我就搜什么”：`semantic_search` 接收 query，对 query 做 embedding，去 pgvector 搜索，可选 rerank，然后返回结果。这个系统已经有了 RAG 的执行链路，但还没有真正的 Query Rewrite / Query Planning 层。工业系统通常会先判断用户问题应该怎样被检索、拆解、补全和路由，再进入 embedding search、keyword search、grep、LSP、graph retrieval 等不同检索器。

  代码场景里这层尤其重要，因为自然语言和代码符号之间有天然鸿沟。用户可能说“检索工具在哪里执行的？”，但代码里可能叫 `SemanticSearchTool`、`Call`、`VectorStore.Search`、`Reranker.Rerank`。如果系统不会把自然语言问题转换成代码符号、文件名、接口名和包名，就只能靠 embedding 碰运气。

  所以更完整的查询链路通常不是“用户问题 -> embedding search”，而是“用户问题 -> 识别意图 -> 补全上下文 -> 生成检索计划 -> 生成多个 query -> 路由到不同检索器 -> 合并结果 -> 去重 -> rerank -> context packing -> 生成答案”。ch07 当前主要覆盖的是 embedding search、vector store、rerank 这一段，Query Rewrite / Planning 属于它前面的查询理解层。

  关键边界是：Query Rewrite 不是让模型直接回答问题，而是让模型生成更好的检索输入；Query Planning 也不是最终答案生成，而是决定为了回答这个问题需要查什么。它们都属于 retrieval preparation。如果这层做得好，后面的 embedding、rerank、context packing 都会更稳定；如果这层缺失，系统经常表现为用户问得稍微含糊一点，RAG 就召回不到关键内容。

### Q7. Retrieval Router 是什么？为什么不能所有问题都默认走 embedding search？

- **一句总结：**
  Retrieval Router 是 RAG 系统里的“检索路由器”：它负责判断当前问题应该走哪种检索方式，而不是默认所有问题都丢给 embedding search。

- **详细回答：**
  在简单 RAG 里，系统通常只有一条路：用户问题进入 embedding，然后做 vector search，再返回 top_k。但真实 Agent 系统里，用户问题类型差异很大，不同问题适合的检索方式不一样。

  比如用户问“这个函数在哪里定义的？”，这类问题更适合 LSP symbol search、grep 精确搜索或直接 read 文件，不一定适合 embedding。因为用户要找的是精确符号位置，不是语义相似段落。

  如果用户问“这段设计和之前哪一章的 memory 机制有关？”，这类问题更适合 semantic search、跨文档检索，可能还需要 rerank。因为它问的是语义关联，不是某个具体字符串。

  如果用户问“列出所有调用 `VectorStore.Search` 的地方”，这类问题更适合 grep、ripgrep 或 LSP references，因为这是确定性代码关系，不需要模型猜。

  Retrieval Router 解决的就是“用户问题来了以后，应该走哪条检索路径”。它可以路由到 embedding search、keyword/full-text search、grep/ripgrep、LSP search、vector + rerank、graph retrieval、direct read 等不同检索器。Router 的核心不是多接几个工具，而是决定什么时候用哪个工具，是否需要多个工具组合，以及结果如何合并。

  Query Planning 和 Retrieval Router 是相邻但不同的概念。Query Planning 更像是在问：为了回答这个问题，需要拆成哪些检索子任务？Retrieval Router 更像是在问：每个子任务应该交给哪个检索器？比如用户问“ch07 的 semantic_search 是怎么从用户 query 找到代码片段的？”，Planning 可能拆成找 tool 入口、找 embedding 调用、找 vector store search、找 rerank 逻辑；Router 会进一步决定 tool 入口用 grep 或文件搜索，embedding 调用用 grep `Embed(`，vector store search 用 grep `Search(`，rerank 逻辑用 grep `Rerank(`。

  Router 可以用规则实现。比如 query 包含“在哪里定义”“谁调用了”“引用”“函数名”，就优先 grep 或 LSP；包含明确文件路径，就 direct read；是“为什么”“有什么区别”“设计原因”，就更适合 semantic search；包含错误日志、配置键、命令输出，就更适合 keyword search 或 grep。规则路由透明、可控、容易 debug，但覆盖不完整。

  更进一步，可以用 LLM 做路由，让模型输出结构化计划，例如 intent、retrievers、queries、need_rerank。LLM 路由更灵活，但也更需要约束，因为模型可能选错工具、生成错误关键词，或者把简单问题规划得过于复杂。工业系统里通常不是纯规则，也不是纯 LLM，而是规则兜底、LLM 规划、检索结果反馈和可观测日志组合在一起。

  不能所有问题都默认走 embedding，是因为 embedding search 擅长语义相似，不擅长精确函数名、短字符串、错误码、配置键、路径、版本号、调用关系和字面量是否存在。比如你要找 `VectorStore.Search`，grep 基本是确定性的；embedding 可能找到“搜索向量库”的相关段落，但未必精确命中这个符号。没有 Router，系统会把所有问题都当成语义检索问题，导致本来可以确定性解决的问题变成概率问题。

  Router 的输出也不一定只有一个检索器。真实问题经常需要组合检索。比如“为什么 semantic_search 的 top_k 没生效？”，Router 可能同时选择 grep `top_k`、read `semantic_search.go`、semantic search “top_k rerank clipping”、grep `Rerank`，然后系统再把结果合并、去重、排序后交给模型。

  放到 ch07 当前实现里看，本章已经有 `semantic_search` tool、embedding、vector store 和 rerank，但还没有明确的 Retrieval Router。现在的结构更像“只要模型调用 semantic_search，就进入向量检索链路”。如果以后要加强，可以在 `semantic_search` 之前加一层：用户问题先经过 Query Planning，再经过 Retrieval Router，然后路由到 embedding、grep、read、LSP 或 graph retrieval，最后 merge、rerank、filter、context packing。

  对 Coding Agent 来说，这一层非常关键。代码问题天然混合了语义理解、精确字符串、文件路径、函数符号、调用关系和历史上下文。没有 Router，系统很容易什么都用 embedding；有了 Router，系统才可能根据问题类型选择更稳的检索方式。

### Q8. 对于增量更新与去重，`ch07` 是怎么解决的？

- **一句总结：**
  `ch07` 不做 chunk 级别 diff，而是按文件判断是否需要重建索引：如果文件没变就跳过，如果文件修改过就删除该文件旧的所有向量块，然后重新切分、embedding、写入。

- **详细回答：**
  `ch07` 的增量更新与去重主要靠“文档路径 + 文件修改时间 + 删除旧 chunk 后重建”来做。核心逻辑在 [indexer.go](/Users/bytedance/vibe-coding/baby-agent/ch07/index/indexer.go)。

```text
indexFile(filePath)
-> 转成相对路径 relPath，作为 documentID
-> os.Stat(filePath) 获取文件修改时间
-> vectorStore.GetDocumentIndexedTime(relPath) 查询该文档上次索引时间
-> 如果没索引过：新建索引
-> 如果索引过且文件修改时间 <= 索引时间：跳过
-> 如果索引过且文件修改时间 > 索引时间：删除旧索引并重新索引
```

  对应关键代码逻辑是：

```go
indexedTime, err := idx.vectorStore.GetDocumentIndexedTime(ctx, relPath)

if !indexedTime.IsZero() {
    if fileInfo.ModTime().Before(indexedTime) || fileInfo.ModTime().Equal(indexedTime) {
        return &FileIndexResult{
            FilePath: filePath,
            Chunks:   0,
            Action:   IndexActionSkip,
        }, nil
    }

    if err := idx.vectorStore.DeleteByDocument(ctx, relPath); err != nil {
        return nil, fmt.Errorf("failed to delete old index: %w", err)
    }
}
```

  - **它用 `relPath` 判断“同一个文档”。**
    `indexFile()` 会先把绝对路径转换成相对路径，并把这个相对路径作为 `DocumentID`。一个文件切出来的所有 chunk 都会带同一个 `DocumentID`。后续查询、删除、判断是否已经索引，都是围绕这个 `DocumentID` 做的。

  - **它用 `GetDocumentIndexedTime()` 判断文件之前是否索引过。**
    在 [pgvector.go](/Users/bytedance/vibe-coding/baby-agent/ch07/db/pgvector.go) 里，`GetDocumentIndexedTime()` 会查询这个 `document_id` 对应的最早一条 chunk 的 `CreatedAt`。如果查不到，说明该文件还没索引过；如果查到了，就把这个时间当成该文件的索引时间。

  - **如果文件没变，就直接跳过。**
    当文件的 `ModTime()` 早于或等于 `indexedTime`，说明当前文件内容在上次索引之后没有变过，因此不需要重新 embedding，也不需要重新写入向量库。这能避免重复索引未修改文件，节省 API 调用和数据库写入。

  - **如果文件变了，就先删除旧索引再重建。**
    一个文件会切成多个 chunk，每个 chunk 都会生成一个向量。如果文件变了，旧 chunk 的内容、行号、语义向量都可能失效。所以 `ch07` 不尝试更新旧 chunk，而是调用 `DeleteByDocument(documentID)` 删除这个文件对应的所有旧 chunk，然后重新读取文件、重新切分、重新 embedding、再 `InsertBatch()` 写入。

  这个方案的优点是实现简单，而且语义上比较稳。它不需要处理旧 chunk 和新 chunk 的对应关系，也不会出现同一个文件的新旧 chunk 混在一起被检索的问题。文件中间加几行后，后面所有 chunk 的行号和内容都可能变化，整文件重建可以避免旧 chunk 残留。

  但它的局限也很明显。

  - **粒度比较粗。**
    哪怕只改了一行，也会删除并重建整个文件的所有 chunk。

  - **依赖文件修改时间。**
    如果文件系统 mtime 不可靠，或者索引时间和文件修改时间存在精度、时钟问题，就可能误判。

  - **没有内容 hash。**
    如果文件内容没变但 mtime 变了，也会重新索引。更稳的做法是记录文件 hash。

  - **不是 chunk 级增量。**
    工业系统可能会计算 chunk hash，只重建变化的 chunk，保留未变 chunk。

  - **删除再插入不是事务化的。**
    当前逻辑里如果删除旧索引成功，但后续 embedding 或插入失败，这个文件会暂时没有索引。工业实现最好使用事务、版本号或 shadow index。

  更成熟的索引系统一般会引入这些字段：

```text
document_id
content_hash
chunk_hash
index_version
indexed_at
source_mtime
status: active / building / stale
```

  更工业化的流程可能是：

```text
扫描文件
-> 计算文件 hash
-> hash 未变：跳过
-> hash 变化：重新切 chunk
-> 对比 chunk hash
-> 只 embedding 新增/修改 chunk
-> 标记删除消失的 chunk
-> 原子切换新版本索引
```

  所以 `ch07` 当前实现是“文件级增量更新”，不是“chunk 级增量更新”。作为教学版足够清楚，能让你理解去重和索引新鲜度问题；但如果做工业产品，需要继续补内容 hash、chunk hash、事务/版本切换和失败恢复。

### Q9. 文件频繁更新会不会导致频繁切分和频繁更新向量数据库，从而让系统不稳定？

- **一句总结：**
  会，这是 RAG 索引系统的真实稳定性问题；`ch07` 当前按文件整块重建索引，适合教学，但工业系统通常要用 debounce、批处理、后台队列、hash、版本化索引和原子切换来稳定更新链路。

- **详细回答：**
  你的担心是对的。`ch07` 当前逻辑是文件级重建：

```text
发现文件修改
-> 删除该文件所有旧 chunk
-> 重新读取文件
-> 重新切分
-> 重新调用 embedding
-> 重新 InsertBatch 写入向量库
```

  这个方案很适合教学，因为逻辑清楚，但如果文件频繁变化，就会带来几个明显问题。

  - **成本问题。**
    embedding 通常是外部模型调用。一个文件只改一行，也要重建整个文件的所有 chunk。如果编辑器自动保存频率很高，每次都触发索引，API 成本会快速上升。

  - **性能问题。**
    切分、embedding、数据库删除和插入都不是零成本。大文件或大仓库里频繁更新，会导致索引任务堆积，影响搜索延迟和数据库吞吐。

  - **索引短暂缺失问题。**
    当前是先 `DeleteByDocument()`，再重新 `InsertBatch()`。如果删除成功后 embedding 失败或插入失败，这个文件在向量库里会暂时没有索引。用户此时搜索，会搜不到这个文件。

  - **新旧数据不一致问题。**
    如果多个索引任务并发处理同一个文件，旧任务可能晚于新任务完成，导致旧版本索引覆盖新版本。这就是 out-of-order write。

  - **检索结果抖动。**
    文件正在频繁保存时，向量库里的 chunk 可能不断删除、重建。用户同一个 query 在短时间内可能得到不同结果，Agent 行为也会跟着抖动。

  - **数据库写放大。**
    文件级重建会放大写入量。一个 100 个 chunk 的文件，只改 1 行也要删除 100 条再插入 100 条。频繁发生时会带来大量无效写入。

  更稳的工业系统通常会这样处理。

  - **Debounce。**
    文件变化后不立刻索引，而是等待一小段安静时间，比如 1 到 5 秒。如果期间继续变化，就重置计时器。这样可以避免编辑器连续保存触发多次重建。

```text
file changed
-> wait 3s
-> if changed again, reset timer
-> stable for 3s
-> index once
```

  - **批处理。**
    把一段时间内的多个文件变化合并成一个索引任务，统一处理，降低数据库和 embedding 服务压力。

  - **后台队列。**
    索引更新不应该阻塞主对话链路。文件变化进入队列，由后台 worker 慢慢处理。队列里还可以做去重：同一个文件多次变化，只保留最后一次。

  - **内容 hash。**
    不只看 mtime，而是计算文件内容 hash。mtime 变了但内容没变，就不重建索引。

  - **Chunk hash。**
    重切分后对每个 chunk 计算 hash，只对变化的 chunk 重新 embedding。没变的 chunk 直接复用旧向量。

  - **版本化索引。**
    不要先删旧索引再插新索引。可以先写入新版本，成功后再切换 active version。

```text
old version: active
new version: building
new version build success
-> switch active version to new
-> cleanup old version
```

  - **单文件串行化。**
    同一个 `documentID` 的索引任务需要串行执行，或者用版本号判断只允许最新任务提交，避免旧任务覆盖新任务。

  - **索引状态可观察。**
    系统应该知道哪些文件是 `active`、`building`、`stale`、`failed`。搜索时可以选择只搜索 active 版本，也可以提示“索引正在更新”。

  所以 `ch07` 当前实现适合理解 RAG 索引闭环，但不是稳定的实时索引系统。它缺少变更防抖、任务队列、内容 hash、chunk hash、版本化索引、原子切换、失败恢复和索引状态监控。

  这也是为什么很多 Coding Agent 不急着用 embedding 索引代码库，而是优先用 `grep / read`。文件系统永远是最新状态，不需要维护一套可能变脏、可能抖动、可能失败的索引系统。

### Q10. 索引新鲜度怎么保证？

- **一句总结：**
  索引新鲜度不是靠“每次文件变化立刻重建”保证，而是靠变更检测、去抖、队列、版本化索引、状态标记、失败恢复和搜索时的 stale 处理共同保证。

- **详细回答：**
  RAG 索引新鲜度指的是向量库里的内容是否和真实数据源一致。对 `ch07` 来说，真实数据源是文件系统里的代码和文档，向量库是 pgvector。

  如果文件已经改了，但向量库还是旧内容，就叫 stale index。如果文件删了，但向量库里还有旧 chunk，也是不新鲜。如果文件新增了，但还没索引，搜索也会漏结果。

  `ch07` 当前用的是最简单的文件级新鲜度判断：

```text
document_id = 文件相对路径
indexedTime = 向量库里该 document 最早 chunk 的 CreatedAt
file.ModTime = 文件系统修改时间

如果 file.ModTime <= indexedTime：跳过
如果 file.ModTime > indexedTime：删除旧索引并重建
如果 document 不存在：新建索引
```

  这能解决基本问题：未修改文件不重复索引，修改过的文件会重建，新增文件会索引。但它还不是完整工业方案。

  当前方案的问题包括：mtime 不完全可靠；文件删除没有处理；删除旧索引后重建失败会导致该文件暂时没有索引；并发更新可能乱序；没有 stale 状态；搜索时不区分版本。

  工业系统一般会分几层保证新鲜度。

  - **内容 hash，而不只看 mtime。**
    为每个文档记录 `document_id`、`content_hash`、`source_mtime`、`indexed_at`。扫描文件后计算 hash，hash 未变就跳过，hash 变化才进入重建。这样 mtime 变了但内容没变不会重建，mtime 没变但内容变了也能发现。

  - **文件 watcher + debounce。**
    实时系统通常会监听 file changed、file created、file deleted、file renamed。但不会立即重建，而是等待 1 到 5 秒安静期。如果继续变化，重置计时器；稳定后只索引一次。这样避免编辑器自动保存导致频繁重建。

  - **后台任务队列。**
    文件变化不直接执行索引，而是进入队列：`change event -> indexing queue -> worker`。队列可以做同一文件去重、合并多次变更、限流、失败重试、优先级调度。这样不会阻塞 Agent 主流程。

  - **版本化索引 / 原子切换。**
    不要先删旧索引再写新索引。更稳的是：

```text
old version: active
new version: building
build success
-> switch active_version to new
-> cleanup old version
```

    搜索只搜 `status = active`、`version = current`。这样即使新版本构建失败，旧版本仍然可用。

  - **document 状态机。**
    给每个 document 维护状态：

```text
active：当前可用
building：正在构建新版本
stale：源文件已变化，但新索引未完成
failed：重建失败
deleted：源文件已删除，等待清理
```

    搜索时可以只搜 active，允许搜 stale 但标注，stale 太久则提醒用户，failed 则回源文件系统。

  - **处理删除和重命名。**
    需要定期或实时对比文件系统中的 `document_id` 集合和向量库中的 `document_id` 集合。如果向量库里有，但文件系统没有，就标记 deleted，从 active 搜索中移除，并异步清理向量。重命名可以看成 old document deleted + new document created；更高级的系统可以用 content hash 判断是 rename，而不是重建。

  - **搜索时做 stale fallback。**
    如果用户问一个文件，而该文件索引状态是 stale 或 failed，不要只信向量库。更稳的做法是直接 `read` 当前文件，或者 `grep` 当前文件系统，或者提示索引正在更新。代码场景尤其需要这个。

  - **定期全量校验。**
    即使有 watcher，也要定期扫描全库，对比 `document_id` 和 `content_hash`，修复漏掉的变更。因为 watcher 可能丢事件，服务也可能重启。

  和 `ch07` 的关系是，当前已经有最小版：

```text
GetDocumentIndexedTime()
file.ModTime()
DeleteByDocument()
InsertBatch()
```

  它能表达“不要重复索引未修改文件”的基本思想。但如果要工业化，建议补 `content_hash`、`chunk_hash`、`document_status`、`index_version`、file watcher + debounce、indexing queue、atomic version switch、delete/rename cleanup、stale search fallback、periodic reconciliation。

  一句话理解，索引新鲜度不是“尽快更新”这么简单，而是要保证不漏更新、不过度更新、更新失败不影响旧索引、搜索时知道索引是否可信、源文件和向量库最终一致。`ch07` 当前是文件级 mtime 判断的教学实现，能说明基本思路；真正工业系统要围绕一致性、失败恢复和搜索时降级来设计。

### Q11. 为什么需要 rerank？Embedding 相似度不够吗？

- **一句总结：**
  Embedding 相似度适合快速召回可能相关的候选，但不够精细；rerank 是在候选集上重新判断 query 和文本的真实相关性，用更高成本换更高排序质量。

- **详细回答：**
  RAG 里通常把检索拆成两阶段：

```text
第一阶段：Embedding 向量检索
从海量文档里快速召回 Top-50 / Top-100 候选

第二阶段：Rerank
对这些候选重新排序，选出最适合进入上下文的 Top-5 / Top-10
```

  Embedding 相似度不是没用，它非常有用，但它解决的是“粗召回”问题，不是“最终排序”问题。

  为什么 embedding 相似度不够？因为 embedding 会把一段文本压缩成一个固定长度向量。这个向量要代表整段文本的语义，所以天然会损失细节。比如 query 是：

```text
Go 里如何避免 tool call 结果污染长期 memory？
```

  候选 chunk 有：

```text
A: memory update 应该区分用户事实、工具结果和模型推断。
B: tool call 的参数 schema 应该设计清楚。
C: Go 里 context.Context 用于取消请求。
D: 长期 memory 应该注入 system prompt。
```

  Embedding 可能觉得 A、B、D 都相关，因为都有 `tool`、`memory`、`Go`、`context` 相关概念。但真正最相关的是 A。向量检索可能能把 A 找回来，但排序未必稳定。

  原因主要有几个。

  - **向量是压缩表示。**
    一个 chunk 可能有很多主题，embedding 只能给它一个向量。它能表达大致语义，但很难表达 query 和 chunk 的细粒度匹配关系。

  - **相似不等于回答问题。**
    一个 chunk 和 query 主题相近，不代表它能回答 query。比如“memory 注入 system prompt”跟“memory 污染”相关，但不一定回答“tool result 怎么防止污染 memory”。

  - **关键词重合会干扰。**
    某些 chunk 只是包含相同术语，比如 `tool`、`memory`、`context`，但实际问题点不同。

  - **长 chunk 容易语义稀释。**
    chunk 里只要有一小段相关内容，整体向量可能看起来相关，但大部分内容是噪声。

  - **向量相似度没有真正逐句对齐 query。**
    它通常是 query vector 和 chunk vector 的一次距离计算，不会细致比较 query 中每个约束是否被 chunk 满足。

  Rerank 模型通常不是只比较两个向量，而是把 query 和 candidate 文本一起输入模型，让模型判断这个 candidate 对 query 的相关性。可以理解成：

```text
Embedding:
query -> vector
chunk -> vector
比较两个 vector 距离

Rerank:
(query, chunk) -> relevance score
```

  也就是说，rerank 会重新阅读用户问了什么、候选文本具体说了什么、候选文本是否真正能回答这个问题。它可以更细粒度地判断：这个 chunk 是主题相关，还是答案相关；这个 chunk 是否满足 query 的关键约束；这个 chunk 是否只是关键词相似。

  为什么不直接对所有文档 rerank？因为 rerank 更贵。如果有 100 万个 chunk，不可能把 query 和每个 chunk 都送进 rerank 模型。成本和延迟都会爆炸。所以才需要两阶段：embedding 便宜、快、可在百万级数据上召回候选；rerank 贵、慢、但更准，只处理几十个候选。

  `ch07` 当前的做法在 [semantic_search.go](/Users/bytedance/vibe-coding/baby-agent/ch07/tool/semantic_search.go) 里：

```text
query
-> EmbeddingService.Embed(query)
-> VectorStore.Search(queryVector, top_k * 2)
-> 提取 candidates
-> RerankService.Rerank(query, candidates)
-> 返回最终结果
```

  代码里特意召回 `top_k * 2`：

```go
vectorResults, err := s.vectorStore.Search(ctx, queryVector, params.TopK*2)
```

  意思是先多召回一些候选，让 rerank 有空间重新排序。否则如果只召回 `top_k`，rerank 只能在很小的集合里调整，提升有限。

  举个例子，用户问：

```text
为什么 memory update 只看 draft.NewMessages？
```

  向量召回可能得到：

```text
1. ch06 memory update 的流程
2. draft.NewMessages 的解释
3. system prompt 注入 memory 的原因
4. CommitTurn 的生命周期
5. tool result 为什么不展示给 TUI
6. memory gate 的设计
```

  向量阶段可能因为 “memory” 这个词，把第 3、第 6 也排得很高。Rerank 会重新看 query 和候选内容，可能把顺序调整成：

```text
1. draft.NewMessages 的解释
2. CommitTurn 的生命周期
3. ch06 memory update 的流程
4. memory gate 的设计
5. system prompt 注入 memory 的原因
6. tool result 为什么不展示给 TUI
```

  这样进入 prompt 的 top results 更可能真正回答问题。

  rerank 不是免费增强，它会增加一次模型调用、增加延迟、增加成本，rerank 模型也可能判断错。候选太长时 rerank 成本更高。还有一个很重要的限制：rerank 只能重排候选，不能重排没被召回的内容。如果正确 chunk 没进候选集，rerank 没机会把它排上来。所以召回阶段仍然很重要。

  更需要 rerank 的场景包括：文档很多、chunk 比较长、query 有多个约束、`top_k` 结果噪声多、用户问题需要精确答案、embedding 模型一般但 rerank 模型更强、希望减少进入 prompt 的无关内容。

  不一定需要 rerank 的场景包括：文档量很小、query 很简单、chunk 很短且结构清晰、embedding 模型效果已经很好、延迟或成本特别敏感、只是粗略探索不要求最终排序很准。

  一句话理解：embedding 像“快速扫一眼，找可能相关的人”；rerank 像“把这些人叫过来逐个面试，判断谁最能回答问题”。所以 embedding 相似度不是不够用，而是更适合做第一阶段召回；rerank 是为了把召回结果变成更可靠、更适合塞进 LLM 上下文的最终证据。

### Q12. `top_k` 怎么设计？为什么不是越多越好？

- **一句总结：**
  `top_k` 不是越多越好，它决定“给模型看多少检索结果”；太少容易漏信息，太多会引入噪声、挤占上下文、增加 rerank/LLM 成本，并让模型更容易被无关内容带偏。

- **详细回答：**
  `top_k` 表示检索时返回多少个结果。在 `ch07` 里，用户传：

```json
{
  "query": "memory update 为什么只看 draft.NewMessages",
  "top_k": 5
}
```

  语义是最终希望拿到 5 个最相关的 chunk。代码里实际会先多召回：

```go
vectorResults, err := s.vectorStore.Search(ctx, queryVector, params.TopK*2)
```

  也就是先向量召回 `top_k * 2` 个候选，再 rerank，最后截断成 `top_k` 个结果。这样做是为了给 rerank 留调整空间。

  如果 `top_k` 太小，比如只取 1 个，风险是正确 chunk 排在第 2 或第 3 时会被漏掉；问题需要多个证据片段时，单个 chunk 不够；多跳问题、跨文件问题很难回答；chunk 切得比较小时，单个 chunk 上下文不完整。

  比如用户问“`ch06` memory update 的完整流程是什么？”，可能需要同时召回 `context.Engine.CommitTurn`、`memory.MultiLevelMemory.Update`、`LLMMemoryUpdater.Update`、`memory.String` 注入 system prompt。如果 `top_k = 1`，模型只看到其中一段，很容易回答不完整。

  但 `top_k` 太大也会有问题。如果 `top_k = 50`，后面的结果相关性通常越来越弱。把弱相关 chunk 也塞进 prompt，会干扰模型判断。每个 chunk 都占 token，`top_k` 越大，留给对话历史、system prompt、工具结果和模型回答的空间越少。结果太多时，模型可能把不相关内容当成证据，或者在多个相似但冲突的 chunk 之间混淆。

  成本和延迟也会上升。如果有 rerank，候选越多，rerank 越贵。最终塞给 LLM 的文本越多，prompt token 成本也越高。给用户返回一堆结果，反而更难解释答案到底来自哪里。

  所以 `top_k` 的本质不是“越多越安全”，而是在召回覆盖和上下文噪声之间找平衡。

  `top_k` 不能单独调，它必须和 chunk size 一起看：

```text
小 chunk + 小 top_k：
容易漏上下文。

小 chunk + 大 top_k：
能补齐上下文，但模型需要拼碎片，rerank 更重要。

大 chunk + 小 top_k：
每个 chunk 信息多，但噪声也多。

大 chunk + 大 top_k：
最容易塞爆上下文，噪声最大。
```

  所以调 `top_k` 的时候，要同时看 chunk size、chunk overlap、rerank 是否启用、context window 大小，以及单次回答需要几个证据片段。

  不同问题类型适合不同 `top_k`。精确事实问题，比如“这个配置项在哪里定义”“某个函数是干什么的”，通常 `top_k = 3~5` 就够。流程类问题，比如“memory update 的完整流程是什么”“请求从 TUI 到 Agent 是怎么走的”，可能需要 `top_k = 5~10`。跨模块问题，比如“修改这个接口会影响哪些地方”“这个功能涉及哪些文件”，可能需要 `top_k = 10~20`，但最好配合 rerank、分轮检索或 `grep / LSP`。总结类问题不应该单纯把 `top_k` 拉很大，更适合分层检索、map-reduce summary、目录结构分析或图检索。

  Agentic RAG 里，`top_k` 可以由 Agent 动态决定。这也是 `ch07` 里 `top_k` 暴露给 tool schema 的原因：

```go
"top_k": {
    "type": "integer",
    "description": "返回结果数量，默认为5",
}
```

  Agent 可以根据问题复杂度选择简单问题 `top_k = 3`，普通问答 `top_k = 5`，复杂跨文件问题 `top_k = 10`。但模型未必总能选对，所以工业系统通常会设置默认值、最小值、最大值，超过上限自动截断，并结合问题类型给建议。比如先 `top_k = 5`，如果结果不足或置信度低，再 query rewrite、第二轮检索或扩大到 `top_k = 10`。这比一开始就 `top_k = 50` 更稳。

  更好的设计是把候选召回数和最终注入数分开。`ch07` 现在是向量召回 `top_k * 2`，最终返回 `top_k`。工业系统里通常会更明确地区分：

```text
candidate_k：向量召回多少候选，比如 50
rerank_top_k：rerank 后保留多少，比如 10
context_top_k：真正注入 prompt 多少，比如 5
```

  每一层目标不同：`candidate_k` 是尽量别漏正确答案，`rerank_top_k` 是保留高质量候选，`context_top_k` 是控制上下文噪声。

  `top_k` 还要配合阈值。如果用户问的问题在知识库里根本没有答案，`top_k = 5` 还是会硬返回 5 个“最像的”chunk。所以工业系统通常会加 `similarity_threshold`、`rerank_score_threshold`、`minimum_confidence`。如果结果低于阈值，就应该告诉模型没有找到足够相关的结果，而不是把弱相关内容硬塞给 LLM。

  评估 `top_k` 是否合适，也不应该靠感觉，而是看指标：`recall@k` 看正确 chunk 是否出现在前 k 个结果里，`precision@k` 看前 k 个结果里有多少真正相关，`MRR` 看第一个正确结果排多靠前，`NDCG` 看排序质量，`answer correctness` 看最终答案是否正确，`faithfulness` 看答案是否忠实于检索内容，同时还要看 latency 和 cost。

  如果增加 `top_k` 后，recall 提升明显，precision 没明显下降，答案正确率提升，token 成本可接受，那可以增大。如果增加 `top_k` 后噪声变多、答案更容易跑偏、成本上升但正确率没提升，那就不该增大。

  一句话理解：`top_k` 是 RAG 系统的“证据预算”。给得太少，模型证据不足；给得太多，模型被噪声淹没。所以 `top_k` 不是越多越好，而是要和 chunk size、rerank、阈值、问题类型、context window、成本一起调。

### Q13. Context Packing 是什么？为什么 RAG 不能直接把 top_k 原样塞进 prompt？

- **一句总结：**
  Context Packing 是 RAG 检索之后、模型生成之前的“上下文装箱”步骤：它决定哪些检索结果进入 prompt、按什么顺序进入、保留多少细节、如何标注来源，以及如何避免把上下文窗口塞满噪声。

- **详细回答：**
  RAG 不是把检索到的 top_k 结果原样拼进 prompt 就结束了。真正进入模型前，还需要回答几个问题：哪些结果值得放进去，每条结果放多少内容，按什么顺序放，重复内容怎么处理，来源信息怎么保留，多个文件和多个 chunk 之间怎么组织，如果上下文预算不够，先删谁。这一步就是 Context Packing。

  它处在用户问题、Query Rewrite / Planning、Retrieval Router、retrieval、rerank / filter 之后，在 LLM generation 之前。前面几步负责找到候选材料，Context Packing 负责把材料整理成模型能用的上下文。

  不能直接塞 top_k，是因为 top_k 只是检索系统认为相关的候选，不等于它们都适合进入 prompt。top_k 里可能有重复内容，比如同一个函数附近被切成多个相邻 chunk，向量搜索可能把它们都召回来；top_k 里也可能有弱相关结果，embedding 相似不等于对当前问题有帮助；top_k 的顺序也不一定适合模型阅读，检索排序是相关性排序，但模型阅读时可能更需要“入口文件 -> 核心实现 -> 辅助结构 -> 测试”的顺序。

  另外，每个 chunk 的边界可能不完整。一个 chunk 可能只包含函数中间部分，缺少函数名、文件路径、前后几行上下文。直接塞进去，模型可能误读。上下文窗口也有限，如果把很多长 chunk 全塞进去，反而会稀释关键信息，让模型忽略真正重要的证据。所以 Context Packing 不是简单拼接，而是一层上下文治理。

  Context Packing 通常会做过滤、去重、扩展上下文、排序、裁剪、标注来源、分组和冲突处理。过滤是删除低分、重复、过旧、不可信或权限不允许的结果；去重是合并相同文件、相邻 chunk 和重复段落；扩展上下文是对命中的 chunk 补充前后几行、函数签名、类名、文件路径；排序可以按相关性、文件结构、调用链顺序、时间顺序或阅读顺序组织内容；裁剪则根据 token budget 控制每条结果长度和总长度。

  代码 RAG 特别依赖好的 packing。比如检索到了某个函数体的一小段，直接塞进去可能不够。模型还需要知道文件路径、函数名、结构体或接口定义、调用方、被调用方、相关注释和测试用例。所以代码场景里常见策略是：命中 chunk 后，补充所在函数完整范围；如果 query 问调用关系，再补充 references；如果 query 问行为，再补充测试；最后按“定义 -> 实现 -> 调用 -> 测试”的顺序塞入 prompt。

  例如用户问“semantic_search 的 top_k 为什么没生效？”，好的 packing 可能不是放 10 个向量结果，而是组织成一个证据包：工具入口里 `semantic_search.go` 如何解析 `top_k`，向量召回时 `VectorStore.Search` 的 limit 是多少，rerank 后是否重新裁剪到 `top_k`，最后再给出当前代码中最终返回结果是否按 `top_k` 截断。这样模型拿到的是可推理证据包，而不是一堆相似文本。

  Context Packing 和 ch05 的上下文管理有关，但边界不同。ch05 管理的是历史消息，决定哪些历史对话保留、压缩、卸载、删除；ch07 的 Context Packing 管理的是检索结果，决定哪些证据进入本轮 prompt、怎么组织、怎么引用。两者都会消耗上下文窗口，所以实际系统里要一起考虑。如果历史对话已经占了很多 token，RAG packing 就必须更激进地裁剪；如果当前问题强依赖检索证据，系统可能要减少历史消息预算，给 RAG 结果更多空间。

  Context Packing 和 rerank 也不是一回事。rerank 解决的是候选结果谁更相关，Context Packing 解决的是相关结果怎么放进上下文，模型才更容易用。rerank 输出的是排序，packing 输出的是 prompt 里的证据结构。比如 rerank 认为 A、B、C 最相关，但 packing 可能决定 A 和 B 是同一个文件相邻 chunk，需要合并；C 缺少前文，需要补充前后 20 行；A 应该放在 C 前面，因为它是入口函数；每段前面还要加文件路径和行号。

  如果 Context Packing 做得不好，会出现信息太多、信息太少、顺序混乱、来源丢失、chunk 断裂、重复严重、冲突未标注等问题。这就是为什么 RAG 工程里常说 retrieval is not enough。检索到只是第一步，怎么把证据组织给模型同样重要。

  ch07 当前的 `semantic_search` 大致做了 query embedding、vector search、rerank、format results。它已经有一个非常基础的 packing：把搜索结果格式化成文本，附带 path、score、content。但它还不是完整的 Context Packing，缺少相邻 chunk 合并、按文件或模块分组、命中 chunk 的前后文扩展、根据 token budget 裁剪、来源引用结构化、冲突和版本标记、根据问题类型补充定义或测试、多检索器结果合并等能力。

  所以 ch07 可以理解为有 format results，但还没有真正的 context packing layer。对教学项目来说这是合理的，它先让你看到 RAG 主链路；后面如果要做工业级 Agent，就必须把 packing 单独抽出来。Context Packing 的意义是把检索结果变成一个模型更容易正确使用的证据包，给模型降低阅读成本，给答案提高可验证性，给上下文窗口做预算管理，也给 RAG 系统减少噪声污染。

### Q14. Citation / Grounding 是什么？为什么 RAG 答案需要可追溯到证据？

- **一句总结：**
  Citation / Grounding 是 RAG 系统里的“答案可验证性”机制：它要求模型的回答尽量绑定到检索证据，并把依据来源暴露给用户，而不是只给一个看似合理但无法追溯的结论。

- **详细回答：**
  RAG 的目标不是让模型看起来更懂，而是让模型基于外部证据回答。但只要最后一步还是 LLM 生成，就仍然存在一个问题：模型可能没有严格按照检索结果回答。它可能引用了检索结果里没有的内容，混合多个来源得出过度推断，把相似但不相关的 chunk 当成证据，遗漏关键限制条件，或者回答正确但用户无法验证。Citation / Grounding 解决的就是这个问题。

  Grounding 可以理解为“把回答锚定到证据上”。也就是说，模型回答时不能只依赖自己参数里的知识，也不能只凭语言流畅度生成，而应该尽量基于检索到的上下文。在 RAG 里，grounding 通常意味着答案中的关键事实应该能在检索结果中找到依据；如果检索结果没有证据，就应该说无法确认；如果不同证据冲突，要说明冲突；如果只是推断，要标注为推断，而不是事实。

  比如用户问“semantic_search 最后有没有按 top_k 截断？”，一个 grounded 的回答应该说明它依据的是 `ch07/tool/semantic_search.go` 中 `Execute` 的逻辑：先用 `top_k * 2` 做向量召回，再 rerank，最后如果 rerank 后结果超过 `top_k`，会截断到 `top_k`。这比只说“应该有截断”可靠，因为前者有证据，后者只是判断。

  Citation 是把 grounding 显式展示出来，也就是回答时告诉用户这个结论来自哪个文件、哪一行、哪个文档、哪个 chunk、哪个来源。Grounding 更强调回答是否基于证据，Citation 更强调证据是否可追溯。一个回答可以 grounded 但没有 citation，比如它确实基于证据回答了，但没告诉你证据在哪；一个回答也可能有 citation 但不 grounded，比如模型随便贴了一个看似相关的文件路径，但结论并不是从那里推出的。

  RAG 需要 Citation / Grounding，是因为 RAG 不是天然可靠。检索结果可能错，模型也可能误用检索结果。如果没有 citation，用户只能看到一个答案，却不知道它依据的是哪段内容、依据是否真的支持这个结论、有没有遗漏其他证据、是不是模型自己编的。在代码场景里尤其明显。如果模型说“这个参数最终没有被使用”，但不给文件路径和代码位置，用户很难验证。

  Citation / Grounding 不是只靠 prompt 里写一句“请引用来源”就能彻底解决。更稳的做法是系统层面配合。首先，检索结果必须带来源元数据，比如 `document_id`、`file_path`、`start_line`、`end_line`、`chunk_id`、`score`、`retriever`、`updated_at`。如果进入 prompt 的证据没有这些信息，模型就没有东西可引用。其次，Context Packing 时要保留引用标记，比如把每个证据块包装成 `[Source 1]`，并附带 path、lines 和 content。这样模型回答时才能说“根据 Source 1”。

  生成 prompt 也要明确约束，比如只基于 Sources 回答、没有证据时说无法确认、每个关键结论后标注 Source ID、不要引用没有出现在 Sources 中的内容。回答后还可以做校验，比如检查模型引用的 Source ID 是否真实存在，是否引用了不存在的文件，是否每个关键结论都有 citation。更进一步可以做 answer verification，让另一个模型或规则检查答案是否被证据支持。

  Context Packing 是 Citation / Grounding 的前置条件。如果 packing 阶段把来源信息丢了，后面就很难 citation。如果 packing 阶段把多个 chunk 混在一起，没有边界，模型也很难知道某句话来自哪里。所以可以理解成：Context Packing 负责把证据整理成可引用结构，Citation 负责把证据位置暴露给用户，Grounding 负责约束答案必须基于证据。

  Citation / Grounding 也会失败。常见失败包括伪引用、过度推断、来源粒度太粗、证据冲突未处理、引用覆盖不全、检索证据本身错误。比如模型引用了一个 Source ID，但结论并不是从这个 Source 推出来的；或者证据只说明 A，模型回答成 A 所以 B。citation 不是万能保证，它只是让错误更容易被发现。

  代码问答里，对 citation 的要求更高。最好引用到文件路径、函数名、行号、相关调用链和测试用例。因为代码结论通常需要精确验证。如果只引用整个文件，粒度还不够好；用户最好能看到具体代码位置。

  ch07 当前的 `semantic_search` 已经会把 path、score、content 格式化进结果，这说明它有 citation 的基础材料。但它还不是完整的 Citation / Grounding 系统。它缺少稳定 source id、行号范围、chunk id 暴露、引用格式约束、回答必须引用 source 的 prompt、引用合法性校验、答案是否被 source 支持的校验、冲突证据处理。所以当前可以说：ch07 有 source-aware result formatting，但还没有完整的 grounded generation。

  这对工业系统很重要。没有 Citation / Grounding，RAG 很容易变成检索只是给模型一点灵感，最终答案仍然不可验证。有了 Citation / Grounding，RAG 才更接近答案基于哪些证据、证据是否足够、用户能不能复查、系统能不能自动评估回答是否越界。尤其在代码、法律、医疗、企业知识库这些场景里，答案可追溯性比回答流畅度更重要。

### Q15. RAG 怎么让模型回答时引用来源？怎么防止模型引用不存在的内容？

- **一句总结：**
  让 RAG 回答可引用，不能只靠一句“请引用来源”的 prompt，而要从检索结果结构、Context Packing、生成约束、输出校验四层一起做，确保模型只能引用系统真实提供过的 source。

- **详细回答：**
  RAG 里“引用来源”有两个目标：用户能验证答案依据，系统能约束模型不要凭空编 source。这件事最容易做错的地方是只在 prompt 里写一句“请在回答中引用来源”。这不够，因为模型可能引用不存在的文件，引用没给过的 Source ID，把 Source 1 的内容归到 Source 2，引用一个相关但不能支持结论的来源，或者为了显得可信而伪造行号。所以引用来源要当成协议设计，而不是文案要求。

  第一步是在 Context Packing 阶段给每个进入 prompt 的证据块分配稳定 Source ID。例如把证据包装成 `[Source S1]`，里面包含 path、lines、chunk_id、score 和 content。关键是模型只能引用 `S1`、`S2` 这种系统真实提供过的 ID，不要让模型自己生成来源名。来源 ID 应该由系统生成，而不是模型生成。

  第二步是在生成 prompt 里明确约束引用规则。比如只基于 Sources 回答，每个关键结论后必须引用 Source ID，如果 Sources 中没有足够证据就说“当前资料无法确认”，不要引用没有提供的 Source ID，不要编造文件路径、行号或文档名，如果多个 Source 冲突就说明冲突。更严格时可以要求固定格式，比如每条结论后使用 `[S1]`，或者输出“结论 / 依据”。关键是让引用变成输出格式的一部分，而不是自由发挥。

  第三步是尽量使用结构化输出，降低伪造空间。比如让模型输出 `answer`、`claims`、`sources`、`unknowns`。这样系统可以检查 `sources` 里的每个 ID 是否真实存在、每条 claim 是否至少有一个 source、是否出现了未提供的 source。结构化输出不能保证模型一定不胡说，但它让校验变得可做。

  第四步是服务端校验 Source ID。生成之后，服务端至少要检查引用的 Source ID 是否在本轮 packed sources 里，有没有引用不存在的 source，有没有关键段落没有 citation，有没有输出文件路径但没有对应 source。如果发现模型引用了不存在的 Source ID，可以直接返回错误让模型重试，删除非法引用并标记答案不可靠，二次调用模型要求修正 citation，或者降级回答“当前答案引用校验失败，无法确认”。工业系统一般不会完全相信模型自己管理 citation。

  第五步是检查 citation 是否真的支持 claim。只检查 Source ID 存在还不够，因为模型可能引用了真实 Source，但这个 Source 并不支持它的结论。比如 Source 里只说 `Search(limit)`，模型却回答“最终结果一定严格等于 top_k”，这就是“合法引用 + 不支持结论”。更强的做法是做 grounding verification，把 claim 和 cited source 配对，判断 source 是否支持 claim。可以用规则做一部分，也可以用另一个模型做 verifier，输出 `supported`、`partially_supported`、`unsupported`、`contradicted`、`not_enough_information` 等结果。

  第六步是不要把 citation 粒度做得太粗。如果只给 `Source: README.md`，模型引用起来很容易，但用户难以验证。代码场景里更好的粒度是文件路径、行号范围和函数名，比如 `[S1] ch07/tool/semantic_search.go:42-108 SemanticSearchTool.Execute`。粒度太细也有问题，如果每三行一个 source，模型会被大量 source ID 淹没。一般要在可验证性和可读性之间平衡，让一个 source 对应一个完整函数、一个完整段落、一个配置块或一个文档小节。

  第七步是处理无证据和冲突证据。可引用系统必须允许模型说“当前资料无法确认”，否则模型为了满足用户，会硬编一个答案。如果 sources 之间冲突，也要要求模型说明冲突，比如 `S1` 表示 A，但 `S2` 表示 B，当前资料存在冲突，无法得出单一结论。这比强行合成一个看似完整的答案更可靠。

  第八步是在 UI 或答案结构上区分引用和推断。有些回答不是纯证据复述，而是基于证据做合理推断。这时应该区分事实、推断和无法确认。事实有明确 source 支持；推断是基于 source 的分析，但 source 没有直接说；无法确认则是 source 不足。这样用户能知道哪些是证据，哪些是模型推理。

  ch07 现在 `semantic_search` 返回里有 path、score、content。要进一步支持 citation，可以补几层：`VectorPointResult` 里保留 `chunk_id`、`path`、`start_line`、`end_line`；`semantic_search` 格式化结果时生成 Source ID，比如 `S1`、`S2`、`S3`；把 source id 和 path/line/content 一起返回给模型；Agent 的回答 prompt 里要求引用 Source ID；回答后检查 Source ID 是否属于本轮搜索结果。如果后面有 Context Packing 层，Source ID 最好在那里统一生成，而不是散落在每个 tool 里。

  核心原则是：引用来源不是让模型更礼貌地说明依据，而是系统协议。系统生成 source，模型只能引用 source，服务端校验 citation，必要时验证 claim 是否被 source 支持。这样才能防止模型引用不存在的内容，也能让用户真的复查答案依据。

### Q16. RAG 常见失败模式有哪些？怎么定位是 retrieval 问题还是 generation 问题？

- **一句总结：**
  RAG 失败通常不是一句“模型答错了”就能解释的，而要拆成 retrieval、ranking、packing、generation 四段定位：先看有没有召回正确证据，再看证据有没有排前面、有没有被正确放进 prompt，最后再看模型有没有基于证据回答。

- **详细回答：**
  RAG 的链路比普通 LLM 问答长，所以失败点也更多。一个典型链路是用户问题进入 query rewrite / planning，再经过 retrieval router，之后可能走 embedding、keyword、grep、LSP 或 graph retrieval，然后做 merge / dedup、rerank、context packing、generation，最后再做 citation / verification。只要其中任一环节出问题，最终答案都可能错。

  所以排查 RAG 问题时，不要先说“模型不行”，而是要问：正确证据有没有被召回，召回后有没有排在前面，排在前面的证据有没有进入 prompt，进入 prompt 的证据是否足够完整，模型有没有正确使用这些证据，答案有没有越过证据做推断。

  第一类常见失败是 query 理解失败。用户问得模糊、有指代、有上下文依赖，系统直接拿原始 query 去 embedding，导致检索方向偏了。比如“为什么这里没生效？”，如果没有把“这里”解析成具体文件、函数、参数或上一轮讨论对象，retrieval 很难命中正确材料。这类问题通常发生在 Query Rewrite / Planning 阶段。

  第二类是路由失败。问题本来应该走 grep、LSP 或 direct read，但系统走了 embedding search。比如 `VectorStore.Search` 在哪里被调用，这是精确符号检索问题。如果走 embedding，可能召回“向量搜索”相关段落，但漏掉真实调用点。这类问题发生在 Retrieval Router 阶段。

  第三类是召回失败。正确证据根本没有出现在候选结果里。原因可能是索引没建、索引过期、文件被过滤掉、chunk 切得不好、embedding 模型不适合当前语料、query embedding 表达不准、`top_k` 太小、向量索引近似搜索漏召回、权限过滤把结果过滤掉。这是典型 retrieval 问题。如果正确证据没有进入候选集，后面 rerank 和 generation 再强也救不回来。

  第四类是排序失败。正确证据被召回了，但排得很靠后，没有进入最终上下文。这可能是 embedding 分数不准，也可能是 rerank 模型不适合当前任务。比如 top 20 里有正确 chunk，但 top 5 没有；如果系统只把 top 5 放进 prompt，最终答案仍然会错。这类问题发生在 ranking / rerank 阶段。

  第五类是 chunk 边界失败。召回的 chunk 只包含局部内容，缺少前后文。比如只召回了函数中间几行，但没有函数名、参数定义、调用入口或错误处理分支。模型看到的是碎片，自然容易误解。这类问题发生在 chunking 和 context expansion 阶段。

  第六类是 context packing 失败。正确证据被召回，也排得不错，但进入 prompt 时被组织坏了。常见表现是重复 chunk 太多、关键证据被裁掉、来源信息丢失、顺序混乱、多个版本混在一起、代码片段缺少文件路径和行号、冲突证据没有标注。这类问题不是 retrieval 没找到，而是 evidence package 做得不好。

  第七类是 generation 失败。正确证据已经进入 prompt，但模型没有正确使用。它可能忽略证据、误读证据、把证据外的信息混进来、过度推断、回答得太笼统、没有承认证据不足、引用了不存在的来源。这才是更纯粹的 generation 问题。

  第八类是 citation / verification 失败。答案看起来有引用，但引用并不支持结论，或者引用粒度太粗，用户无法验证。比如模型说“根据 Source 2”，但 Source 2 只提到了相关函数，没有证明最终结论。这类问题说明 grounding 不够严格。

  区分 retrieval 问题和 generation 问题，最实用的方法是看正确证据在哪一步丢了。候选集里没有正确证据，就是 retrieval 问题；候选集里有，但 rerank 后掉到后面，是 ranking 问题；rerank 后有，但没有进入 prompt，是 context packing 或 token budget 问题；进入 prompt 了，但上下文断裂或来源混乱，是 packing 问题；完整证据已经进入 prompt，但模型仍然答错，才是 generation / grounding 问题。

  以“semantic_search 的 top_k 最终有没有生效？”为例，排查时不要直接看最终回答，而是逐层看。先看 retrieval candidates 里有没有召回 `ch07/tool/semantic_search.go`、`top_k` 解析逻辑、rerank 后裁剪逻辑；如果没有，说明 retrieval 或 query rewrite 有问题。再看 rerank 后这些正确 chunk 有没有排到前几；如果召回了但排很后，说明 rerank 或 score 有问题。再看最终 prompt 里有没有 `Execute` 函数相关代码、`top_k * 2`、最后截断到 `top_k` 的代码；如果没有，说明 packing 或 token budget 有问题。最后才看模型是否基于这些代码得出结论，是否混淆了向量召回数量和最终返回数量，是否引用了对应代码位置。

  一个可 debug 的 RAG 系统至少要记录原始用户 query、rewrite 后的 query、planning 结果、router 选择了哪些 retriever、每个 retriever 的 query、每个 retriever 的候选结果和分数、merge / dedup 后的结果、rerank 前后排序、最终进入 prompt 的 source 列表、每个 source 的 token 数、模型回答、模型引用的 source、verification 结果。如果没有这些日志，RAG 失败时只能猜。

  最关键的诊断原则是不要用最终答案倒推一切。正确做法是保留每个阶段的中间产物，然后判断证据是没找到，找到了但没排前，排前了但没放进 prompt，放进 prompt 但不完整，还是模型没按证据回答。这能把一个模糊的“RAG 效果不好”，拆成具体可修的问题。

  ch07 现在已经有基础链路：query embedding、vector search、rerank、format results，也加了一些日志，能看到 embedding、search、rerank、index 的过程。但如果要完整定位失败，还需要更系统的 tracing，比如 query rewrite / planning 结果、router 决策、最终 prompt 中包含哪些 source、每个 source 的 token 占用、模型回答和 source 的对应关系。所以 ch07 当前能 debug 一部分 retrieval 和 rerank 问题，但还不能完整 debug RAG 端到端质量。

### Q17. 一个 RAG 系统应该打哪些日志，才能定位召回和回答质量问题？

- **一句总结：**
  RAG 日志不能只记录最终答案，而要记录从“用户 query 被怎么理解”到“哪些证据进入 prompt”再到“模型如何引用证据”的完整链路，否则出错时无法判断是 query、retrieval、rerank、packing 还是 generation 的问题。

- **详细回答：**
  一个 RAG 系统最需要的不是大量日志，而是能还原每个关键决策点的 tracing。因为 RAG 失败时，最终表现通常只是答案不准、答案没引用、引用不支持结论、搜不到该搜的内容、模型忽略了检索结果。但这些现象背后的原因可能完全不同。所以日志设计要围绕一个核心问题：正确证据到底在哪一步丢了？

  请求入口日志要先记录用户原始请求和会话上下文摘要，比如 `request_id`、`user_id / workspace_id`、`conversation_id`、`turn_id`、原始 user query、当前会话摘要或关键上下文、时间戳、使用的模型和配置。这层日志用于复现问题。如果没有 `request_id / turn_id`，后面每个阶段的日志都串不起来。真实系统还要注意不要直接把敏感原文写入长期日志，需要做脱敏、权限控制和保留周期管理。

  Query Rewrite / Planning 日志要记录系统如何理解用户问题，比如是否触发 query rewrite、rewrite 后的 query、生成了哪些子问题、planning 输出的检索计划、为什么选择这个 plan。如果 rewrite 错了，后面的 retrieval 再强也会沿着错误方向搜。

  Retrieval Router 日志要记录系统为什么选择某些检索器，比如 router 输入、router 输出、选择了哪些 retriever、没有选择哪些 retriever、选择原因、fallback 策略。这层日志能判断问题是不是“该走 grep，却走了 embedding”。

  每个 retriever 的执行日志要分别记录。对 embedding search，要记录 embedding model、embedding dimensions、query embedding 是否成功、vector store、`top_k / candidate_k`、distance metric、filters、raw candidates、candidate scores、latency。对 keyword / grep，要记录 search keywords、搜索目录、include / exclude 规则、命中文件、命中行数、latency。对 LSP，要记录 symbol name、definition / references / implementation 查询类型、返回位置、latency。这一层最关键的是 raw candidates，也就是最早召回出来的候选结果。如果正确证据不在 raw candidates 里，就是召回问题。

  如果系统有多个 retriever，就一定要记录 merge / dedup 过程，比如每个 retriever 返回多少结果、合并前候选数、去重规则、被去掉的结果、合并后候选数。否则会出现一种很难排查的问题：某个 retriever 其实找到了正确证据，但 merge / dedup 时被删掉了。

  Rerank 日志要记录 rerank 前后顺序，比如 rerank model、输入候选数、rerank 前排名和分数、rerank 后排名和分数、被保留的 top_n、被丢弃的候选、latency。排查时重点看正确证据是否被召回，召回后排第几，rerank 后排第几，有没有被 top_n 截掉。如果正确证据 raw candidates 里有，但 rerank 后掉出最终集合，就是排序问题。

  Context Packing 日志是 RAG debug 里非常关键但经常缺失的一层。需要记录最终进入 prompt 的 source 列表、每个 source 的文件路径 / 文档 ID / 行号、每个 source 的 token 数、source 的来源 retriever、source 的原始分数和 rerank 分数、是否做了 chunk 合并、是否做了上下文扩展、是否做了裁剪、被裁掉的 source、总 token budget、RAG context token 占用、历史对话 token 占用、剩余 token budget。这层日志能回答正确证据是否真的进入了 prompt、进入 prompt 时是否完整、有没有因为 token 超预算被裁掉、有没有丢掉文件路径和行号。很多 RAG 看起来像 retrieval 失败，其实是 packing 阶段把证据裁掉了。

  Prompt / Message 日志要能追踪最终发给模型的请求结构，比如 system prompt 版本、developer / instruction prompt 版本、最终消息数量、每类消息 token 数、RAG evidence block 的位置、是否要求 citation、是否要求无法确认时拒答。真实系统不一定能长期保存完整 prompt 原文，尤其涉及隐私和安全，但至少要保存 prompt template 版本、source id 列表和 token 结构。这能判断是不是 prompt 没有要求模型基于证据回答，source 是否放得太靠后被模型忽略，历史上下文是否压过了 RAG 证据。

  Generation 日志要记录模型输出本身和使用情况，比如 model、temperature、max tokens、stream / non-stream、usage、finish_reason、最终答案、引用的 source id、latency、是否发生重试、错误类型。如果是流式，还要记录首 token 延迟、总 chunk 数、是否中途断流、断流时已经输出多少、是否完成 citation。这层日志能判断是不是模型生成阶段出了问题。

  如果系统有 Grounding / Verification 层，还要记录答案是否真的被证据支持，比如答案中的关键 claim、每个 claim 对应的 source、source 是否支持 claim、是否存在无引用 claim、是否存在伪引用、是否存在过度推断、verification 结果。这是判断“引用看起来有，但其实不支持结论”的关键。

  如果不想一开始做太复杂，最小可用日志集合至少应该包括 `request_id / turn_id`、`original_query`、`rewritten_query`、`router decision`、`retriever raw candidates + score`、`rerank before / after`、`final packed sources`、`token usage`、`final answer`、`cited sources`。这已经能定位大多数问题。

  ch07 目前已经加了一些基础日志，能看到 indexing、embedding request / response、vector search、rerank、`semantic_search` 执行过程。这对观察 RAG 主链路有帮助。但如果要定位召回和回答质量，还缺 query rewrite / planning 日志、router 决策日志、merge / dedup 日志、最终 packed sources 日志、prompt token budget 日志、answer citation 日志、grounding verification 日志。因为 ch07 当前还没有这些完整能力，所以日志也只能覆盖一部分链路。

  核心原则是：日志不是为了看起来很多，而是为了能回答用户问题被改写成了什么、系统为什么选择这些检索器、正确证据有没有被召回、召回后有没有被 rerank 排前、最终有没有进入 prompt、进入 prompt 时是否完整、模型有没有引用它、引用是否真的支持结论。只要日志能回答这些问题，RAG 的问题就能从“感觉效果不好”变成可定位、可复现、可修复的工程问题。

### Q18. RAG 怎么做评测？

- **一句总结：**
  RAG 评测要分开看两件事：检索有没有把正确材料找回来，生成回答有没有基于这些材料正确作答；不能只看最终回答“感觉还行”。

- **详细回答：**
  RAG 评测通常拆成两层：

```text
Retrieval Evaluation：检索质量
Generation Evaluation：回答质量
```

  因为 RAG 出错可能发生在不同位置：

```text
问题 -> 检索错了 -> 模型没看到正确材料 -> 回答错
问题 -> 检索对了 -> 模型误读材料 -> 回答错
问题 -> 检索太多噪声 -> 模型被带偏 -> 回答错
```

  第一层先评测检索。检索评测关心的是：正确 chunk 有没有被召回，正确 chunk 排得够不够靠前，`top_k` 里噪声多不多。

  常见指标包括：

```text
Recall@k：正确答案来源是否出现在前 k 个结果里。
Precision@k：前 k 个结果里有多少是真正相关的。
MRR：第一个正确结果排在第几位。
NDCG：综合考虑相关性和排序位置。
```

  比如有一个问题：

```text
Q: ch06 memory update 为什么放在 CommitTurn 之后？
```

  人工标注正确来源是：

```text
ch06/context/engine.go
ch06/LEARNING_QA.md 中 CommitTurn 相关段落
```

  如果检索 `top 5` 里出现了正确 chunk，`Recall@5` 就算命中。如果正确 chunk 排第 1，比排第 5 更好，`MRR / NDCG` 会体现出来。

  第二层再评测生成回答。生成评测关心的是：回答是否正确，回答是否基于检索材料，有没有编造，有没有漏掉关键点，有没有引用错误来源。

  常见指标包括：

```text
Answer Correctness：答案是否正确。
Faithfulness / Groundedness：答案是否忠实于检索内容，是否有无依据发挥。
Context Relevance：检索上下文是否和问题相关。
Answer Relevance：回答是否真正回答用户问题。
Citation Accuracy：引用的来源是否真的支持回答。
```

  比如检索结果里没有提到“后台模型”，但回答说“这里一定使用后台模型”，这就是不 grounded。

  RAG 评测的核心是评测集。至少要准备问题、标准答案或判定标准、正确来源文档或 chunk、问题类型和难度。例如：

```json
{
  "question": "ch07 为什么需要 rerank？",
  "expected_answer_points": [
    "embedding 是粗召回",
    "rerank 是精排",
    "rerank 只能重排已召回候选",
    "rerank 增加成本和延迟"
  ],
  "gold_sources": [
    "ch07/LEARNING_QA.md#Q9",
    "ch07/tool/semantic_search.go"
  ],
  "type": "conceptual",
  "difficulty": "medium"
}
```

  不要只准备简单问题。评测集应该覆盖精确事实问题、流程类问题、跨文件问题、多跳问题、总结类问题、不存在答案的问题、容易混淆的问题。其中“不存在答案的问题”很重要，否则系统会养成无论如何都硬答。

  如果最终回答错了，要能分阶段定位问题。

```text
检索没召回正确 chunk：
问题在 chunking / embedding / query rewrite / top_k / vector store。

召回了但排很后：
问题在 embedding 排序 / rerank / candidate_k。

召回了正确 chunk 但回答错：
问题在 prompt / 上下文组织 / 模型理解 / 证据引用。

召回了很多无关 chunk：
问题在 chunk size / top_k / threshold / metadata filter。
```

  这比只看“模型答错了”有用得多。

  RAG 的很多参数都要靠评测调，比如 chunk size、overlap、embedding model、candidate_k、top_k、rerank model、rerank_top_k、similarity threshold、metadata filter、query rewrite。可以比较不同方案：

```text
方案 A: chunk=500 tokens, top_k=5, no rerank
方案 B: chunk=1000 tokens, top_k=5, rerank
方案 C: chunk=500 tokens, candidate_k=50, rerank_top_k=5
```

  然后看它们的 `Recall@5`、`NDCG@5`、答案正确率、faithfulness、平均延迟和平均 token 成本。

  离线评测之外，线上还要看运行指标，比如检索为空比例、低分结果比例、用户追问或纠错率、回答引用来源点击率、平均检索延迟、embedding/rerank 成本、索引新鲜度、索引失败率。

  对 `ch07` 这种代码索引系统，还要关注多少文件已索引、多少文件 stale、最近一次索引时间、embedding 失败数、重建索引耗时、向量库 chunk 数量。

  `ch07` 当前实现了 RAG 链路，但没有真正的评测系统。它缺少评测问题集、gold sources 标注、自动跑检索的脚本、`Recall@k / MRR / NDCG` 计算、回答质量评测、不同参数组合对比和索引新鲜度评估。

  如果要补一个最小评测闭环，可以这样做：

```text
1. 准备 eval_cases.json
2. 每条 case 写 question 和 gold document/chunk
3. 调 semantic_search
4. 记录 top_k 结果
5. 判断 gold source 是否命中
6. 输出 Recall@k / MRR
```

  最小数据格式可以是：

```json
{
  "question": "为什么需要 rerank？",
  "gold_documents": ["ch07/LEARNING_QA.md"],
  "gold_keywords": ["粗召回", "精排", "候选"]
}
```

  最小评测指标可以先做 `Recall@5`、`MRR`、平均延迟，后面再加答案评测。

  一句话理解，RAG 评测不是问“模型答得好不好”，而是问：该找的材料找到了吗，找到的材料排前面了吗，模型有没有忠实使用这些材料，成本和延迟能接受吗。只有这几层都看，才知道是检索问题、排序问题、上下文问题，还是生成问题。

### Q19. 一般工业系统怎么判断 RAG 的召回阈值？

- **一句总结：**
  工业里 RAG 召回阈值通常不是拍脑袋设一个固定值，而是通过离线评测和线上反馈一起确定，在“召回足够多正确证据”和“过滤掉低质量噪声”之间找平衡。

- **详细回答：**
  工业系统里常见会同时看两类分数：`vector similarity score` 和 `rerank relevance score`。一般不会只靠向量相似度，因为 embedding 分数不一定稳定可比。更常见做法是先用较宽松的 similarity threshold 召回候选，再用 rerank score threshold 做最终过滤。

  典型流程是：

```text
准备评测集
-> 每个问题标注 gold source
-> 跑不同 threshold
-> 观察 recall / precision / answer quality / cost
-> 选一个业务可接受的平衡点
```

  例如测试这些阈值组合：

```text
similarity_threshold = 0.65 / 0.70 / 0.75 / 0.80
rerank_threshold = 0.3 / 0.5 / 0.7
```

  然后看这些指标：

```text
Recall@k：正确材料有没有被找回来
Precision@k：返回结果里噪声多不多
Answer correctness：最终答案是否正确
Faithfulness：答案是否忠实于材料
No-answer accuracy：没有答案时能否拒答
Latency / cost：延迟和成本是否可接受
```

  不能只看 recall。阈值太低时，召回更多，正确材料更可能进来，但噪声也更多，模型更容易被带偏，成本更高。阈值太高时，噪声少，但正确材料可能被过滤掉，模型看不到证据，就容易拒答或答不全。所以阈值本质上是 tradeoff。

  工业里常见策略有几类。

  - **分层阈值。**
    向量召回阈值宽松一点，rerank 阈值严格一点。

```text
candidate_k = 50
similarity_threshold = 0.65
rerank_top_k = 10
rerank_threshold = 0.55
context_top_k = 5
```

  - **动态阈值。**
    不同问题类型用不同阈值。精确事实问题阈值高，宁可少召回；探索或总结问题阈值低，允许更多材料；高风险问题阈值更高，并要求引用来源。

  - **相对阈值。**
    不只看绝对分数，也看 top1 和后续结果的差距。

```text
top1 = 0.86
top2 = 0.84
top3 = 0.83
=> 多个结果都可能相关

top1 = 0.86
top2 = 0.55
top3 = 0.52
=> 可能只有 top1 真相关
```

  - **无答案判断。**
    如果最高分也低于阈值，就不要强行回答。

```text
if top_score < threshold:
    return "没有找到足够相关的资料"
```

  - **按来源加权。**
    官方文档、当前代码、近期索引、用户指定范围，可以阈值稍低或权重更高；过期文档、低可信来源，阈值更高。

  初期可以采用一个务实流程：先不要设太严阈值，保证 recall；加 rerank，把候选重新排序；观察低分结果是否经常污染回答；再逐步提高阈值；最后用评测集固定参数。

  对于 `ch07`，更工业化的参数应该拆开：

```text
candidate_k
similarity_threshold
rerank_top_k
rerank_score_threshold
context_top_k
max_context_tokens
```

  而不是只靠一个 `top_k`。核心原则是：召回阶段宁可宽一点，注入阶段必须严一点。模型最终看到的材料要少而准。

### Q20. RAG 怎么避免把错误检索结果喂给模型？

- **一句总结：**
  RAG 不能只靠“检索 `top_k` 然后塞给模型”，而要在检索、过滤、重排、注入、生成和引用阶段都设防；核心原则是宁可少给，也不要把低置信、过时、冲突、无来源的内容当成证据喂给模型。

- **详细回答：**
  错误检索结果进入 prompt 后，模型很容易把它当成上下文证据。尤其 RAG prompt 通常会暗示“以下是相关资料”，这会让模型倾向于相信这些资料。所以防错不是一个点，而是一条链路。

  - **检索前先缩小搜索范围。**
    不要让系统在全库里乱搜。可以通过 metadata filter 缩小范围，比如只搜索当前 workspace、指定 repo、最近版本、官方文档、某种文件类型、用户指定目录。比如用户问 `ch07` 的 `semantic_search` 怎么工作，就不应该去召回 `ch03`、`ch06` 的内容，除非用户明确要跨章节比较。这一步能减少“主题相关但上下文错误”的结果。

  - **检索阶段设置相似度阈值。**
    不要硬返回 `top_k` 个结果。如果最高分都很低，说明知识库里可能没有答案，应该返回“没有找到足够相关的资料”，而不是强行塞 5 个“最像但其实不相关”的 chunk。

```text
if top_score < similarity_threshold:
    no relevant context
```

  - **召回后做 rerank 和 score threshold。**
    向量相似度只是粗召回。召回后应该用 rerank 做精排，并设置 rerank 阈值。

```text
vector search -> candidates
rerank(query, candidates)
filter by rerank_score
```

    这样可以过滤掉“词相似但不能回答问题”的 chunk。

  - **去重和聚类。**
    很多错误不是单条错，而是重复噪声太多。比如 10 个 chunk 都来自同一个无关文件，模型会误以为这个文件很重要。可以限制同一文档最多取 N 个 chunk，对相似 chunk 去重，按文件、模块、主题分组，优先保留覆盖面更好的结果。

  - **检查来源可信度和新鲜度。**
    不是所有来源同等可信。可以给来源打权重：当前代码 > 官方文档 > 项目 README > 历史聊天记录 > 旧文档 > 非权威网页。还要看当前索引是否 stale、文档是否过期、代码是否刚刚修改但索引未更新。过时来源不应该直接进入高优先级上下文。

  - **做冲突检测。**
    如果检索结果之间互相矛盾，不能直接全塞给模型。比如一个 chunk 说项目使用 Go 1.25，另一个 chunk 说项目使用 Go 1.20。系统应该标注冲突，或者优先选择更可信、更近、更权威的来源。必要时要回源验证。

  - **注入前控制上下文预算。**
    就算结果相关，也不一定都要喂给模型。可以分层注入：高置信结果原文注入，中置信结果摘要注入，低置信结果不注入或只作为候选。同时限制 `max_context_tokens`、`max_chunks_per_document`、`max_total_chunks`，避免低价值内容挤占关键证据。

  - **生成阶段约束模型只能基于证据回答。**
    Prompt 里要明确：只基于提供的检索结果回答；如果检索结果不足，说明无法确认；不要把不相关资料当作证据；回答中引用支持结论的来源；如果资料互相冲突，指出冲突。这不能完全保证，但能降低模型自由发挥。

  - **要求引用和可追溯。**
    输出时要求引用来源，例如“根据 `ch07/tool/semantic_search.go`”或“根据 `ch07/rag/type.go`”。引用不是为了好看，而是为了让系统和用户能检查这句话到底被哪个 chunk 支持，引用内容是否真的支持结论。如果答案没有任何来源支持，就应该降级或拒答。

  - **高风险问题要回源验证。**
    尤其代码场景不能只信索引。更稳的链路是检索命中某个 chunk 后，再 `read` 当前文件，或者用 `grep / LSP` 验证符号，然后再回答。因为索引可能 stale，chunk 可能过时，embedding 可能召回近似但错误的内容。这也是代码 Agent 更依赖 `grep / read` 的原因。

  - **必须允许无答案。**
    RAG 系统必须允许“不知道”。如果检索结果不够相关，应该返回“没有找到足够相关的上下文”，而不是让模型根据弱相关资料硬编。这类 no-answer case 应该进入评测集。

  `ch07` 当前已经有 vector search、rerank、`top_k`、来源位置 `DocumentID / StartPos / EndPos`，但还缺 similarity threshold、rerank score threshold、metadata filter、source ranking、stale index detection、conflict detection、dedup、`max_context_tokens`、citation enforcement、`read / grep` 回源验证和 no-answer policy。所以它现在是一个教学版 RAG 检索工具，还不是可靠证据系统。

  一句话理解：RAG 的目标不是“把最像的内容都塞给模型”，而是把足够相关、足够可信、足够新鲜、能支持答案的证据交给模型。错误检索结果的防护要贯穿检索前、检索后、注入前和生成后。

### Q21. RAG 和 Memory 的边界是什么？

- **一句总结：**
  RAG 是“从外部知识库按需召回资料”，Memory 是“从历史交互中沉淀长期状态”；二者都可能进入上下文，但来源、生命周期、写入方式、可信度和治理问题完全不同。

- **详细回答：**
  可以先用一句话区分：

```text
RAG：我现在需要查什么资料？
Memory：系统长期应该记住什么？
```

  它们都能让模型看到更多信息，但不是同一种东西。

  RAG 面向的是外部知识库，比如代码仓库、技术文档、API 文档、知识库、论文、产品说明、历史 issue、FAQ。它的典型流程是：

```text
用户问题
-> 生成 query
-> 检索外部知识库
-> 召回相关 chunk
-> 把 chunk 放进上下文
-> 模型基于 chunk 回答
```

  RAG 的特点是按需读取、围绕当前 query、知识来源通常是外部文档、结果可以很多、不一定每轮都需要、检索结果应该可引用。比如用户问 `ch07` 的 `semantic_search` 是怎么实现的，系统应该从代码和文档里检索 `ch07/tool/semantic_search.go`、`ch07/rag/type.go`、`ch07/db/pgvector.go`。这些是外部知识，不是用户长期偏好。

  Memory 面向的是系统从历史交互中学到的长期状态，比如用户偏好、用户工作习惯、项目长期约定、跨会话稳定事实、用户明确纠正过的信息、当前 workspace 的长期知识。它的典型流程是：

```text
一轮对话结束
-> 判断哪些信息值得长期保留
-> 更新 memory
-> 持久化
-> 未来会话作为背景注入
```

  Memory 的特点是跨会话存在，不一定来自外部文档，通常来自用户交互和系统观察，写入门槛应该更高，内容应该更稳定，并且会持续影响未来回答。比如用户说“后续学习笔记都按照 Q&A 形式记录”，这适合进入 memory，因为它是用户长期偏好。

  核心区别可以从几个维度看：

```text
来源：
RAG 来自外部知识库。
Memory 来自历史交互和长期状态沉淀。

生命周期：
RAG 是按需召回，当前 query 用完可以丢。
Memory 是跨会话持久化，会长期存在。

写入方式：
RAG 通常先离线建索引，文档更新后重建。
Memory 通常在对话后由系统判断是否写入。

读取方式：
RAG 根据当前问题检索相关片段。
Memory 通常作为长期背景注入，或按需召回。

可信度：
RAG 的可信度取决于来源文档和索引新鲜度。
Memory 的可信度取决于写入规则、用户纠正、来源标记。

风险：
RAG 的风险是召回错、过时、噪声污染。
Memory 的风险是误记、过度泛化、长期污染。
```

  不能把两者混在一起。如果把 RAG 当 Memory，会把大量外部文档当成长期偏好，memory 会迅速膨胀，不相关资料会长期污染 system prompt。比如把整份 README 记进 memory 就是错的。README 应该进入 RAG 索引，需要时检索，不应该作为用户长期记忆常驻。

  如果把 Memory 当 RAG，也会出问题。用户偏好每次都要检索，长期状态不稳定，系统可能忘记重要约定。比如用户已经明确说“学习笔记用 Q&A 风格”，这不应该每次靠搜索历史文档猜，而应该作为 memory 稳定生效。

  和 `ch06 / ch07` 的关系是：`ch06` 是 Memory，链路是：

```text
conversation messages
-> MemoryUpdater
-> Global / Workspace Memory
-> memory.String()
-> system prompt
```

  它解决的是跨会话之后系统应该记住什么。

  `ch07` 是 RAG，链路是：

```text
files
-> chunking
-> embedding
-> vector store
-> semantic_search tool
-> 检索结果进入上下文
```

  它解决的是当前问题需要查哪些外部资料。所以 `ch06 memory = 长期状态层`，`ch07 RAG = 外部知识召回层`。

  二者可以协作。比如用户问“继续按照我之前喜欢的学习笔记风格，总结 ch07 的 RAG”，系统需要 Memory 知道用户喜欢什么学习笔记风格，也需要 RAG 检索 `ch07` 的 README 和代码。也就是说，Memory 决定回答偏好和长期背景，RAG 提供当前任务需要的知识证据。

  二者也会互相影响。RAG 检索到的内容，可能经过判断后变成 Memory。例如 RAG 发现“这个项目长期使用 Go 1.25”“测试命令是 `go test -v ./...`”，如果这是稳定项目事实，可以沉淀到 Workspace Memory。但不是所有 RAG 内容都应该进 Memory。一次性检索到的某个函数实现，不应该自动变成长期记忆。

  所以需要 gate：

```text
RAG result -> 是否长期稳定？是否跨会话有价值？是否来源可信？ -> Memory
```

  更成熟的系统会把它们分层：

```text
Working Context：当前对话窗口
Memory：长期用户偏好、项目约定、历史稳定状态
RAG：外部文档、代码、知识库、工具检索结果
Storage：原始资料、索引、日志、审计数据
```

  一句话理解：RAG 像“查资料”，Memory 像“记习惯和长期事实”。查到的资料不一定要记住，记住的东西也不应该每次都重新查。二者都能进入上下文，但必须分清来源、生命周期和治理方式。

### Q22. RAG 应该作为 Agent 内部能力，还是作为 tool 暴露给模型？

- **一句总结：**
  RAG 可以作为 Agent 内部能力，也可以作为 tool 暴露给模型；区别在于“谁拥有检索决策权”：内部能力由系统自动检索，tool 模式由模型决定何时检索、检索什么、是否继续检索。

- **详细回答：**
  RAG 有两种常见接入方式。第一种是内部能力：用户问题进来后，系统自动检索，把检索结果塞进 prompt，然后模型回答。这种方式下，模型不需要知道“检索”是一个工具，它只看到系统已经准备好的上下文。

  第二种是 tool 模式：用户问题进来后，模型判断是否需要搜索，模型调用 search tool，系统返回检索结果，模型基于结果继续回答。这种方式下，RAG 是模型可调用的一种行动。

  这两种方式没有绝对优劣，关键看产品形态和任务复杂度。RAG 作为内部能力时，更像传统 RAG。系统在模型回答前自动执行检索，模型收到的是“用户问题 + 检索证据”。它的优点是稳定、可控、延迟更容易估算。每次都走固定检索链路，结果格式统一，系统容易做 citation / grounding，也不依赖模型主动决定是否搜索。它适合企业知识库问答、客服知识库、产品文档问答这类问答型产品。

  内部 RAG 的缺点是灵活性不足。即使不需要检索，也可能检索；复杂任务里不能让模型边读边搜；模型无法根据第一轮结果决定下一步查什么；多跳问题需要系统提前规划好。所以它更适合问题主要是单轮知识库问答、数据源固定、检索策略稳定、用户期望每次都基于资料回答、需要严格 citation / grounding、延迟和成本要可控、模型不需要多步探索的场景。

  RAG 作为 tool 时，更接近 Agentic RAG。模型可以决定是否需要检索、检索什么 query、调用几次、是否换关键词再搜、是否结合其他工具、什么时候停止检索并回答。这适合开放式任务、代码分析、多步骤研究任务。比如用户问“为什么 ch07 的 semantic_search top_k 没生效？”，模型可能先调用 `semantic_search` 搜 `top_k`，发现结果不够，再调用 grep 搜 `top_k`，再 read 某个文件，最后回答。

  tool 模式的优点是灵活，能做多轮检索和探索。缺点是更难治理。模型可能不该搜时搜，该搜时不搜，query 写得不好，重复搜索，搜索成本不可控，工具结果污染上下文，更难保证 citation，也更难做延迟预算。所以 tool 模式需要更强的 tool schema、system prompt、调用预算、日志、失败处理和权限控制。

  真实系统里经常混合使用。一种常见设计是内部 RAG 做默认背景召回，search tool 允许模型主动补查。系统先自动检索和用户问题最相关的文档，把基础证据放进 prompt；如果模型发现证据不足，再调用 search tool 做更有针对性的检索。另一种设计是 router 决定是否自动检索，同时模型也可以显式调用检索 tool。也就是说，系统负责默认稳态，模型负责复杂探索。

  ch07 更偏 tool 模式。`semantic_search` 被封装成一个 tool，模型可以决定是否调用它、传什么 query、传多少 `top_k`。这就是本章强调的 Agentic RAG：检索不是固定 pipeline，而是 Agent 可以选择的一种动作。但 ch07 当前还比较教学化，它还没有检索调用预算、多检索器 router、自动背景 RAG、检索结果 citation 校验、tool 调用失败恢复、多轮检索规划。所以可以说 ch07 展示了 RAG-as-tool 的雏形，但还不是完整工业 Agentic RAG。

  选择内部 RAG 还是 tool 模式，本质上是在回答：检索决策权应该交给系统，还是交给模型？如果你希望每次回答都稳定基于资料，系统应该掌控检索；如果你希望 Agent 能像人一样边查边判断，模型应该拥有部分检索决策权。但模型拥有决策权以后，系统必须补治理能力，比如最多查几次、每次查多少、哪些工具可用、是否允许访问某些数据、如何记录检索轨迹、如何验证最终答案。

  所以不是 tool 更高级，而是 tool 更灵活，也更难控。一个实用判断是：检索是回答前的固定准备动作，就更适合内部 RAG；检索是任务执行过程中的可选行动，就更适合 RAG tool。比如“根据公司报销制度，差旅补贴是多少？”更适合内部 RAG；“帮我查一下这个项目里为什么 semantic_search 的结果不对”更适合 tool 模式，因为模型需要边搜、边读、边判断。

### Q23. 为什么代码场景里 `grep / read / LSP` 经常优于 embedding？

- **一句总结：**
  代码场景里很多问题需要“精确事实”，而不是“语义相似”；`grep / read / LSP` 直接基于当前文件和语言结构，结果新鲜、可验证、确定性强，所以经常比 embedding 更适合作为 Coding Agent 的主检索手段。

- **详细回答：**
  Embedding 的强项是语义相似，比如“删除用户”“移除账号”“remove account”可能语义接近。但代码场景里，大量检索目标不是“意思相近”，而是“精确命中”。

  比如你要找 `deleteUser`，`grep` 可以直接找到函数定义、所有调用点、测试里怎么用、注释里哪里提到。而 embedding 可能会召回 `removeAccount`、`deleteOrder`、`createUser`、`userRepository`。这些在语义上可能相关，但不是你要的精确符号。所以第一点是：代码标识符本身就是高精度语义锚点。

  代码里的语义通常绑定在符号上，比如函数名、变量名、类型名、接口名、配置 key、错误码、路由路径、数据库字段、package/module 名。这些东西不需要近似语义搜索，而需要精确查找。例如 `OPENAI_API_KEY`、`CommitTurn`、`semantic_search`、`DeleteByDocument`，你想找的通常就是这个名字本身。Embedding 把它变成向量，反而可能损失精确性。

  另一个关键点是文件系统永远比索引新鲜。代码会频繁变化，向量索引是某个时间点的快照。如果你刚改了 `CommitTurn(...)`，但索引还没更新，embedding 检索可能查到旧内容。`read` 直接读当前文件，`grep` 直接搜当前文件系统，天然最新。对 Coding Agent 来说，过时信息非常危险，慢一点但正确通常比快但可能旧更好。

  `LSP / AST` 还能理解代码结构。它们可以回答这个函数定义在哪里、这个接口有哪些实现、这个变量有哪些引用、这个方法属于哪个类型、这个 symbol rename 会影响哪些文件。这些是 embedding 不擅长保证正确性的。Embedding 只能说“这段文本和 query 相似”，不能保证这是定义不是调用，这是当前作用域里的变量不是同名变量，这是实际实现不是注释里提到，这是编译器认可的引用关系。

  `grep / read` 也更可解释、可验证。`grep` 的结果是文件路径、行号、匹配文本；`read` 的结果是当前文件真实内容。Agent 可以继续 `grep -> read -> 再 grep -> 再 read`，每一步都可追踪。Embedding 检索结果则更难解释：为什么这个 chunk 相似度是 0.78，为什么另一个是 0.74，为什么真正相关的没进 `top_k`，这些都更难调试。

  Coding Agent 的搜索通常是迭代式的，不是一次检索就结束。例如：

```text
grep "CommitTurn"
-> read context/engine.go
-> grep "CommitTurn(" 看调用点
-> read agent.go
-> grep "Memory.Update"
-> read memory/memory.go
```

  这是推理驱动的搜索路径。Agent 根据每一步结果决定下一步。普通 RAG 更像 `query -> top_k chunks -> answer` 的一次性召回。代码任务经常需要顺着真实依赖链走，`grep / read / LSP` 更适合这种交互式探索。

  要让 embedding 在代码场景稳定，还要维护索引，包括 file watcher、debounce、增量索引、chunk hash、版本切换、索引状态、失败恢复、stale 检测。否则索引容易过时、抖动、缺失。但 `grep / read` 不需要这些。文件是什么，它搜到的就是什么。

  这不是说 embedding 在代码场景没用。它适合 README、设计文档、注释、Changelog、issue、自然语言需求，也适合“哪个模块大概负责 X”“哪里实现了类似 Y 的功能”“跨语言语义搜索”这类问题。当你不知道函数名，只知道意图，比如“哪里处理用户登录后的 token 刷新”“哪个模块负责把订单状态同步到外部系统”，embedding 可以帮你粗筛候选，再用 `grep / read / LSP` 验证。

  实际更稳的 Coding Agent 检索形态是混合的：

```text
embedding：找候选区域 / 语义粗筛
grep：精确找符号和关键词
read：查看真实源码
LSP/AST：验证定义、引用、类型、调用关系
LLM：解释和规划下一步
```

  也就是说，`Embedding = 发现线索`，`grep / read / LSP = 验证事实`。不要反过来让 embedding 当最终事实来源。

  和 `ch07` 的关系是，`ch07` 目前做的是：

```text
file -> chunk -> embedding -> pgvector -> semantic_search
```

  这对学习 RAG 很好，但如果要做工业级 Coding Agent，不能只靠它。应该让 Agent 有多种工具：`semantic_search` 找语义相关候选，`grep` 找精确符号或文本，`read` 读当前文件，`LSP` 找定义、引用、类型关系。对于代码回答，最终最好回源验证：

```text
semantic_search 命中候选
-> read 当前文件确认
-> grep/LSP 查引用
-> 再回答
```

  一句话理解：代码不是普通自然语言文档。自然语言问答需要语义相似，代码修改需要当前、精确、可验证。所以 `grep / read / LSP` 经常比 embedding 更适合作为代码 Agent 的事实基础。

### Q24. 既然代码场景里 `grep / read / LSP` 经常优于 embedding，怎么从用户 query 转成要 grep 的关键词？

- **一句总结：**
  从用户 query 到 grep 关键词，不是简单把整句话拿去搜，而是让 Agent 抽取“可能出现在代码里的稳定符号”：函数名、类型名、配置 key、错误码、API 路径、领域名词，再通过多轮搜索不断修正关键词。

- **详细回答：**
  用户 query 往往是自然语言，比如：

```text
记忆更新什么时候触发？
```

  但代码里不一定写着“记忆更新什么时候触发”。代码里可能是 `CommitTurn`、`Memory.Update`、`LLMMemoryUpdater`、`SetMemoryEventHook`。所以中间需要一个 query understanding / query planning 步骤：

```text
用户自然语言问题
-> 提取可能的代码概念
-> 生成 grep 关键词
-> grep/read 验证
-> 根据结果继续扩展或收缩关键词
```

  - **优先抽取精确符号。**
    如果用户 query 里已经有代码符号，直接搜。比如用户问 `CommitTurn` 是在哪里调用的，就搜 `CommitTurn`；用户问 `OPENAI_API_KEY` 是在哪里读取的，就搜 `OPENAI_API_KEY`；用户问 `semantic_search tool` 怎么执行，可以搜 `semantic_search`、`SemanticSearch`、`SemanticSearchTool`。这类最稳，因为 query 里已经有代码中的字面锚点。

  - **从自然语言映射到代码命名风格。**
    如果用户说的是中文或自然语言，要生成可能的英文/代码表达。比如“记忆更新什么时候触发？”可以生成 `memory`、`update`、`Memory.Update`、`MemoryUpdater`、`CommitTurn`、`memory event`，再考虑 Go 命名风格，如 `UpdateMemory`、`MemoryUpdate`、`SetMemoryEventHook`。这一步本质上是让模型根据代码命名习惯做候选生成，但这些候选还不是事实，需要 grep 验证。

  - **按领域名词拆 query。**
    用户 query 通常包含几个概念，比如“记忆 / 更新 / 触发”。可以先映射成英文候选：`memory`、`update`、`trigger`、`commit`、`event`、`hook`。然后先搜最核心概念：

```bash
rg "Memory|memory" ch06
```

    再根据结果扩展：

```bash
rg "Update" ch06/memory ch06/context
rg "CommitTurn" ch06
```

    关键是不要一开始就搜整句。

  - **用文件结构缩小范围。**
    不要全仓库乱搜。先判断章节和模块：用户问 memory 就优先搜 `ch06/memory`、`ch06/context`；用户问 RAG 就优先搜 `ch07/rag`、`ch07/index`、`ch07/tool`；用户问 TUI 就优先搜对应章节的 `tui` 目录。范围越小，噪声越少。

  - **搜不到时扩展关键词。**
    如果直接搜不到，就换同义词、大小写、命名风格。例如用户问“长期记忆写入在哪里？”，可以先搜：

```bash
rg "long.?term|memory" ch06
```

    如果不够，再搜：

```bash
rg "Store|Save|Persist|Write" ch06
rg "Update\(" ch06
```

  - **搜到太多时收缩关键词。**
    如果 `rg "Update" ch06` 结果太多，就加限定：

```bash
rg "Memory.*Update|Update.*Memory" ch06
rg "func .*Update" ch06/memory
rg "CommitTurn|Memory.Update" ch06/context ch06/memory
```

    或者先用文件名和目录过滤，例如只搜 `ch06/memory`。

  - **关键词生成可以分层。**
    可以让 Agent 生成几类搜索词：精确符号如 `CommitTurn`、`MemoryUpdater`、`SetMemoryEventHook`；领域词如 `memory`、`update`、`commit`、`event`；行为词如 `store`、`persist`、`save`、`load`、`trigger`；目录猜测如 `ch06/memory`、`ch06/context`、`ch06/tui`。搜索顺序通常是先搜精确符号，再搜领域词，再搜行为词，最后扩大目录。

  grep 不是一次性的，而是迭代的。真实流程通常是：

```text
用户 query
-> 猜关键词 memory/update
-> rg "memory|Memory" ch06
-> 发现 memory.Update / MemoryUpdater
-> rg "Memory.Update|MemoryUpdater" ch06
-> 发现 CommitTurn 调用
-> rg "CommitTurn" ch06
-> read engine.go
-> 得出答案
```

  这就是 Coding Agent 的优势：它不是 query 一次就结束，而是根据搜索结果继续调整计划。

  LSP 可以补齐 grep 的弱点。grep 是文本匹配，不懂语言结构。找到符号后，LSP 更适合回答定义在哪里、引用有哪些、实现有哪些、类型是什么。更稳的流程是：

```text
grep 找到候选符号
-> LSP 查定义/引用
-> read 关键文件
```

  embedding 在这里也可以辅助。如果完全不知道关键词，可以先用 embedding 找候选文件或模块。比如用户问“系统什么时候把长期状态写进上下文？”，你不知道关键词是 `memory.String()` 还是 `BuildSystemPrompt()`，可以先 `semantic_search("长期状态 写进 上下文")` 找到 `ch06/context/engine.go`，再 `read engine.go`，再 grep `memory.String|BuildSystemPrompt`。

  所以一句话理解，从 query 到 grep 关键词，本质上是 Agent 做一个小型检索计划：

```text
自然语言意图
-> 可能的代码符号
-> 可能的英文命名
-> 可能的目录范围
-> grep 验证
-> 根据结果迭代
```

  不能把用户原句直接 grep，而要把它翻译成代码世界里可能真实出现的符号和词。

### Q25. Hybrid Search 是什么？为什么语义搜索要和全文搜索结合？

- **一句总结：**
  Hybrid Search 是把语义搜索和全文搜索结合起来：语义搜索负责找“意思相近”的内容，全文搜索负责找“字面精确命中”的内容，两者融合后比单独使用 embedding 更稳，尤其适合代码、错误码、API 名、配置项这类精确符号很多的场景。

- **详细回答：**
  RAG 里常见两类搜索：语义搜索和全文搜索。语义搜索通常是 embedding / vector search，全文搜索通常是 keyword / BM25 / inverted index。它们解决的问题不同。

  语义搜索把 query 和文档都转成向量，然后按向量相似度找结果。它擅长同义表达、自然语言问题、跨语言表达、概念相近，以及用户不知道精确关键词的场景。比如用户问“怎么删除用户？”，语义搜索可能召回“移除账号”“delete account”“用户注销流程”。即使字面没有“删除用户”，也可能找到。

  全文搜索基于倒排索引、关键词、BM25 这类算法。它擅长精确词命中、函数名、变量名、错误码、API 路径、配置 key、日志关键字、专有名词。比如 `OPENAI_API_KEY`、`CommitTurn`、`DELETE /users/:id`、`ERR_CONNECTION_RESET`，这种情况下全文搜索通常比 embedding 更可靠，因为你要找的就是这个字符串本身。

  单独语义搜索的问题是，embedding 会把文本压缩成语义向量，适合相似语义，但会损失精确符号信息。例如你搜索 `CommitTurn`，embedding 可能召回 `StartTurn`、`AbortTurn`、`BuildRequestMessages`、`context lifecycle`。这些概念相关，但你真正想找的是精确的 `CommitTurn` 定义和调用点。代码场景里这种问题非常多。

  单独全文搜索的问题是，它依赖字面匹配。如果用户不知道准确词，就容易搜不到。比如用户问“记忆是什么时候写入长期存储的？”，代码里可能写的是 `CommitTurn()`、`memory.Update()`、`Storage.Store()`。全文搜索如果直接搜“长期存储”，可能找不到正确代码；语义搜索可以通过含义找到相关 chunk。

  Hybrid Search 的核心思想是两者结合：

```text
query
-> 语义搜索召回一批结果
-> 全文搜索召回一批结果
-> 合并去重
-> 融合排序
-> rerank
-> 返回最终 top_k
```

  也可以写成：

```text
semantic_results = vector_search(query)
keyword_results = full_text_search(query)
merged = merge(semantic_results, keyword_results)
reranked = rerank(query, merged)
```

  融合排序常见有几种方式。

  - **加权分数。**
    直接把语义分数和关键词分数加权：

```text
final_score = alpha * semantic_score + (1 - alpha) * keyword_score
```

    如果 `alpha = 0.7`，表示更相信语义搜索。如果是代码符号问题，可以降低 `alpha`，例如 `alpha = 0.3`，表示更相信全文搜索。

  - **RRF（Reciprocal Rank Fusion）。**
    不直接比较原始分数，而是看排名：

```text
score = 1 / (k + rank_semantic) + 1 / (k + rank_keyword)
```

    RRF 很常用，因为不同搜索系统的分数尺度不一样，向量相似度和 BM25 分数不一定可直接相加。

  - **先召回后 rerank。**
    最稳的做法通常是：

```text
semantic top 50
keyword top 50
merge 去重
rerank top 100
最终取 top 5
```

    这样 rerank 负责最终相关性判断。

  Hybrid Search 在代码库搜索、错误排查、API / 配置检索、用户描述模糊但代码是精确符号的场景中特别有用。比如 `CommitTurn 在哪里调用？`，全文搜索能精确命中 `CommitTurn`，语义搜索能补充 `turn lifecycle`、`commit draft` 相关内容。再比如 `ERR_CONNECTION_RESET`，全文搜索能精确找错误码，语义搜索能找相关解释。

  和 `ch07` 的关系是，[tool.go](/Users/bytedance/vibe-coding/baby-agent/ch07/tool/tool.go) 里已经预留了：

```go
AgentToolSemanticSearch AgentTool = "semantic_search"
AgentToolFullTextSearch AgentTool = "full_text_search"
AgentToolHybridSearch   AgentTool = "hybrid_search"
```

  但当前只实现了 `semantic_search`。也就是说，本章已经在设计上暗示语义搜索不是全部，未来可以补全文搜索和混合搜索。如果实现 Hybrid Search，可能会新增 `FullTextSearchTool`、`HybridSearchTool`、`FullTextStore / BM25 index`、`mergeResults()`，以及对合并结果再 rerank 的逻辑。

  Hybrid Search 更稳，是因为它降低了单一检索方式的盲区：语义搜索容易漏掉精确符号，全文搜索容易漏掉同义表达。Hybrid Search 同时覆盖两边，后面再通过 rerank、阈值、回源验证来控噪声。

  一句话理解：语义搜索像“理解你想找什么”，全文搜索像“精确找到你写了什么”。Hybrid Search 把这两种能力合在一起，所以比单独 embedding 更适合复杂真实系统，尤其是代码和技术文档场景。

### Q26. Graph RAG 和普通 RAG 有什么关系？

- **一句总结：**
  Graph RAG 可以理解为普通 RAG 的增强形态：普通 RAG 主要按“文本片段相似度”找资料，Graph RAG 额外把实体、关系、路径、社区这些结构化知识建成图，用图结构帮助检索和推理。

- **详细回答：**
  普通 RAG 解决的是“从一堆文本里找相关片段”，Graph RAG 解决的是“从一堆文本背后的实体关系网络里找相关知识”。它不是替代普通 RAG，而是在普通 RAG 检索之外增加一层知识图谱式的结构化召回和关系推理。

  普通 RAG 的核心链路通常是：

```text
文档
-> 切分 chunk
-> embedding
-> 向量库
-> query embedding
-> 相似度搜索
-> top-k chunks
-> LLM 基于 chunks 回答
```

  它的检索单位主要是文本 chunk。优点是简单、通用、容易落地；缺点是 chunk 之间的关系弱，容易只找局部相似片段，对跨文档、多跳关系、全局总结不够强。检索结果通常是平铺的 top-k 文本，系统并不知道实体之间怎么连起来。

  Graph RAG 会在普通文档处理之外，多做一层结构抽取：

```text
文档
-> 抽取实体 Entity
-> 抽取关系 Relation
-> 构建图 Graph
-> 可能做社区发现 / 图摘要
-> query 时结合向量检索 + 图检索
-> LLM 基于文本片段 + 图结构回答
```

  它的检索单位不只是 chunk，还包括实体、关系、路径、子图、社区摘要和文档片段。

  比如文档里有这些事实：

```text
Alice 是 Project X 的负责人。
Project X 依赖 Service A。
Service A 由 Bob 维护。
```

  普通 RAG 可能只是召回几个相关 chunk。Graph RAG 会显式形成关系：

```text
Alice -> leads -> Project X
Project X -> depends_on -> Service A
Service A -> maintained_by -> Bob
```

  如果你问“谁可能影响 Project X 的稳定性？”，Graph RAG 可以沿着关系链找到：

```text
Project X -> depends_on -> Service A -> maintained_by -> Bob
```

  这就是普通向量相似度不擅长的关系路径推理。

  两者关系可以这样理解：

```text
普通 RAG：文本相似度检索
Graph RAG：文本检索 + 图结构检索 + 关系推理
```

  Graph RAG 通常不是完全抛弃普通 RAG，而是组合使用：

```text
query
-> 向量检索：找相关文本 chunk
-> 图检索：找相关实体、关系、邻居、路径、社区
-> 合并上下文
-> LLM 回答
```

  普通 RAG 适合 FAQ 问答、文档片段查找、技术文档问答，以及用户问题通常能被局部文本回答的场景。例如“这个函数怎么配置”“这个 API 参数是什么意思”“README 里怎么启动项目”，通常普通 RAG 就够了。

  Graph RAG 更适合需要跨文档整合、多跳关系推理、全局总结，或者文档里有大量实体和关系的场景。例如“这个系统里哪些服务依赖支付网关”“某个客户投诉涉及哪些产品、团队和历史问题”“这堆论文里有哪些研究方向互相关联”。这类问题如果只靠 top-k chunk，召回结果可能很碎片；Graph RAG 可以先围绕实体和关系组织上下文。

  但 Graph RAG 的代价也更高。它需要实体抽取、关系抽取、图存储、图更新，还要处理错误实体和错误关系。查询时也要决定走向量检索、图检索，还是混合检索。图谱抽错后，错误会被结构化地传播，影响可能比普通 chunk 召回错误更大。所以不是所有 RAG 都应该升级成 Graph RAG。

  和 `ch07` 的关系是：`ch07` 当前实现的是普通语义 RAG。

```text
file -> chunk -> embedding -> pgvector -> semantic_search tool
```

  如果后面要演进成 Graph RAG，可能会增加这些组件：

```text
EntityExtractor
RelationExtractor
GraphStore
GraphRetriever
CommunitySummarizer
HybridRetriever
```

  对应新链路会变成：

```text
文档
-> chunking
-> embedding index

同时：
文档 / chunk
-> entity extraction
-> relation extraction
-> graph index

查询时：
query
-> semantic search 找 chunk
-> entity linking 找相关节点
-> graph traversal 找邻居 / 路径 / 子图
-> 合并 chunk + graph context
-> LLM 回答
```

  所以你可以把 Graph RAG 理解成 `ch07` 后面可能扩展出的“结构化知识增强检索层”。它解决的是普通向量 RAG 对关系、多跳、全局结构不敏感的问题。

### Q27. Graph RAG 在代码分析领域稳定吗？有什么适合的应用场景？

- **一句总结：**
  代码场景里 Graph RAG 不适合做唯一事实来源，更适合作为“结构化导航层”或“全局关系分析层”：帮助 Agent 找模块关系、依赖路径、影响范围、调用链候选，再回到源码、LSP、grep 做验证。

- **详细回答：**
  是的，Graph RAG 在代码分析领域不是天然稳定。它更适合“关系已经比较明确，且图结构收益大于抽取错误成本”的场景，而不是替代 `grep / read / LSP` 这类确定性工具。

  代码本身是高精度系统。函数名、类型、调用关系、依赖边界通常应该由确定性工具解析出来，而不是让 LLM 从文本里猜。如果 Graph RAG 用 LLM 抽图，可能出现这些错误：

```text
函数 A 调用了函数 B       实际没有调用
模块 X 依赖模块 Y         实际只是注释里提到
Service A 由 Bob 维护     实际是历史文档过期
函数 X 负责鉴权          实际只是部分参与鉴权
```

  这些错误一旦进图，就不只是一次检索错误，而会影响后续路径搜索、影响分析、社区聚类和总结。所以代码领域的关键原则是：图谱里的硬关系尽量来自确定性分析，LLM 只负责解释、总结、补充弱语义，不负责生成唯一事实。

  适合 Graph RAG 的代码场景主要有几类。

  - **代码依赖图 / 调用图分析。**
    用 AST、LSP、编译器、静态分析工具抽取真实关系，例如 `function A -> calls -> function B`、`module X -> imports -> module Y`、`class C -> implements -> interface I`、`route R -> handled_by -> Handler H`。这类图适合回答“修改这个函数会影响哪些调用方”“这个模块被哪些服务依赖”“这个接口有哪些实现”“某个 API 请求经过哪些 handler / service / repository”。

  - **影响范围分析。**
    当你改一个函数、配置、数据库表、API schema 时，可以沿图找上下游：`changed function -> callers -> API handlers -> tests -> dependent packages`。这比纯向量搜索更强，因为影响关系不是文本相似，而是结构依赖。

  - **大型代码库导航。**
    对超大仓库，Graph RAG 可以先给 Agent 一个结构地图，比如模块有哪些、模块之间怎么依赖、核心入口在哪里、某个功能跨哪些目录。然后 Agent 再用 `grep / read` 深入具体文件。

  - **架构理解和文档生成。**
    Graph RAG 可以把代码结构、调用关系、README、设计文档结合起来，生成服务拓扑、模块职责、关键流程图、依赖关系说明、风险传播路径。但最终仍应该引用源码位置，不能只信图摘要。

  - **跨系统实体关系。**
    如果项目涉及服务、数据库表、消息队列、API、配置、部署资源，Graph RAG 很有用。例如 `Service A -> writes -> table users`、`Service B -> consumes -> topic order_created`、`API X -> depends_on -> Redis key Y`、`Job Z -> updates -> feature flag F`。这类关系分散在代码、配置、Terraform、文档里，单纯 grep 很难形成全局视角。

  - **安全 / 合规 / 数据流分析。**
    比如追踪敏感数据：`user.email -> request DTO -> service layer -> database column -> log output -> analytics event`。图结构适合做路径追踪。但这种场景要求图关系必须高度可靠，最好来自静态分析、类型系统、数据流分析，而不是纯 LLM 抽取。

  不适合的场景也要明确。精确查找函数定义、确认某个调用是否存在、代码正在频繁变化的小项目、或者关系没有可靠抽取来源时，Graph RAG 往往不如 `grep / read / LSP` 稳。尤其是只能靠 LLM 从代码片段里猜关系时，Graph RAG 风险很高。

  更稳的工程形态不是：

```text
LLM 读代码 -> 抽实体关系 -> 建图 -> 直接相信图
```

  而是：

```text
AST / LSP / 编译器 / 静态分析
-> 抽取硬关系图

文档 / 注释 / README / issue
-> LLM 抽取软语义图

查询时：
-> 图检索找候选路径
-> grep/read/LSP 回源验证
-> LLM 基于已验证源码回答
```

  也就是说，`Graph = 导航和候选生成`，`Source code = 最终事实来源`。这才稳定。

  在代码分析里，可以这样分工：

```text
grep/read/LSP：精确事实验证
普通 RAG：自然语言文档、注释、README、历史说明
Graph RAG：模块关系、依赖路径、影响范围、跨系统结构
LLM：解释、总结、规划下一步检索
```

  所以 Graph RAG 在代码领域不是没用，而是不能被当作“替代源码”的检索系统。它更像一张地图：地图能帮你找到路线，但真正过桥之前，还是要看桥是不是真的在那里。
