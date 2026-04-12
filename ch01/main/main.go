package main

import (
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"

	"babyagent/ch01"
	"babyagent/shared"
)

func main() {
	_ = godotenv.Load()

	useRaw := flag.Bool("raw", false, "use raw http implementation")
	useStream := flag.Bool("stream", false, "use streaming response")
	query := flag.String("q", "hello", "prompt text")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	modelConf := shared.NewModelConfig()
	mode := "sdk-nonstream"
	if *useRaw && *useStream {
		mode = "raw-stream"
	} else if *useRaw {
		mode = "raw-nonstream"
	} else if *useStream {
		mode = "sdk-stream"
	}
	log.Printf("[ch01] start mode=%s model=%s base_url=%s api_key_set=%t query_len=%d query=%q", mode, modelConf.Model, modelConf.BaseURL, modelConf.ApiKey != "", len(*query), *query)

	switch {
	case *useRaw && *useStream:
		ch01.StreamingRequestRawHTTP(ctx, modelConf, *query)
	case *useRaw:
		ch01.NonStreamingRequestRawHTTP(ctx, modelConf, *query)
	case *useStream:
		ch01.StreamingRequestSDK(ctx, modelConf, *query)
	default:
		ch01.NonStreamingRequestSDK(ctx, modelConf, *query)
	}
}
