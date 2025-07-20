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

// Handler handles the S03E01 task execution
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

// Execute runs the S03E01 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S03E01 documents processing task")

	// Get API key from environment
	apiKey := h.config.AIDevs.APIKey
	if apiKey == "" {
		return fmt.Errorf("AI_DEVS_API_KEY is required")
	}

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, apiKey)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S03E01 task failed: %w", err)
		}
		return fmt.Errorf("S03E01 task failed: %w", err)
	}

	// Print processing statistics if available
	if result.ProcessingStats != nil {
		h.service.PrintProcessingStats(result.ProcessingStats)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Total files processed: %d", result.TotalFiles)
	log.Printf("Keywords generated: %d", result.KeywordsGenerated)

	fmt.Println("=== Documents Processing Results ===")
	fmt.Printf("Files processed: %d\n", result.TotalFiles)
	fmt.Printf("Keywords generated: %d\n", result.KeywordsGenerated)
	fmt.Println("====================================")
	fmt.Println("Documents processing successful!")
	fmt.Printf("Response: %s\n", result.Response)

	return nil
}
