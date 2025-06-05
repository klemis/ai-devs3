package app

import (
	"fmt"
	"log"
)

func (app *App) RunS01E05(apiKey string) (string, error) {
	url := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/cenzura.txt", apiKey)

	// Fetch raw text data (non-JSON)
	data, err := app.httpClient.FetchData(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch data: %w", err)
	}

	fmt.Printf("Fetched data from %s:\n", url)
	fmt.Println(data)

	// Use Ollama to censor the text
	censoredText, err := app.ollamaClient.CensorText(data)
	if err != nil {
		return "", fmt.Errorf("failed to censor text with Ollama: %w", err)
	}

	fmt.Printf("Censored text:%s\n", censoredText)

	// Send the censored text as the answer
	response := app.httpClient.BuildResponse("CENZURA", censoredText)

	res, err := app.httpClient.PostReport(response)
	if err != nil {
		log.Fatalf("Failed to post report: %v", err)
		return "", err
	}

	return res, nil
}
