package e05

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/ollama"
	pkgerrors "ai-devs3/pkg/errors"
)

// Handler orchestrates the S01E05 text censoring task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S01E05 handler
func NewHandler(cfg *config.Config) *Handler {
	// Initialize dependencies
	httpClient := http.NewClient(cfg.HTTP)
	ollamaClient := ollama.NewClient(cfg.Ollama)

	// Create service
	service := NewService(cfg, httpClient, ollamaClient)

	return &Handler{
		service: service,
		config:  cfg,
	}
}

// Execute runs the S01E05 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S01E05 text censoring task")

	// Get API key from environment
	apiKey := os.Getenv("AI_DEVS_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("AI_DEVS_API_KEY environment variable not set")
	}

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, apiKey)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S01E05 task failed: %w", err)
		}
		return fmt.Errorf("S01E05 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Original text length: %d", len(result.OriginalText))
	log.Printf("Censored text length: %d", len(result.CensoredText))

	fmt.Println("=== Original Text ===")
	fmt.Println(result.OriginalText)
	fmt.Println()

	fmt.Println("=== Censored Text ===")
	fmt.Println(result.CensoredText)
	fmt.Println()

	fmt.Println("Text censoring successful!")
	fmt.Printf("Response: %s\n", result.Response)

	return nil
}
