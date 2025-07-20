package e01

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	pkgerrors "ai-devs3/pkg/errors"
)

// Handler orchestrates the S02E01 audio transcription and analysis task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S02E01 handler
func NewHandler(cfg *config.Config) *Handler {
	// Initialize dependencies
	httpClient := http.NewClient(cfg.HTTP)
	llmClient := openai.NewClient(cfg.OpenAI)

	// Create service
	service := NewService(cfg, httpClient, llmClient)

	return &Handler{
		service: service,
		config:  cfg,
	}
}

// Execute runs the S02E01 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S02E01 audio transcription and analysis task")

	// Get API key from environment
	apiKey := os.Getenv("AI_DEVS_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("AI_DEVS_API_KEY environment variable not set")
	}

	// Default audio directory path
	audioDir := "../lessons-md/przesluchania"

	// Check if audio directory exists
	if _, err := os.Stat(audioDir); os.IsNotExist(err) {
		return fmt.Errorf("audio directory not found: %s", audioDir)
	}

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, audioDir, apiKey)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S02E01 task failed: %w", err)
		}
		return fmt.Errorf("S02E01 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Success: %t", result.Success)
	log.Printf("Transcripts processed: %d", result.TranscriptCount)
	log.Printf("Analysis result: %s", result.Analysis.Answer)

	fmt.Println("=== Analysis Thinking ===")
	fmt.Println(result.Analysis.Thinking)
	fmt.Println()

	fmt.Println("=== Final Answer ===")
	fmt.Printf("Professor Maj's institute is located on: %s\n", result.Analysis.Answer)
	fmt.Println()

	if result.Success {
		fmt.Println("Audio transcription and analysis successful!")
		fmt.Printf("Response: %s\n", result.Response)
	} else {
		fmt.Println("Audio transcription and analysis completed but may need manual review:")
		fmt.Printf("Response: %s\n", result.Response)
	}

	// Show combined transcripts length
	fmt.Printf("\nCombined transcripts length: %d characters\n", len(result.TotalTranscripts))

	return nil
}
