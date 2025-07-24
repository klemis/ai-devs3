package ocr

import (
	"context"
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
)

// Handler handles the OCR utility execution
type Handler struct {
	config     *config.Config
	httpClient *http.Client
	llmClient  *openai.Client
	service    *Service
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	httpClient := http.NewClient(cfg.HTTP)
	llmClient := openai.NewClient(cfg.OpenAI)
	service := NewService(httpClient, llmClient)

	return &Handler{
		config:     cfg,
		httpClient: httpClient,
		llmClient:  llmClient,
		service:    service,
	}
}

// Execute runs the OCR utility
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting OCR processing")

	// Get image URL from command line args or use default
	args := os.Args
	var imageURL string

	if len(args) > 2 && args[len(args)-1] != "ocr" {
		// Use the last argument as image URL if provided
		imageURL = args[len(args)-1]
		log.Printf("Using provided image URL: %s", imageURL)
	} else {
		// Use default image URL
		imageURL = h.service.GetDefaultImageURL()
		log.Printf("Using default image URL: %s", imageURL)
	}

	// Process the image
	result, err := h.service.ProcessImageFromURL(imageURL)
	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}

	// Display results
	if result.Error != "" {
		log.Printf("Error occurred: %s", result.Error)
		return fmt.Errorf("OCR processing failed: %s", result.Error)
	}

	fmt.Printf("Fetched data from %s\n", result.URL)
	fmt.Printf("Extracted text: %s\n", result.ExtractedText)

	return nil
}

// ExecuteWithURL runs the OCR utility with a specific URL
func (h *Handler) ExecuteWithURL(ctx context.Context, imageURL string) error {
	log.Printf("Starting OCR processing for URL: %s", imageURL)

	// Process the image
	result, err := h.service.ProcessImageFromURL(imageURL)
	if err != nil {
		return fmt.Errorf("failed to process image: %w", err)
	}

	// Display results
	if result.Error != "" {
		log.Printf("Error occurred: %s", result.Error)
		return fmt.Errorf("OCR processing failed: %s", result.Error)
	}

	fmt.Printf("Fetched data from %s\n", result.URL)
	fmt.Printf("Extracted text: %s\n", result.ExtractedText)

	return nil
}
