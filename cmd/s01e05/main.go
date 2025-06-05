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
	ollamaClient := service.NewOllamaClient()

	app := app.NewApp(client, nil, nil, nil, ollamaClient)

	answer, err := app.RunS01E05(apiKey)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Printf("answer: %s\n", answer)
}
