package main

import (
	"fmt"
	"log"

	"ai-devs3/internal/service"
)

func main() {
	cetralaURL := "https://c3ntrala.ag3nts.org/"

	pageFetcher := &service.HTTPPageFetcher{}
	page, err := pageFetcher.FetchPage(cetralaURL)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	llmClient := &service.OpenAIClient{}
	answer, err := llmClient.FindFlag(page)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Println("LLM analysis:")
	fmt.Println(answer)
}
