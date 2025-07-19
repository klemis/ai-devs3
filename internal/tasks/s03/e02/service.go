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

// ProcessWeaponReportsTask processes all weapon reports and answers the query
func (s *Service) ProcessWeaponReportsTask(apiKey string) (string, error) {
	ctx := context.Background()

	// Step 1: Setup Qdrant collection
	log.Println("Setting up Qdrant collection...")
	if err := s.setupQdrantCollection(ctx); err != nil {
		return "", fmt.Errorf("failed to setup Qdrant collection: %w", err)
	}

	// Step 2: Process weapon reports
	log.Println("Processing weapon reports...")
	reports, err := s.loadWeaponReports()
	if err != nil {
		return "", fmt.Errorf("failed to load weapon reports: %w", err)
	}

	log.Printf("Found %d weapon reports", len(reports))

	// Step 3: Generate embeddings and store in Qdrant
	if err := s.processAndStoreReports(ctx, reports); err != nil {
		return "", fmt.Errorf("failed to process and store reports: %w", err)
	}

	// Step 4: Query for theft mention
	log.Println("Searching for theft mention...")
	theftQuery := "W raporcie, z którego dnia znajduje się wzmianka o kradzieży prototypu broni?"
	date, err := s.searchForTheft(ctx, theftQuery)
	if err != nil {
		return "", fmt.Errorf("failed to search for theft: %w", err)
	}

	log.Printf("Found theft mention in report from date: %s", date)
	return date, nil
}

// setupQdrantCollection creates or recreates the collection for weapon reports
func (s *Service) setupQdrantCollection(ctx context.Context) error {
	// Check if collection exists
	exists, err := s.qdrantClient.CollectionExists(ctx, s.collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	// Delete existing collection if it exists
	if exists {
		log.Println("Deleting existing collection...")
		if err := s.qdrantClient.DeleteCollection(ctx, s.collectionName); err != nil {
			return fmt.Errorf("failed to delete existing collection: %w", err)
		}
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

// processAndStoreReports generates embeddings and stores reports in Qdrant
func (s *Service) processAndStoreReports(ctx context.Context, reports []WeaponReport) error {
	var points []*qdrant.PointStruct

	for i, report := range reports {
		log.Printf("Processing report %d/%d: %s", i+1, len(reports), report.Filename)

		// Generate embedding for the report content
		embedding, err := s.generateEmbedding(report.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding for %s: %w", report.Filename, err)
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

		// Add small delay to respect API rate limits
		time.Sleep(100 * time.Millisecond)
	}

	// Upsert all points to Qdrant
	log.Println("Storing points in Qdrant...")
	_, err := s.qdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collectionName,
		Points:         points,
	})
	if err != nil {
		return fmt.Errorf("failed to upsert points to Qdrant: %w", err)
	}

	log.Printf("Successfully stored %d reports in Qdrant", len(points))
	return nil
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
