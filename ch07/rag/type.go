package shared

import (
	"context"
	"time"
)

// Vector 表示 embedding 模型输出的语义向量。RAG 里所有语义相似度搜索最终都要落到这个向量上。
// 文本经过 embedding 后会变成一个 float32 数组，维度由具体模型和配置决定，
// 例如 512 维或 1536 维。
type Vector = []float32

// Chunk 表示 RAG 检索中的最小文本单位。
// 长文档会先被切成多个 Chunk，后续 embedding、向量存储、召回和重排都围绕 Chunk 进行。
type Chunk struct {
	Content string `json:"content"` // Content 是实际文本内容
	Meta    Meta   `json:"meta"`    // Meta 是来源信息
}

// Meta 表示 Chunk 的来源信息。
// 对代码仓库来说，DocumentID 通常是文件路径，StartPos/EndPos 通常是行号；
// 对其他文档，也可以表示字符偏移、段落编号等位置范围。
type Meta struct {
	StartPos   int    `json:"start_pos"`   // 起始位置（可以是行号、字符偏移、段落索引等）
	EndPos     int    `json:"end_pos"`     // 结束位置
	DocumentID string `json:"document_id"` // 文档唯一标识（文件路径、URL、ID等）
}

// VectorPoint 表示“一个 Chunk + 它的向量”。
// 这是写入向量库的基本单位：向量用于相似度搜索，Chunk 用于在命中后还原原文和来源。
type VectorPoint struct {
	Vector Vector `json:"vector"`
	Chunk  Chunk  `json:"chunk"`
}

// VectorPointResult 表示向量搜索命中的结果。
// 它在 VectorPoint 的基础上增加 Score，用来表示 query 向量和该 Chunk 向量的相似度。
type VectorPointResult struct {
	VectorPoint
	Score float32 `json:"score"`
}

type RerankService interface {
	Rerank(ctx context.Context, query string, candidates []Chunk) ([]Chunk, error)
}

type EmbeddingService interface {
	Embed(ctx context.Context, chunk string) (Vector, error)
}

type ChunkerService interface {
	Chunk(documentID, content string) []Chunk
}

// VectorStore 向量存储接口，抽象向量数据库操作
type VectorStore interface {
	// InsertBatch 批量插入向量点
	InsertBatch(ctx context.Context, vps []VectorPoint) error

	// Search 执行向量相似度搜索
	Search(ctx context.Context, queryVector Vector, limit int) ([]VectorPointResult, error)

	// DeleteByDocument 删除指定文档的所有向量
	DeleteByDocument(ctx context.Context, documentID string) error

	// GetDocumentIndexedTime 获取文档的索引时间，用于去重判断
	// 返回零值时间表示文档不存在
	GetDocumentIndexedTime(ctx context.Context, documentID string) (time.Time, error)

	// Clear 清空所有向量数据
	Clear(ctx context.Context) error

	// Close 关闭连接
	Close() error
}
