package index

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"babyagent/ch07/rag"
)

// Indexer 索引器，负责将代码仓库索引到向量数据库
type Indexer struct {
	rootPath     string
	fileWalker   *FileWalker
	chunker      shared.ChunkerService
	vectorStore  shared.VectorStore
	embedService shared.EmbeddingService
}

// IndexerConfig 索引器配置
type IndexerConfig struct {
	RootPath    string
	ChunkerType shared.ChunkerType
	MaxLines    int
	MaxChars    int
}

// NewIndexer 创建新的索引器
func NewIndexer(config IndexerConfig, vectorStore shared.VectorStore, embedService shared.EmbeddingService) *Indexer {
	return &Indexer{
		rootPath:     config.RootPath,
		fileWalker:   NewFileWalker(),
		chunker:      shared.NewChunker(config.ChunkerType, config.MaxLines, config.MaxChars),
		vectorStore:  vectorStore,
		embedService: embedService,
	}
}

// Index 执行索引操作
func (idx *Indexer) Index(ctx context.Context) (*IndexResult, error) {
	startTime := time.Now()
	log.Printf("[ch07:index] start serial indexing root=%s", idx.rootPath)

	// 1. 遍历文件
	files, err := idx.fileWalker.Walk(idx.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to walk files: %w", err)
	}
	log.Printf("[ch07:index] discovered files=%d root=%s", len(files), idx.rootPath)

	result := &IndexResult{
		TotalFiles: len(files),
	}

	// 2. 处理每个文件
	for _, filePath := range files {
		fileResult, err := idx.indexFile(ctx, filePath)
		if err != nil {
			result.FailedFiles++
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filePath, err))
			continue
		}

		result.TotalChunks += fileResult.Chunks

		// 根据操作类型更新统计
		switch fileResult.Action {
		case IndexActionSkip:
			result.SkippedFiles++
		case IndexActionReindex:
			result.ReindexedFiles++
			result.SuccessFiles++
		default:
			result.SuccessFiles++
		}
	}

	result.Duration = time.Since(startTime)
	log.Printf("[ch07:index] completed serial indexing total_files=%d success=%d skipped=%d reindexed=%d failed=%d chunks=%d duration=%s",
		result.TotalFiles, result.SuccessFiles, result.SkippedFiles, result.ReindexedFiles, result.FailedFiles, result.TotalChunks, result.Duration)
	return result, nil
}

// IndexConcurrent 并发执行索引操作
func (idx *Indexer) IndexConcurrent(ctx context.Context, concurrency int) (*IndexResult, error) {
	startTime := time.Now()

	files, err := idx.fileWalker.Walk(idx.rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to walk files: %w", err)
	}

	result := &IndexResult{
		TotalFiles: len(files),
	}

	// 创建任务队列
	fileChan := make(chan string, len(files))
	for _, file := range files {
		fileChan <- file
	}
	close(fileChan)

	// 启动 worker
	var wg sync.WaitGroup
	resultMu := sync.Mutex{}

	if concurrency <= 0 {
		concurrency = 10
	}
	log.Printf("[ch07:index] start concurrent indexing root=%s files=%d concurrency=%d", idx.rootPath, len(files), concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for filePath := range fileChan {
				fileResult, err := idx.indexFile(ctx, filePath)
				if err != nil {
					log.Printf("[ch07:index] worker=%d file_failed path=%s err=%v", workerID, filePath, err)
					resultMu.Lock()
					result.FailedFiles++
					result.Errors = append(result.Errors, fmt.Errorf("%s: %w", filePath, err))
					resultMu.Unlock()
					continue
				}
				log.Printf("[ch07:index] worker=%d file_done path=%s action=%s chunks=%d", workerID, filePath, fileResult.Action, fileResult.Chunks)

				resultMu.Lock()
				result.TotalChunks += fileResult.Chunks

				// 根据操作类型更新统计
				switch fileResult.Action {
				case IndexActionSkip:
					result.SkippedFiles++
				case IndexActionReindex:
					result.ReindexedFiles++
					result.SuccessFiles++
				default:
					result.SuccessFiles++
				}
				resultMu.Unlock()
			}
		}(i + 1)
	}

	wg.Wait()
	result.Duration = time.Since(startTime)
	log.Printf("[ch07:index] completed concurrent indexing total_files=%d success=%d skipped=%d reindexed=%d failed=%d chunks=%d duration=%s",
		result.TotalFiles, result.SuccessFiles, result.SkippedFiles, result.ReindexedFiles, result.FailedFiles, result.TotalChunks, result.Duration)
	return result, nil
}

