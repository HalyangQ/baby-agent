package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"babyagent/ch07/rag"
)

// PGVectorStore 使用 pgvector 存储和检索向量
type PGVectorStore struct {
	db        *gorm.DB
	dimension int
}

// DocumentChunk 文档块模型
type DocumentChunk struct {
	ID         uint      `gorm:"primaryKey"`
	Content    string    `gorm:"type:text;not null"`
	DocumentID string    `gorm:"type:text;not null;index"`
	StartPos   int       `gorm:"not null"`
	EndPos     int       `gorm:"not null"`
	Embedding  string    `gorm:"type:vector(1536)"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (DocumentChunk) TableName() string {
	return "document_chunks"
}

// ToChunk 转换为 shared.Chunk
func (d *DocumentChunk) ToChunk() shared.Chunk {
	return shared.Chunk{
		Content: d.Content,
		Meta: shared.Meta{
			DocumentID: d.DocumentID,
			StartPos:   d.StartPos,
			EndPos:     d.EndPos,
		},
	}
}

// Config pgvector 配置
type Config struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	Dimension int
}

// NewPGVectorStore 创建一个新的 PGVectorStore
func NewPGVectorStore(config Config) (*PGVectorStore, error) {
	if config.Dimension == 0 {
		config.Dimension = 1536
	}
	log.Printf("[ch07:pgvector] connect host=%s port=%d database=%s dimension=%d", config.Host, config.Port, config.Database, config.Dimension)

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Database,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	store := &PGVectorStore{
		db:        db,
		dimension: config.Dimension,
	}

	// 初始化表
	if err := store.initTable(); err != nil {
		return nil, fmt.Errorf("failed to initialize table: %w", err)
	}
	log.Printf("[ch07:pgvector] ready database=%s dimension=%d", config.Database, config.Dimension)

	return store, nil
}

// initTable 创建必要的表和索引
func (s *PGVectorStore) initTable() error {
	log.Printf("[ch07:pgvector] initializing table and vector extension")
	// 创建扩展
	if err := s.db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		return fmt.Errorf("failed to create vector extension: %w", err)
	}

	// 自动迁移（创建表）
	if err := s.db.AutoMigrate(&DocumentChunk{}); err != nil {
		return fmt.Errorf("failed to migrate table: %w", err)
	}

	// 创建向量索引（IVFFlat 索引，适用于余弦相似度）
	indexSQL := `
		CREATE INDEX IF NOT EXISTS idx_document_chunks_embedding
		ON document_chunks
		USING ivfflat (embedding vector_cosine_ops)
		WITH (lists = 100)
	`
	if err := s.db.Exec(indexSQL).Error; err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	log.Printf("[ch07:pgvector] initialized table=document_chunks index=idx_document_chunks_embedding")

	return nil
}

// Insert 插入一个向量点
func (s *PGVectorStore) Insert(ctx context.Context, vp shared.VectorPoint) error {
	if len(vp.Vector) != s.dimension {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", s.dimension, len(vp.Vector))
	}
	log.Printf("[ch07:pgvector] insert document=%s range=%d-%d vector_dim=%d",
		vp.Chunk.Meta.DocumentID, vp.Chunk.Meta.StartPos, vp.Chunk.Meta.EndPos, len(vp.Vector))

	doc := &DocumentChunk{
		Content:    vp.Chunk.Content,
		DocumentID: vp.Chunk.Meta.DocumentID,
		StartPos:   vp.Chunk.Meta.StartPos,
		EndPos:     vp.Chunk.Meta.EndPos,
		Embedding:  vectorToPGVector(vp.Vector),
	}

	return s.db.WithContext(ctx).Create(doc).Error
}

// InsertBatch 批量插入向量点
func (s *PGVectorStore) InsertBatch(ctx context.Context, vps []shared.VectorPoint) error {
	start := time.Now()
	log.Printf("[ch07:pgvector] insert_batch count=%d dimension=%d", len(vps), s.dimension)
	docs := make([]*DocumentChunk, len(vps))
	for i, vp := range vps {
		if len(vp.Vector) != s.dimension {
			return fmt.Errorf("vector dimension mismatch at index %d: expected %d, got %d", i, s.dimension, len(vp.Vector))
		}

		docs[i] = &DocumentChunk{
			Content:    vp.Chunk.Content,
			DocumentID: vp.Chunk.Meta.DocumentID,
			StartPos:   vp.Chunk.Meta.StartPos,
			EndPos:     vp.Chunk.Meta.EndPos,
			Embedding:  vectorToPGVector(vp.Vector),
		}
	}

	if err := s.db.WithContext(ctx).CreateInBatches(docs, 100).Error; err != nil {
		return err
	}
	log.Printf("[ch07:pgvector] insert_batch completed count=%d duration=%s", len(vps), time.Since(start))
	return nil
}

// Search 执行向量相似度搜索
func (s *PGVectorStore) Search(ctx context.Context, queryVector shared.Vector, limit int) ([]shared.VectorPointResult, error) {
	if len(queryVector) != s.dimension {
		return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d", s.dimension, len(queryVector))
	}
	start := time.Now()
	log.Printf("[ch07:pgvector] search vector_dim=%d limit=%d", len(queryVector), limit)

	vectorStr := vectorToPGVector(queryVector)

	var results []struct {
		ID         int
		Content    string
		DocumentID string
		StartPos   int
		EndPos     int
		Score      float32
	}

	query := `
		SELECT id, content, document_id, start_pos, end_pos,
		       1 - (embedding <=> ?) as score
		FROM document_chunks
		ORDER BY embedding <=> ?
		LIMIT ?
	`

	err := s.db.WithContext(ctx).Raw(query, vectorStr, vectorStr, limit).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	vectorPointResults := make([]shared.VectorPointResult, len(results))
	for i, r := range results {
		vectorPointResults[i] = shared.VectorPointResult{
			VectorPoint: shared.VectorPoint{
				Vector: nil,
				Chunk: shared.Chunk{
					Content: r.Content,
					Meta: shared.Meta{
						DocumentID: r.DocumentID,
						StartPos:   r.StartPos,
						EndPos:     r.EndPos,
					},
				},
			},
			Score: r.Score,
		}
	}

	log.Printf("[ch07:pgvector] search completed results=%d duration=%s", len(vectorPointResults), time.Since(start))
	return vectorPointResults, nil
}

// DeleteByDocument 删除指定文档的所有向量
func (s *PGVectorStore) DeleteByDocument(ctx context.Context, documentID string) error {
	log.Printf("[ch07:pgvector] delete_document document=%s", documentID)
	return s.db.WithContext(ctx).Where("document_id = ?", documentID).Delete(&DocumentChunk{}).Error
}

// Clear 清空表
func (s *PGVectorStore) Clear(ctx context.Context) error {
	log.Printf("[ch07:pgvector] clear table=document_chunks")
	return s.db.WithContext(ctx).Exec("TRUNCATE TABLE document_chunks").Error
}

// Count 返回文档块总数
func (s *PGVectorStore) Count(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&DocumentChunk{}).Count(&count).Error
	return count, err
}

// GetByDocumentID 获取指定文档的所有文档块
func (s *PGVectorStore) GetByDocumentID(ctx context.Context, documentID string) ([]shared.Chunk, error) {
	var docs []DocumentChunk
	err := s.db.WithContext(ctx).Where("document_id = ?", documentID).Find(&docs).Error
	if err != nil {
		return nil, err
	}

	chunks := make([]shared.Chunk, len(docs))
	for i, d := range docs {
		chunks[i] = d.ToChunk()
	}

	return chunks, nil
}

// Close 关闭数据库连接
func (s *PGVectorStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// vectorToPGVector 将 shared.Vector 转换为 pgvector 格式字符串
func vectorToPGVector(v shared.Vector) string {
	if len(v) == 0 {
		return "[]"
	}

	strValues := make([]string, len(v))
	for i, val := range v {
		strValues[i] = fmt.Sprintf("%f", val)
	}

	return "[" + strings.Join(strValues, ",") + "]"
}

// GetDocumentIndexedTime 获取文档的索引时间（返回最早的索引时间）
func (s *PGVectorStore) GetDocumentIndexedTime(ctx context.Context, documentID string) (time.Time, error) {
	var doc DocumentChunk
	err := s.db.WithContext(ctx).Where("document_id = ?", documentID).Order("created_at ASC").First(&doc).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("[ch07:pgvector] document_not_indexed document=%s", documentID)
			return time.Time{}, nil // 文档不存在，返回零值时间
		}
		return time.Time{}, err
	}
	log.Printf("[ch07:pgvector] document_indexed document=%s indexed_at=%s", documentID, doc.CreatedAt.Format(time.RFC3339))
	return doc.CreatedAt, nil
}

// GetDocumentChunkCount 获取文档的文档块数量
func (s *PGVectorStore) GetDocumentChunkCount(ctx context.Context, documentID string) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&DocumentChunk{}).Where("document_id = ?", documentID).Count(&count).Error
	return count, err
}
