package e01

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

// Handler orchestrates the S01E01 robot authentication task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S01E01 handler
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

// Execute runs the S01E01 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S01E01 robot authentication task")

	// Task configuration
	loginURL := "https://xyz.ag3nts.org/"
	creds := &Credentials{
		Username: "tester",
		Password: "574e112a",
	}

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, loginURL, creds)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S01E01 task failed: %w", err)
		}
		return fmt.Errorf("S01E01 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Flag extracted: %s", result.Flag)

	fmt.Println("Secret page content:")
	fmt.Println(result.Content)
	fmt.Println("Please submit the flag to https://c3ntrala.ag3nts.org/")

	return nil
}