// indexFile 索引单个文件（带去重）
func (idx *Indexer) indexFile(ctx context.Context, filePath string) (*FileIndexResult, error) {
	// 转换为相对路径
	relPath, err := filepath.Rel(idx.rootPath, filePath)
	if err != nil {
		relPath = filePath
	}

	// 获取文件信息（用于检查修改时间）
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// 检查文件是否已索引
	indexedTime, err := idx.vectorStore.GetDocumentIndexedTime(ctx, relPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check file index status: %w", err)
	}

	// 如果文件已索引
	if !indexedTime.IsZero() {
		// 文件未修改，跳过
		if fileInfo.ModTime().Before(indexedTime) || fileInfo.ModTime().Equal(indexedTime) {
			log.Printf("[ch07:index] skip unchanged document=%s file_mtime=%s indexed_at=%s",
				relPath, fileInfo.ModTime().Format(time.RFC3339), indexedTime.Format(time.RFC3339))
			return &FileIndexResult{
				FilePath: filePath,
				Chunks:   0,
				Action:   IndexActionSkip,
			}, nil
		}

		// 文件已修改，删除旧记录
		log.Printf("[ch07:index] reindex modified document=%s file_mtime=%s indexed_at=%s",
			relPath, fileInfo.ModTime().Format(time.RFC3339), indexedTime.Format(time.RFC3339))
		if err := idx.vectorStore.DeleteByDocument(ctx, relPath); err != nil {
			return nil, fmt.Errorf("failed to delete old index: %w", err)
		}
	}

	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 切分文本
	chunks := idx.chunker.Chunk(relPath, string(content))
	if len(chunks) == 0 {
		log.Printf("[ch07:index] skip empty document=%s", relPath)
		return &FileIndexResult{Chunks: 0, Action: IndexActionSkip}, nil
	}
	log.Printf("[ch07:index] chunked document=%s bytes=%d chunks=%d", relPath, len(content), len(chunks))

	// 获取向量嵌入
	vectorPoints, err := idx.embedChunks(ctx, chunks)
	if err != nil {
		return nil, fmt.Errorf("failed to embed chunks: %w", err)
	}

	// 批量插入数据库
	if err := idx.vectorStore.InsertBatch(ctx, vectorPoints); err != nil {
		return nil, fmt.Errorf("failed to insert vectors: %w", err)
	}
	log.Printf("[ch07:index] inserted document=%s chunks=%d", relPath, len(chunks))

	// 确定操作类型
	action := IndexActionNew
	if !indexedTime.IsZero() {
		action = IndexActionReindex
	}

	return &FileIndexResult{
		FilePath: filePath,
		Chunks:   len(chunks),
		Action:   action,
	}, nil
}

// embedChunks 批量获取向量嵌入
func (idx *Indexer) embedChunks(ctx context.Context, chunks []shared.Chunk) ([]shared.VectorPoint, error) {
	vectorPoints := make([]shared.VectorPoint, len(chunks))
	log.Printf("[ch07:index] embedding chunks=%d", len(chunks))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, chunk := range chunks {
		wg.Add(1)
		go func(chunkIdx int, c shared.Chunk) {
			defer wg.Done()

			vector, err := idx.embedService.Embed(ctx, c.Content)
			if err != nil {
				log.Printf("[ch07:index] embed_failed document=%s range=%d-%d err=%v", c.Meta.DocumentID, c.Meta.StartPos, c.Meta.EndPos, err)
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			vectorPoints[chunkIdx] = shared.VectorPoint{
				Vector: vector,
				Chunk:  c,
			}
			mu.Unlock()
		}(i, chunk)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	log.Printf("[ch07:index] embedding completed chunks=%d", len(chunks))
	return vectorPoints, nil
}

// Search 在索引中搜索相似内容
func (idx *Indexer) Search(ctx context.Context, query string, limit int) ([]shared.VectorPointResult, error) {
	start := time.Now()
	log.Printf("[ch07:index] search query=%q limit=%d", query, limit)
	queryVector, err := idx.embedService.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	results, err := idx.vectorStore.Search(ctx, queryVector, limit)
	if err != nil {
		return nil, err
	}
	log.Printf("[ch07:index] search completed results=%d duration=%s", len(results), time.Since(start))
	return results, nil
}

// Clear 清空索引
func (idx *Indexer) Clear(ctx context.Context) error {
	log.Printf("[ch07:index] clear index root=%s", idx.rootPath)
	return idx.vectorStore.Clear(ctx)
}

// Close 关闭索引器
func (idx *Indexer) Close() error {
	return idx.vectorStore.Close()
}

// IndexResult 索引结果
type IndexResult struct {
	TotalFiles     int
	SuccessFiles   int
	FailedFiles    int
	SkippedFiles   int // 跳过的文件（已索引且未修改）
	ReindexedFiles int // 重新索引的文件（已索引但已修改）
	TotalChunks    int
	Duration       time.Duration
	Errors         []error
}

// FileIndexResult 单文件索引结果
type FileIndexResult struct {
	FilePath string
	Chunks   int
	Action   IndexAction // 执行的操作
}

// IndexAction 索引操作类型
type IndexAction string

const (
	IndexActionNew     IndexAction = "new"     // 新索引
	IndexActionSkip    IndexAction = "skip"    // 跳过（已索引且未修改）
	IndexActionReindex IndexAction = "reindex" // 重新索引（已修改）
)
