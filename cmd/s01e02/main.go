package main

import (
	"fmt"
	"log"

	"ai-devs3/internal/app"
	"ai-devs3/internal/service"
)

func main() {
	// Configuration
	url := "https://xyz.ag3nts.org/verify"

	// Create dependencies
	roboISOService := &service.RoboISOService{URL: url}
	llmClient := &service.OpenAIClient{}

	app := app.NewApp(nil, nil, llmClient, roboISOService, nil)

	answer, err := app.RunS01E02()
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("answer: %s\n", answer)

}
