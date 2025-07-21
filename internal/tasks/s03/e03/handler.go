package e03

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

// Handler handles the S03E03 task execution
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

// Execute runs the S03E03 database task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S03E03 database query task")

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
			return fmt.Errorf("S03E03 task failed: %w", err)
		}
		return fmt.Errorf("S03E03 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Generated SQL query: %s", result.GeneratedQuery)
	log.Printf("Found datacenter IDs: %v", result.DatacenterIDs)
	log.Printf("Processing time: %.2f seconds", result.ProcessingTime)

	fmt.Println("=== Database Query Results ===")
	fmt.Printf("Generated Query: %s\n", result.GeneratedQuery)
	fmt.Printf("Active Datacenters (inactive managers): %v\n", result.DatacenterIDs)
	fmt.Printf("Total count: %d\n", len(result.DatacenterIDs))
	fmt.Printf("Processing time: %.2f seconds\n", result.ProcessingTime)
	fmt.Println("==============================")
	fmt.Println("Database task successful!")
	fmt.Printf("Response: %s\n", result.Response)

	return nil
}
