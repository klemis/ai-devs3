package e02

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"

	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
)

// Service handles weapon reports vector processing
type Service struct {
	httpClient     *http.Client
	llmClient      *openai.Client
	qdrantClient   *qdrant.Client
	collectionName string
}

// NewService creates a new service instance
func NewService(httpClient *http.Client, llmClient *openai.Client, quadrantClient *qdrant.Client) (*Service, error) {
	return &Service{
		httpClient:     httpClient,
		llmClient:      llmClient,
		qdrantClient:   quadrantClient,
		collectionName: "weapon_reports",
	}, nil
}

// ExecuteTask executes the complete S03E02 task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	startTime := time.Now()

	// Execute the weapon reports task
	answer, stats, err := s.processWeaponReportsTask(ctx, apiKey)
	if err != nil {
		return nil, err
	}

	// Submit response
	response, err := s.submitWeaponReportsResponse(ctx, apiKey, answer)
	if err != nil {
		return nil, errors.NewTaskError("s03e02", "submit_response", err)
	}

	// Calculate final stats
	stats.ProcessingTime = time.Since(startTime).Seconds()

	return &TaskResult{
		Response:         response,
		Answer:           answer,
		ProcessingStats:  stats,
		ReportsProcessed: stats.ReportsProcessed,
	}, nil
}

// processWeaponReportsTask processes all weapon reports and answers the query
func (s *Service) processWeaponReportsTask(ctx context.Context, apiKey string) (string, *ProcessingStats, error) {
	stats := &ProcessingStats{
		VectorDimensions: 3072, // text-embedding-3-large dimensions
	}

	// Step 1: Setup Qdrant collection
	log.Println("Setting up Qdrant collection...")
	if err := s.setupQdrantCollection(ctx); err != nil {
		return "", stats, errors.NewTaskError("s03e02", "setup_collection", err)
	}
	stats.CollectionSetup = true

	// Step 2: Process weapon reports
	log.Println("Processing weapon reports...")
	reports, err := s.loadWeaponReports()
	if err != nil {
		return "", stats, errors.NewTaskError("s03e02", "load_reports", err)
	}

	log.Printf("Found %d weapon reports", len(reports))
	stats.ReportsProcessed = len(reports)

	// Calculate total data size
	var totalSize int64
	for _, report := range reports {
		totalSize += int64(len(report.Content))
	}
	stats.TotalDataSize = totalSize

	// Step 3: Generate embeddings and store in Qdrant
	embeddingsCount, err := s.processAndStoreReports(ctx, reports)
	if err != nil {
		return "", stats, errors.NewTaskError("s03e02", "process_store_reports", err)
	}
	stats.EmbeddingsGenerated = embeddingsCount

	// Step 4: Query for theft mention
	log.Println("Searching for theft mention...")
	searchStart := time.Now()
	theftQuery := "W raporcie, z którego dnia znajduje się wzmianka o kradzieży prototypu broni?"
	date, err := s.searchForTheft(ctx, theftQuery)
	if err != nil {
		return "", stats, errors.NewTaskError("s03e02", "search_theft", err)
	}
	stats.SearchTime = time.Since(searchStart).Seconds()

	log.Printf("Found theft mention in report from date: %s", date)
	return date, stats, nil
}

// setupQdrantCollection creates or recreates the collection for weapon reports
func (s *Service) setupQdrantCollection(ctx context.Context) error {
	// Check if collection exists
	exists, err := s.qdrantClient.CollectionExists(ctx, s.collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	// Skip collection creation if it already exists
	if exists {
		log.Println("Collection already exists, skipping collection creation")
		return nil
	}

	// Create new collection with proper vector size for text-embedding-3-large
	log.Println("Creating new collection...")
	err = s.qdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: s.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     3072, // text-embedding-3-large dimensions
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	log.Println("Collection created successfully")
	return nil
}

// loadWeaponReports loads all weapon test reports from the directory
func (s *Service) loadWeaponReports() ([]WeaponReport, error) {
	reportsDir := "../lessons-md/pliki_z_fabryki/do-not-share"
	var reports []WeaponReport

	err := filepath.WalkDir(reportsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Only process .txt files
		if !strings.HasSuffix(path, ".txt") || d.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Extract date from filename
		filename := filepath.Base(path)
		date, err := s.extractDateFromFilename(filename)
		if err != nil {
			log.Printf("Warning: failed to extract date from %s: %v", filename, err)
			// Use a default date if extraction fails
			date = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		report := WeaponReport{
			ID:       uuid.New().String(),
			Date:     date,
			Filename: filename,
			Content:  string(content),
		}

		reports = append(reports, report)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk reports directory: %w", err)
	}

	return reports, nil
}

// extractDateFromFilename extracts date from filename format like "2024_01_08.txt"
func (s *Service) extractDateFromFilename(filename string) (time.Time, error) {
	// Remove .txt extension
	name := strings.TrimSuffix(filename, ".txt")

	// Split by underscore
	parts := strings.Split(name, "_")
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid filename format: %s", filename)
	}

	// Parse year, month, day
	dateStr := fmt.Sprintf("%s-%s-%s", parts[0], parts[1], parts[2])
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse date from %s: %w", dateStr, err)
	}

	return date, nil
}

