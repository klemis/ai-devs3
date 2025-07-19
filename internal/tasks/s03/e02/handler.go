package e02

import (
	"context"
	"fmt"
	"log"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"

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
	if err != nil {
		log.Printf("Warning: failed to initialize service with Qdrant: %v", err)
		// Continue with nil service to allow graceful error handling in Execute
	}

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

	// Process the weapon reports task
	answer, err := h.service.ProcessWeaponReportsTask(apiKey)
	if err != nil {
		return fmt.Errorf("failed to process weapon reports task: %w", err)
	}

	// Submit response
	response := h.httpClient.BuildAIDevsResponse("wektory", apiKey, answer)
	result, err := h.httpClient.PostReport(ctx, h.config.AIDevs.BaseURL, response)
	if err != nil {
		return fmt.Errorf("failed to submit report: %w", err)
	}

	log.Printf("Successfully processed weapon reports, response: %s", result)
	return nil
}
