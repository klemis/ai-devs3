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

	"github.com/qdrant/go-client/qdrant"
)

// Handler handles the S03E02 task execution
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
	// Initialize Qdrant client
	qdrantClient, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Qdrant.Host,
		Port:   cfg.Qdrant.Port,
		APIKey: cfg.Qdrant.APIKey,
		UseTLS: cfg.Qdrant.UseTLS,
	})
	if err != nil {
		log.Printf("Warning: failed to initialize Qdrant client: %v", err)
		// Continue with nil client to allow graceful error handling in Execute
	}

	// Initialize service with Qdrant connection
	service, err := NewService(httpClient, llmClient, qdrantClient)

	return &Handler{
		config:     cfg,
		httpClient: httpClient,
		llmClient:  llmClient,
		service:    service,
	}
}

// Execute runs the S03E02 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S03E02 weapon reports vector search task")

	// Check if service was initialized properly
	if h.service == nil {
		return fmt.Errorf("service not initialized - check Qdrant connection")
	}

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
			return fmt.Errorf("S03E02 task failed: %w", err)
		}
		return fmt.Errorf("S03E02 task failed: %w", err)
	}

	// Print processing statistics if available
	if result.ProcessingStats != nil {
		h.service.PrintProcessingStats(result.ProcessingStats)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Reports processed: %d", result.ReportsProcessed)
	log.Printf("Answer found: %s", result.Answer)

	fmt.Println("=== Vector Search Results ===")
	fmt.Printf("Reports processed: %d\n", result.ReportsProcessed)
	fmt.Printf("Answer: %s\n", result.Answer)
	fmt.Println("=============================")

	fmt.Println("Weapon reports processing successful!")
	fmt.Printf("Response: %s\n", result.Response)

	return nil
}