// submitWeaponReportsResponse submits the weapon reports results to the centrala API
func (s *Service) submitWeaponReportsResponse(ctx context.Context, apiKey string, answer string) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("wektory", apiKey, answer)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit weapon reports response: %w", err)
	}

	return result, nil
}

// PrintProcessingStats prints detailed processing statistics
func (s *Service) PrintProcessingStats(stats *ProcessingStats) {
	fmt.Println("=== S03E02 Processing Statistics ===")
	fmt.Printf("Reports processed: %d\n", stats.ReportsProcessed)
	fmt.Printf("Embeddings generated: %d\n", stats.EmbeddingsGenerated)
	fmt.Printf("Vector dimensions: %d\n", stats.VectorDimensions)
	fmt.Printf("Total data size: %d bytes\n", stats.TotalDataSize)
	fmt.Printf("Collection setup: %t\n", stats.CollectionSetup)
	fmt.Printf("Search time: %.2f seconds\n", stats.SearchTime)
	fmt.Printf("Total processing time: %.2f seconds\n", stats.ProcessingTime)
	fmt.Println("=====================================")
}

// processAndStoreReports generates embeddings and stores reports in Qdrant
func (s *Service) processAndStoreReports(ctx context.Context, reports []WeaponReport) (int, error) {
	var points []*qdrant.PointStruct

	for i, report := range reports {
		log.Printf("Processing report %d/%d: %s", i+1, len(reports), report.Filename)

		// Generate embedding for the report content
		embedding, err := s.generateEmbedding(report.Content)
		if err != nil {
			return 0, fmt.Errorf("failed to generate embedding for %s: %w", report.Filename, err)
		}

		// Create Qdrant point
		payload := map[string]any{
			"date":     report.Date.Format("2006-01-02"),
			"filename": report.Filename,
			"content":  report.Content,
		}

		// Convert float64 to float32 for Qdrant
		embeddingFloat32 := make([]float32, len(embedding))
		for i, v := range embedding {
			embeddingFloat32[i] = float32(v)
		}

		point := &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(uint64(i + 1)),
			Vectors: qdrant.NewVectors(embeddingFloat32...),
			Payload: qdrant.NewValueMap(payload),
		}

		points = append(points, point)
	}

	// Upsert all points to Qdrant
	log.Println("Storing points in Qdrant...")
	_, err := s.qdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collectionName,
		Points:         points,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to upsert points to Qdrant: %w", err)
	}

	log.Printf("Successfully stored %d reports in Qdrant", len(points))
	return len(points), nil
}

// generateEmbedding generates embedding for text using OpenAI
func (s *Service) generateEmbedding(text string) ([]float64, error) {
	prompt := openai.EmbeddingRequest{
		Input: text,
		Model: "text-embedding-3-large",
	}

	response, err := s.llmClient.CreateEmbedding(prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data received")
	}

	return response.Data[0].Embedding, nil
}

// searchForTheft searches for reports mentioning theft and returns the date
func (s *Service) searchForTheft(ctx context.Context, query string) (string, error) {
	// Generate embedding for the query
	queryEmbedding, err := s.generateEmbedding(query)
	if err != nil {
		return "", fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Convert float64 to float32 for Qdrant
	queryEmbeddingFloat32 := make([]float32, len(queryEmbedding))
	for i, v := range queryEmbedding {
		queryEmbeddingFloat32[i] = float32(v)
	}

	// Search in Qdrant using Query method
	searchResult, err := s.qdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.collectionName,
		Query:          qdrant.NewQuery(queryEmbeddingFloat32...),
		Limit:          qdrant.PtrOf(uint64(1)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return "", fmt.Errorf("failed to search in Qdrant: %w", err)
	}

	if len(searchResult) == 0 {
		return "", fmt.Errorf("no results found for query")
	}

	// Extract date from the most similar result
	result := searchResult[0]
	payload := result.GetPayload()

	dateValue, exists := payload["date"]
	if !exists {
		return "", fmt.Errorf("date not found in result payload")
	}

	date := dateValue.GetStringValue()
	if date == "" {
		return "", fmt.Errorf("invalid date in result payload")
	}

	log.Printf("Found most relevant result with score: %f", result.GetScore())
	log.Printf("Report filename: %s", payload["filename"].GetStringValue())

	return date, nil
}
