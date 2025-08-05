package e05

import (
	"context"
	"errors"
	"fmt"
	"log"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/neo4j"
	pkgerrors "ai-devs3/pkg/errors"
)

// Handler handles the S03E05 task execution
type Handler struct {
	config      *config.Config
	httpClient  *http.Client
	neo4jClient *neo4j.Client
	service     *Service
}

// NewHandler creates a new handler instance
func NewHandler(cfg *config.Config) *Handler {
	httpClient := http.NewClient(cfg.HTTP)

	// Neo4j client will be created in Execute to handle potential connection errors
	service := NewService(httpClient)

	return &Handler{
		config:     cfg,
		httpClient: httpClient,
		service:    service,
	}
}

// Execute runs the S03E05 connections task
func (h *Handler) Execute(ctx context.Context) error {
	log.Println("Starting S03E05 connections task (Neo4j graph database)")

	// Get API key from environment
	apiKey := h.config.AIDevs.APIKey
	if apiKey == "" {
		return fmt.Errorf("AI_DEVS_API_KEY is required")
	}

	// Create Neo4j client
	neo4jClient, err := neo4j.NewClient(&h.config.Neo4j)
	if err != nil {
		return fmt.Errorf("failed to create Neo4j client: %w", err)
	}
	defer func() {
		if closeErr := neo4jClient.Close(ctx); closeErr != nil {
			log.Printf("Warning: failed to close Neo4j connection: %v", closeErr)
		}
	}()

	// Verify Neo4j connectivity
	if err := neo4jClient.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("failed to connect to Neo4j: %w", err)
	}
	log.Println("Neo4j connection verified successfully")

	// Set Neo4j client in service
	h.service.SetNeo4jClient(neo4jClient)

	// Execute the task
	result, err := h.service.ExecuteTask(ctx, apiKey)
	if err != nil {
		var taskErr pkgerrors.TaskError
		if errors.As(err, &taskErr) {
			log.Printf("Task failed at step %s: %v", taskErr.Step, taskErr.Err)
			return fmt.Errorf("S03E05 task failed: %w", err)
		}
		return fmt.Errorf("S03E05 task failed: %w", err)
	}

	// Log results
	log.Printf("Task completed successfully!")
	log.Printf("Shortest path found: %v", result.ShortestPath)
	log.Printf("Path string: %s", result.PathString)
	log.Printf("Users loaded: %d", result.Stats.UsersLoaded)
	log.Printf("Connections loaded: %d", result.Stats.ConnectionsLoaded)
	log.Printf("Neo4j nodes created: %d", result.Stats.NodesCreated)
	log.Printf("Neo4j relationships created: %d", result.Stats.RelationshipsCreated)
	log.Printf("Processing time: %.2f seconds", result.ProcessingTime)

	fmt.Println("=== Connections Task Results ===")
	fmt.Printf("Shortest Path: %v\n", result.ShortestPath)
	fmt.Printf("Path String: %s\n", result.PathString)
	fmt.Printf("Path Length: %d steps\n", len(result.ShortestPath)-1)
	fmt.Println()
	fmt.Println("=== Processing Statistics ===")
	fmt.Printf("Users loaded from MySQL: %d\n", result.Stats.UsersLoaded)
	fmt.Printf("Connections loaded from MySQL: %d\n", result.Stats.ConnectionsLoaded)
	fmt.Printf("Neo4j nodes created: %d\n", result.Stats.NodesCreated)
	fmt.Printf("Neo4j relationships created: %d\n", result.Stats.RelationshipsCreated)
	fmt.Printf("Processing time: %.2f seconds\n", result.ProcessingTime)
	fmt.Println("================================")
	fmt.Println("Connections task successful!")
	fmt.Printf("Response: %s\n", result.Response)

	return nil
}
