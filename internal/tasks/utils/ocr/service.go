package ocr

import (
	"context"
	"fmt"
	"log"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
)

// Service handles the OCR processing task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
}

// NewService creates a new service instance
func NewService(httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
	}
}

// ProcessImageFromURL fetches an image from URL and extracts text using OCR
func (s *Service) ProcessImageFromURL(imageURL string) (*OCRResult, error) {
	log.Printf("Fetching image from URL: %s", imageURL)

	// Fetch binary image data
	imageData, err := s.httpClient.FetchBinaryData(context.Background(), imageURL)
	if err != nil {
		return &OCRResult{
			URL:   imageURL,
			Error: fmt.Sprintf("Failed to fetch image: %v", err),
		}, err
	}

	if len(imageData) == 0 {
		return &OCRResult{
			URL:   imageURL,
			Error: "Received empty image data",
		}, fmt.Errorf("received empty image data")
	}

	log.Printf("Successfully fetched image data (%d bytes)", len(imageData))

	// Extract text from image using LLM
	extractedText, err := s.llmClient.ExtractTextFromImage(context.Background(), imageData)
	if err != nil {
		return &OCRResult{
			URL:   imageURL,
			Error: fmt.Sprintf("Failed to extract text: %v", err),
		}, err
	}

	return &OCRResult{
		URL:           imageURL,
		ExtractedText: extractedText,
	}, nil
}

// GetDefaultImageURL returns the default image URL used in the original OCR task
func (s *Service) GetDefaultImageURL() string {
	return "https://assets-v2.circle.so/837mal5q2pf3xskhmfuybrh0uwnd"
}
