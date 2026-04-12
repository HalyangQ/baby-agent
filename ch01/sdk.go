package ch01

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"babyagent/shared"
)

func NonStreamingRequestSDK(ctx context.Context, modelConf shared.ModelConfig, query string) {
	start := time.Now()
	log.Printf("[ch01][sdk][nonstream] build request model=%s query_len=%d", modelConf.Model, len(query))
	client := openai.NewClient(option.WithBaseURL(modelConf.BaseURL), option.WithAPIKey(modelConf.ApiKey))

	req := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(query),
		},
		Model: modelConf.Model,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	}

	resp, err := client.Chat.Completions.New(ctx, req)
	if err != nil {
		log.Fatalf("failed to send a new completion request: %v", err)
		return
	}
	log.Printf("[ch01][sdk][nonstream] response received elapsed=%s choices=%d", time.Since(start), len(resp.Choices))

	if len(resp.Choices) == 0 {
		log.Printf("no choices returned, resp: %s", resp.RawJSON())
		return
	}

	log.Printf("[ch01][sdk][nonstream] resp content len=%d", len(resp.Choices[0].Message.Content))
	log.Printf("resp content: %s", resp.Choices[0].Message.Content)
	log.Printf("token usage: %s", resp.Usage.RawJSON())
}

func StreamingRequestSDK(ctx context.Context, modelConf shared.ModelConfig, query string) {
	start := time.Now()
	log.Printf("[ch01][sdk][stream] build request model=%s query_len=%d", modelConf.Model, len(query))
	client := openai.NewClient(option.WithBaseURL(modelConf.BaseURL), option.WithAPIKey(modelConf.ApiKey))

	req := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(query),
		},
		Model: modelConf.Model,
		StreamOptions: openai.ChatCompletionStreamOptionsParam{
			IncludeUsage: openai.Bool(true),
		},
	}

	stream := client.Chat.Completions.NewStreaming(ctx, req)
	chunkCount := 0
	outputLen := 0
	finalUsage := ""
	log.Printf("[ch01][sdk][stream] streaming output begin")

	for stream.Next() {
		chunk := stream.Current()
		chunkCount++

		streamChunk := OpenAIChatCompletionStreamChunk{}
		if err := json.Unmarshal([]byte(chunk.RawJSON()), &streamChunk); err != nil {
			log.Printf("[ch01][sdk][stream] failed to parse chunk=%d: %v", chunkCount, err)
			continue
		}
		for _, c := range streamChunk.Choices {
			piece := c.Delta.Content
			if piece != "" {
				fmt.Print(piece)
				outputLen += len(piece)
			}
		}

		if chunk.Usage.TotalTokens != 0 {
			finalUsage = chunk.Usage.RawJSON()
		}
	}

	if stream.Err() != nil {
		log.Fatalf("stream error: %v", stream.Err())
		return
	}
	if outputLen > 0 {
		fmt.Println()
	}
	if strings.TrimSpace(finalUsage) != "" {
		log.Printf("token usage: %s", finalUsage)
	}
	log.Printf("[ch01][sdk][stream] stream finished elapsed=%s chunks=%d output_len=%d", time.Since(start), chunkCount, outputLen)
}
