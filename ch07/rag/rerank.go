package shared

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPRerankConfig HTTP Rerank服务配置
type HTTPRerankConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

// DefaultHTTPRerankConfig 默认配置
func DefaultHTTPRerankConfig(apiKey string) HTTPRerankConfig {
	return HTTPRerankConfig{
		APIKey:  apiKey,
		BaseURL: "",
		Model:   "rerank",
	}
}

// HTTPRerankService HTTP Rerank服务实现
type HTTPRerankService struct {
	client *resty.Client
	config HTTPRerankConfig
}

// NewHTTPRerankService 创建HTTP Rerank服务
func NewHTTPRerankService(config HTTPRerankConfig) *HTTPRerankService {
	client := resty.New().
		SetBaseURL(config.BaseURL).
		SetHeader("Authorization", "Bearer "+config.APIKey).
		SetHeader("Content-Type", "application/json")

	return &HTTPRerankService{
		client: client,
		config: config,
	}
}

// rerankRequest Rerank请求
type rerankRequest struct {
	Model     string   `json:"model"`
	Query     string   `json:"query"`
	Documents []string `json:"documents"`
	TopN      int      `json:"top_n,omitempty"`
}

// rerankResponse Rerank响应
type rerankResponse struct {
	ID      string `json:"id"`
	Results []struct {
		Document       string  `json:"document"`
		Index          int     `json:"index"`
		RelevanceScore float32 `json:"relevance_score"`
	} `json:"results"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Created   string `json:"created"`
	RequestID string `json:"request_id"`
}

// Rerank 对候选文档进行重排序
func (s *HTTPRerankService) Rerank(ctx context.Context, query string, candidates []Chunk) ([]Chunk, error) {
	if len(candidates) == 0 {
		log.Printf("[ch07:rerank] skip empty candidates query=%q", query)
		return candidates, nil
	}
	start := time.Now()
	log.Printf("[ch07:rerank] request model=%s query=%q candidates=%d", s.config.Model, query, len(candidates))

	// 构建请求
	documents := make([]string, len(candidates))
	for i, chunk := range candidates {
		documents[i] = chunk.Content
	}

	req := rerankRequest{
		Model:     s.config.Model,
		Query:     query,
		Documents: documents,
		TopN:      len(candidates),
	}

	var body rerankResponse
	r := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&body)

	resp, err := r.Post("/rerank")
	if err != nil {
		return nil, fmt.Errorf("failed to call rerank API: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("rerank API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	// 根据重排序结果重新组织候选文档
	result := make([]Chunk, len(body.Results))
	for i, item := range body.Results {
		if item.Index < 0 || item.Index >= len(candidates) {
			return nil, fmt.Errorf("rerank response index out of range: %d candidates=%d", item.Index, len(candidates))
		}
		result[i] = candidates[item.Index]
	}

	log.Printf("[ch07:rerank] response results=%d prompt_tokens=%d total_tokens=%d duration=%s",
		len(result), body.Usage.PromptTokens, body.Usage.TotalTokens, time.Since(start))
	return result, nil
}
