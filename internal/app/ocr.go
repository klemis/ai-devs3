package app

import (
	"fmt"
)

func (app *App) RunOCR() error {
	url := "https://assets-v2.circle.so/837mal5q2pf3xskhmfuybrh0uwnd"

	data, err := app.httpClient.FetchImage(url)
	if err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	fmt.Printf("Fetched data from %s\n", url)

	// ProcessImage
	// processedImage, err := app.imageProcessor.ProcessImage(path, 2048)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to process image: %w", err)
	// }
	//
	// base64Data := base64.StdEncoding.EncodeToString(imageBytes)

	// ExtractTextFromImage
	text, err := app.llmClient.ExtractTextFromImage(data)
	if err != nil {
		return fmt.Errorf("failed to extract text from image: %w", err)
	}

	fmt.Printf("Extracted text:%s\n", text)

	return nil
}
