package main

import (
	"fmt"
	"log"

	"ai-devs3/internal/app"
	"ai-devs3/internal/domain"
	"ai-devs3/internal/service"
)

func main() {
	// Configuration
	loginURL := "https://xyz.ag3nts.org/"
	creds := domain.Credentials{
		Username: "tester",
		Password: "574e112a",
	}

	// Create dependencies
	pageFetcher := &service.HTTPPageFetcher{}
	questionExtractor := &service.RegexQuestionExtractor{}
	llmClient := &service.OpenAIClient{}
	httpService := &service.HTTPService{URL: loginURL}

	// Create and run the application
	robotApp := app.NewApp(pageFetcher, questionExtractor, llmClient, httpService)
	secretContent, err := robotApp.Run(loginURL, creds)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Println("Secret page content:")
	fmt.Println(secretContent)
	fmt.Println("Please submit the flag to https://c3ntrala.ag3nts.org/")
}
