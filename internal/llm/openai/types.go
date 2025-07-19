package openai

import (
	"encoding/json"
	"fmt"
)

// CategorizationResult represents the LLM's categorization response
type CategorizationResult struct {
	Thinking      string `json:"_thinking"`
	Category      string `json:"category"`
	Justification string `json:"justification"`
}

// RoboISOAnswer represents an answer for RoboISO protocol
type RoboISOAnswer struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

// parseJSONResponse parses a JSON response from OpenAI into the target structure
func parseJSONResponse(content string, target interface{}) error {
	if err := json.Unmarshal([]byte(content), target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w, content: %s", err, content)
	}
	return nil
}

// ImageProcessingResult represents processed image metadata
type ImageProcessingResult struct {
	Base64Data string
	Width      int
	Height     int
	TokenCost  int
}

// EmbeddingRequest represents a request to generate embeddings
type EmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// EmbeddingResponse represents the response from embedding API
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
}
