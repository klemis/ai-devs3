package main

import (
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

	answer, err := app.RunS02E02(apiKey)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	log.Printf("Final answer: %s", answer)
}
