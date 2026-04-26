package shared

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

// HTTPEmbeddingConfig HTTP Embedding服务配置
type HTTPEmbeddingConfig struct {
	APIKey     string
	BaseURL    string
	Model      string
	Dimensions int
}

// DefaultHTTPEmbeddingConfig 默认配置
func DefaultHTTPEmbeddingConfig(apiKey string) HTTPEmbeddingConfig {
	return HTTPEmbeddingConfig{
		APIKey:     apiKey,
		BaseURL:    "",
		Model:      "embedding",
		Dimensions: 512,
	}
}

// HTTPEmbeddingService HTTP Embedding服务实现
type HTTPEmbeddingService struct {
	client *resty.Client
	config HTTPEmbeddingConfig
}

// NewHTTPEmbeddingService 创建HTTP Embedding服务
func NewHTTPEmbeddingService(config HTTPEmbeddingConfig) *HTTPEmbeddingService {
	client := resty.New().
		SetBaseURL(config.BaseURL).
		SetHeader("Authorization", "Bearer "+config.APIKey).
		SetHeader("Content-Type", "application/json")

	return &HTTPEmbeddingService{
		client: client,
		config: config,
	}
}

// embeddingRequest Embedding请求
type embeddingRequest struct {
	Model      string `json:"model"`
	Input      string `json:"input"`
	Dimensions int    `json:"dimensions,omitempty"`
}

// embeddingResponse Embedding响应
type embeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Index     int       `json:"index"`
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// Embed 将文本转换为向量
func (s *HTTPEmbeddingService) Embed(ctx context.Context, text string) (Vector, error) {
	start := time.Now()
	log.Printf("[ch07:embedding] request model=%s dimensions=%d chars=%d", s.config.Model, s.config.Dimensions, len(text))
	var body embeddingResponse

	req := embeddingRequest{
		Model: s.config.Model,
		Input: text,
	}

	// 只有非零维度才设置
	if s.config.Dimensions > 0 {
		req.Dimensions = s.config.Dimensions
	}

	r := s.client.R().
		SetContext(ctx).
		SetBody(req).
		SetResult(&body)

	resp, err := r.Post("/embeddings")
	if err != nil {
		return nil, fmt.Errorf("failed to call embedding API: %w", err)
	}
	if resp.IsError() {
		return nil, fmt.Errorf("embedding API returned status %d: %s", resp.StatusCode(), resp.String())
	}

	if len(body.Data) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}

	vector := Vector(body.Data[0].Embedding)
	log.Printf("[ch07:embedding] response model=%s vector_dim=%d prompt_tokens=%d total_tokens=%d duration=%s",
		body.Model, len(vector), body.Usage.PromptTokens, body.Usage.TotalTokens, time.Since(start))
	return vector, nil
}
