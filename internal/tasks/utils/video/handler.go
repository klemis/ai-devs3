package video

import (
	"context"
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
)

// Handler handles the video transcription utility execution
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

// Execute runs the video transcription utility with default URL
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting video transcription processing")

	// Get video URL from command line args or use default
	args := os.Args
	var videoURL string

	if len(args) > 2 && args[len(args)-1] != "video" {
		// Use the last argument as video URL if provided
		videoURL = args[len(args)-1]
		log.Printf("Using provided video URL: %s", videoURL)
	} else {
		// Use default video URL
		videoURL = h.service.GetDefaultVideoURL()
		log.Printf("Using default video URL: %s", videoURL)
	}

	// Process the video
	result, err := h.service.ProcessVideoFromURL(ctx, videoURL)
	if err != nil {
		return fmt.Errorf("failed to process video: %w", err)
	}

	// Display results
	if result.Error != "" {
		log.Printf("Error occurred: %s", result.Error)
		return fmt.Errorf("video transcription failed: %s", result.Error)
	}

	fmt.Printf("Processed video from %s\n", result.VideoURL)
	if result.Duration != "" {
		fmt.Printf("Duration: %s\n", result.Duration)
	}
	if result.FileSize > 0 {
		fmt.Printf("Audio file size: %d bytes\n", result.FileSize)
	}

	// Save transcription to file
	transcriptPath, err := h.service.SaveTranscriptionToFile(result.Transcription, result.VideoURL)
	if err != nil {
		log.Printf("Warning: failed to save transcription to file: %v", err)
	} else {
		fmt.Printf("Transcription saved to: %s\n", transcriptPath)
	}

	// Display audio file location if available
	if result.AudioFile != "" {
		fmt.Printf("Processed audio file saved to data directory\n")
	}

	fmt.Printf("Transcription: %s\n", result.Transcription)

	return nil
}

// ExecuteWithURL runs the video transcription utility with a specific URL
func (h *Handler) ExecuteWithURL(ctx context.Context, videoURL string) error {
	log.Printf("Starting video transcription processing for URL: %s", videoURL)

	// Process the video
	result, err := h.service.ProcessVideoFromURL(ctx, videoURL)
	if err != nil {
		return fmt.Errorf("failed to process video: %w", err)
	}

	// Display results
	if result.Error != "" {
		log.Printf("Error occurred: %s", result.Error)
		return fmt.Errorf("video transcription failed: %s", result.Error)
	}

	fmt.Printf("Processed video from %s\n", result.VideoURL)
	if result.Duration != "" {
		fmt.Printf("Duration: %s\n", result.Duration)
	}
	if result.FileSize > 0 {
		fmt.Printf("Audio file size: %d bytes\n", result.FileSize)
	}

	// Save transcription to file
	transcriptPath, err := h.service.SaveTranscriptionToFile(result.Transcription, result.VideoURL)
	if err != nil {
		log.Printf("Warning: failed to save transcription to file: %v", err)
	} else {
		fmt.Printf("Transcription saved to: %s\n", transcriptPath)
	}

	// Display audio file location if available
	if result.AudioFile != "" {
		fmt.Printf("Processed audio file saved to data directory\n")
	}

	fmt.Printf("Transcription: %s\n", result.Transcription)

	return nil
}
