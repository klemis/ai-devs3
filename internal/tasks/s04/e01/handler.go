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

// Handler handles the S04E01 task execution
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

// Execute runs the S04E01 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S04E01 image restoration and description task")

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
			return fmt.Errorf("S04E01 task failed: %w", err)
		}
		return fmt.Errorf("S04E01 task failed: %w", err)
	}

	// Print processing statistics if available
	if result.ProcessingStats != nil {
		h.service.PrintProcessingStats(result.ProcessingStats)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Photos processed: %d", result.PhotosProcessed)
	log.Printf("Operations performed: %d", result.OperationsCount)
	log.Printf("Photos selected: %d", len(result.SelectedPhotos))

	fmt.Println("=== Image Restoration Results ===")
	fmt.Printf("Photos processed: %d\n", result.PhotosProcessed)
	fmt.Printf("Operations performed: %d\n", result.OperationsCount)
	fmt.Printf("Photos selected for description: %d\n", len(result.SelectedPhotos))
	fmt.Println("===================================")
	fmt.Println("Image restoration and description successful!")
	fmt.Printf("Final response: %s\n", result.Response)

	return nil
}
