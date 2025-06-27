package main

import (
	"ai-devs3/internal/app"
	"ai-devs3/internal/service"
	"log"
)

func main() {
	llmClient := &service.OpenAIClient{}
	client := &service.HTTPClientImpl{}

	app := app.NewApp(client, nil, llmClient, nil, nil)

	err := app.RunOCR()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
}
