package ch01

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"babyagent/shared"
)

type RequestMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ResponseMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	FinishReason string `json:"finish_reason"`

	ReasoningContent *string `json:"reasoning_content"` // vary by different model provider
	Reasoning        *string `json:"reasoning"`         // vary by different model provider
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type OpenAIChatCompletionResponse struct {
	Choices []struct {
		Message ResponseMessage `json:"message"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
}

type OpenAIChatCompletionStreamChunk struct {
	Choices []struct {
		Delta ResponseMessage `json:"delta"`
	} `json:"choices"`
	Usage *Usage `json:"usage,omitempty"`
}

type OpenAIChatCompletionRequest struct {
	Model         string                  `json:"model"`
	Messages      []RequestMessage        `json:"messages"`
	Stream        bool                    `json:"stream"`
	StreamOptions *OpenAIStreamOptionsReq `json:"stream_options,omitempty"`
}

type OpenAIStreamOptionsReq struct {
	IncludeUsage bool `json:"include_usage"`
}

func NonStreamingRequestRawHTTP(ctx context.Context, modelConf shared.ModelConfig, query string) {
	start := time.Now()
	client := http.Client{}

	requestBody := OpenAIChatCompletionRequest{
		Messages: []RequestMessage{
			{Role: "user", Content: query},
		},
		Model:  modelConf.Model,
		Stream: false,
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("failed to marshal request body: %v", err)
		return
	}
	log.Printf("[ch01][raw][nonstream] request prepared model=%s query_len=%d body_len=%d", modelConf.Model, len(query), len(bodyBytes))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", modelConf.BaseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		log.Fatalf("failed to create http request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+modelConf.ApiKey)

	httpResp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("failed to send http request: %v", err)
		return
	}
	defer httpResp.Body.Close()

	log.Printf("[ch01][raw][nonstream] response status=%d elapsed=%s", httpResp.StatusCode, time.Since(start))
	if httpResp.StatusCode != 200 {
		log.Fatalf("failed to send http request: %v", httpResp.StatusCode)
		return
	}

	respBodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		log.Fatalf("failed to read http response: %v", err)
		return
	}

	resp := OpenAIChatCompletionResponse{}
	if err := json.Unmarshal(respBodyBytes, &resp); err != nil {
		log.Fatalf("failed to unmarshal http response: %v", err)
		return
	}
	log.Printf("[ch01][raw][nonstream] response parsed choices=%d body_len=%d", len(resp.Choices), len(respBodyBytes))

	if len(resp.Choices) == 0 {
		log.Printf("no choices returned, resp: %v", resp)
		return
	}
	log.Printf("[ch01][raw][nonstream] resp content len=%d", len(resp.Choices[0].Message.Content))
	log.Printf("resp content: %s", resp.Choices[0].Message.Content)
	log.Printf("token usage: %+v", resp.Usage)
}

func StreamingRequestRawHTTP(ctx context.Context, modelConf shared.ModelConfig, query string) {
	start := time.Now()
	client := http.Client{}

	requestBody := OpenAIChatCompletionRequest{
		Messages: []RequestMessage{
			{Role: "user", Content: query},
		},
		Model:  modelConf.Model,
		Stream: true,
		StreamOptions: &OpenAIStreamOptionsReq{
			IncludeUsage: true,
		},
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatalf("failed to marshal request body: %v", err)
		return
	}
	log.Printf("[ch01][raw][stream] request prepared model=%s query_len=%d body_len=%d", modelConf.Model, len(query), len(bodyBytes))

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", modelConf.BaseURL), bytes.NewReader(bodyBytes))
	if err != nil {
		log.Fatalf("failed to create http request: %v", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+modelConf.ApiKey)
	log.Printf("[ch01][raw][stream] request sent")

	httpResp, err := client.Do(httpReq)
	if err != nil {
		log.Fatalf("failed to send http request: %v", err)
		return
	}
	defer httpResp.Body.Close()

	log.Printf("[ch01][raw][stream] response status=%d elapsed=%s", httpResp.StatusCode, time.Since(start))
	if httpResp.StatusCode != 200 {
		log.Fatalf("failed to send http request: %v", httpResp.StatusCode)
		return
	}

	scanner := bufio.NewScanner(httpResp.Body)
	lineCount := 0
	chunkCount := 0
	doneReceived := false
	outputLen := 0
	var finalUsage *Usage
	log.Printf("[ch01][raw][stream] streaming output begin")
	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "data:") {
			v := strings.TrimPrefix(line, "data:")

			if strings.TrimSpace(v) == "[DONE]" {
				doneReceived = true
				break
			}

			chunk := OpenAIChatCompletionStreamChunk{}
			if err := json.Unmarshal([]byte(v), &chunk); err != nil {
				log.Fatalf("failed to unmarshal chunk: %v", err)
				return
			}
			chunkCount++
			for _, c := range chunk.Choices {
				piece := c.Delta.Content
				if piece != "" {
					fmt.Print(piece)
					outputLen += len(piece)
				}
			}
			if chunk.Usage != nil {
				finalUsage = chunk.Usage
			}
		}
	}

	if scanner.Err() != nil {
		log.Fatalf("failed to read http response: %v", scanner.Err())
		return
	}
	if outputLen > 0 {
		fmt.Println()
	}
	if finalUsage != nil {
		log.Printf("token usage: %+v", finalUsage)
	}
	log.Printf("[ch01][raw][stream] stream finished elapsed=%s lines=%d chunks=%d output_len=%d done=%t", time.Since(start), lineCount, chunkCount, outputLen, doneReceived)
}
