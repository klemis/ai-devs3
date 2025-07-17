package e02

import (
	"context"
	"errors"
	"fmt"
	"log"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	pkgerrors "ai-devs3/pkg/errors"
)

// Handler orchestrates the S01E02 RoboISO verification task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S01E02 handler
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

// Execute runs the S01E02 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S01E02 RoboISO verification task")

	// Task configuration
	verifyURL := "https://xyz.ag3nts.org/verify"

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, verifyURL)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S01E02 task failed: %w", err)
		}
		return fmt.Errorf("S01E02 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Final response: %s", result.FinalResponse)
	log.Printf("Success: %t", result.Success)
	log.Printf("Message count: %d", result.MessageCount)

	return nil
}
