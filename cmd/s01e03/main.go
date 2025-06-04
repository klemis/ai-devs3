package main

import (
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/app"
	"ai-devs3/internal/service"
)

func main() {
	apiKey := os.Getenv("AI_DEVS_API_KEY")
	if apiKey == "" {
		log.Fatal("AI_DEVS_API_KEY environment variable not set")
	}

	fetcher := &service.HTTPClientImpl{}
	llmClient := &service.OpenAIClient{}

	app := app.NewApp(fetcher, nil, llmClient, nil)

	answer, err := app.RunS01E03(apiKey)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Printf("answer: %s\n", answer)
}
