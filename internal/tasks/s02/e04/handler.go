package e04

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/internal/storage/cache"
	pkgerrors "ai-devs3/pkg/errors"
)

// Handler orchestrates the S02E04 file categorization task
type Handler struct {
	service *Service
	config  *config.Config
}

// NewHandler creates a new S02E04 handler
func NewHandler(cfg *config.Config) *Handler {
	// Initialize dependencies
	httpClient := http.NewClient(cfg.HTTP)
	llmClient := openai.NewClient(cfg.OpenAI)

	// Create cache
	fileCache, err := cache.NewFileCache(cfg.Cache)
	if err != nil {
		log.Printf("Failed to create file cache: %v", err)
		return nil
	}
	taskCache := cache.NewTaskCache(fileCache, "s02e04")

	// Create service
	service := NewService(cfg, httpClient, llmClient, taskCache)

	return &Handler{
		service: service,
		config:  cfg,
	}
}

// Execute runs the S02E04 task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S02E04 file categorization task")

	// Get API key from environment
	apiKey := os.Getenv("AI_DEVS_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("AI_DEVS_API_KEY environment variable not set")
	}

	// Default files directory
	filesDir := "../lessons-md/pliki_z_fabryki"

	// Check if files directory exists
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		return fmt.Errorf("files directory not found: %s", filesDir)
	}

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, filesDir, apiKey)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S02E04 task failed: %w", err)
		}
		return fmt.Errorf("S02E04 task failed: %w", err)
	}

	// Print processing statistics
	h.service.PrintProcessingStats(result.ProcessingStats)

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Success: %t", result.Success)
	log.Printf("Total files processed: %d", result.TotalFiles)
	log.Printf("Files categorized: %d", result.CategorizedCount)
	log.Printf("People files: %d", len(result.CategorizedFiles.People))
	log.Printf("Hardware files: %d", len(result.CategorizedFiles.Hardware))

	fmt.Println("=== Categorization Results ===")
	fmt.Printf("People files (%d):\n", len(result.CategorizedFiles.People))
	for _, filename := range result.CategorizedFiles.People {
		fmt.Printf("  - %s\n", filename)
	}
	fmt.Printf("\nHardware files (%d):\n", len(result.CategorizedFiles.Hardware))
	for _, filename := range result.CategorizedFiles.Hardware {
		fmt.Printf("  - %s\n", filename)
	}
	fmt.Println("==============================")

	if result.Success {
		fmt.Println("File categorization successful!")
		fmt.Printf("Response: %s\n", result.Response)
	} else {
		fmt.Println("File categorization completed but may need manual review:")
		fmt.Printf("Response: %s\n", result.Response)
	}

	return nil
}
