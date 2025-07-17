package e03

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

// Handler orchestrates the S01E03 JSON data processing task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S01E03 handler
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

// Execute runs the S01E03 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S01E03 JSON data processing task")

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
			return fmt.Errorf("S01E03 task failed: %w", err)
		}
		return fmt.Errorf("S01E03 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Corrections made: %d", result.Corrected)
	log.Printf("LLM answers provided: %d", result.LLMAnswers)
	log.Printf("Math answers calculated: %d", result.MathAnswers)
	fmt.Printf("Response: %s\n", result.Response)

	return nil
}
