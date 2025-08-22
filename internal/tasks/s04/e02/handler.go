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

// Handler handles the S04E02 task execution
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

// Execute runs the S04E02 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S04E02 text classification research task")

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
			return fmt.Errorf("S04E02 task failed: %w", err)
		}
		return fmt.Errorf("S04E02 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Total lines processed: %d", result.TotalLines)
	log.Printf("Correct classifications: %d", result.CorrectCount)
	log.Printf("Correct answer IDs: %v", result.CorrectAnswers)

	fmt.Println("=== Text Classification Results ===")
	fmt.Printf("Total lines processed: %d\n", result.TotalLines)
	fmt.Printf("Reliable classifications: %d\n", result.CorrectCount)
	fmt.Printf("Correct answer IDs: %v\n", result.CorrectAnswers)
	fmt.Println("====================================")
	fmt.Println("Text classification task successful!")
	fmt.Printf("Central response: %s\n", result.Response)

	return nil
}
