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

	client := &service.HTTPClientImpl{
		APIKey: apiKey,
	}
	llmClient := &service.OpenAIClient{}

	app := app.NewApp(client, nil, llmClient, nil, nil)

	answer, err := app.RunS02E01()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Printf("answer: %s\n", answer)
}
