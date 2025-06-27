package app

import (
	"fmt"
	"log"
)

// RunS02E03 handles the robot description and image generation task
func (app *App) RunS02E03(apiKey string) (string, error) {
	// Step 1: Fetch robot description from the API
	url := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/robotid.json", apiKey)

	log.Printf("Fetching robot description from: %s", url)

	data, err := app.httpClient.FetchJSONData(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch robot description: %w", err)
	}

	// Extract the description from the JSON response
	description, ok := data["description"].(string)
	if !ok {
		return "", fmt.Errorf("description field not found or not a string in response")
	}

	log.Printf("Original robot description: %s", description)

	// Step 2: Use OpenAI to analyze the description and extract keywords for DALL-E
	dallePrompt, err := app.llmClient.ExtractKeywordsForDALLE(description)
	if err != nil {
		return "", fmt.Errorf("failed to extract keywords for DALL-E: %w", err)
	}

	log.Printf("DALL-E optimized prompt: %s", dallePrompt)

	// Step 3: Generate image using DALL-E 3
	imageURL, err := app.llmClient.GenerateImageWithDALLE(dallePrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate image with DALL-E: %w", err)
	}

	log.Printf("Generated image URL: %s", imageURL)

	// Step 4: Send the image URL as the answer
	response := app.httpClient.BuildResponse("robotid", imageURL)

	result, err := app.httpClient.PostReport(response)
	if err != nil {
		return "", fmt.Errorf("failed to post report: %w", err)
	}

	return result, nil
}
