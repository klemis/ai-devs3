package e02

import (
	"time"

	"github.com/google/uuid"
)

// WeaponReport represents a weapon test report document
type WeaponReport struct {
	ID       string    `json:"id"`
	Date     time.Time `json:"date"`
	Filename string    `json:"filename"`
	Content  string    `json:"content"`
}

// ReportEmbedding represents a report with its vector embedding
type ReportEmbedding struct {
	Report    WeaponReport `json:"report"`
	Embedding []float64    `json:"embedding"`
}

// QdrantPoint represents a point to be stored in Qdrant
type QdrantPoint struct {
	ID      uuid.UUID      `json:"id"`
	Vector  []float64      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

// SearchResult represents a search result from Qdrant
type SearchResult struct {
	ID      string         `json:"id"`
	Score   float64        `json:"score"`
	Payload map[string]any `json:"payload"`
}

// VektorAnswer represents the answer structure for the wektory task
type VektorAnswer string

// TaskResult represents the final result of the S03E02 task
type TaskResult struct {
	Response         string
	Answer           string
	ProcessingStats  *ProcessingStats
	ReportsProcessed int
}

// ProcessingStats represents statistics about the vector processing
type ProcessingStats struct {
	ReportsProcessed    int
	EmbeddingsGenerated int
	ProcessingTime      float64
	CollectionSetup     bool
	SearchTime          float64
	VectorDimensions    int
	TotalDataSize       int64
}

// EmbeddingResponse represents OpenAI embedding API response
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}
