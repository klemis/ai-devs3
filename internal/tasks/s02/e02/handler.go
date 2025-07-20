package e02

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/image"
	"ai-devs3/internal/llm/openai"
	pkgerrors "ai-devs3/pkg/errors"
)

// Handler orchestrates the S02E02 map analysis task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S02E02 handler
func NewHandler(cfg *config.Config) *Handler {
	// Initialize dependencies
	httpClient := http.NewClient(cfg.HTTP)
	llmClient := openai.NewClient(cfg.OpenAI)
	imageProcessor := image.NewProcessor(*cfg)

	// Create service
	service := NewService(cfg, httpClient, llmClient, imageProcessor)

	return &Handler{
		service: service,
		config:  cfg,
	}
}

// Execute runs the S02E02 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S02E02 map analysis task")

	// Get API key from environment
	apiKey := os.Getenv("AI_DEVS_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("AI_DEVS_API_KEY environment variable not set")
	}

	// Default fragments directory and count
	fragmentsDir := "data/s02e02"
	numFragments := 4

	// Validate fragments directory
	if err := h.service.ValidateFragmentsDirectory(fragmentsDir, numFragments); err != nil {
		return fmt.Errorf("fragments directory validation failed: %w", err)
	}

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, fragmentsDir, numFragments, apiKey)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S02E02 task failed: %w", err)
		}
		return fmt.Errorf("S02E02 task failed: %w", err)
	}

	// Print analysis result
	h.service.PrintAnalysisResult(result.AnalysisResult)

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Identified City: %s", result.IdentifiedCity)
	log.Printf("Confidence: %s", result.Confidence)
	log.Printf("Fragments processed: %d", result.FragmentCount)

	fmt.Println("Map analysis successful!")
	fmt.Printf("Final Answer: %s\n", result.IdentifiedCity)
	fmt.Println("Please submit your answer to Centrala manually.")

	return nil
}
