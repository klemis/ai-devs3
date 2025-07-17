package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"ai-devs3/internal/config"
	"ai-devs3/pkg/errors"
)

// Client wraps Ollama client with configuration and error handling
type Client struct {
	client  *http.Client
	baseURL string
	config  config.OllamaConfig
}

// ChatMessage represents a single message in the chat
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
}

// ChatRequest represents the chat request structure for Ollama API
type ChatRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage  `json:"messages"`
	Stream   bool           `json:"stream"`
	Options  map[string]any `json:"options"`
}

// ChatResponse represents the response structure from Ollama chat API
type ChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

// NewClient creates a new Ollama client with the given configuration
func NewClient(cfg config.OllamaConfig) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: cfg.BaseURL,
		config:  cfg,
	}
}

// CensorText sends text to Ollama for censoring according to specific rules
func (c *Client) CensorText(ctx context.Context, text string) (string, error) {
	systemPrompt := `You are a text replacement tool. Your ONLY task is to find and replace specific phrases with "CENZURA" while keeping ALL other text EXACTLY unchanged.

	REPLACEMENT RULES:
		1. Name + surname together → "CENZURA" (e.g., "Jan Nowak" → "CENZURA")
		2. Age numbers → "CENZURA" (e.g., "32" → "CENZURA")
		3. City names → "CENZURA" (e.g., "Katowice" → "CENZURA")
		4. Street addresses: Keep "ul." or "ulicy" or "przy ul." but replace everything after with "CENZURA"
		   - "ul. Różanej 12" → "ul. CENZURA"
		   - "przy ul. Różanej 12" → "przy ul. CENZURA"
		   - "ulicy Pięknej 5" → "ulicy CENZURA"

	CRITICAL REQUIREMENTS:
		- Do NOT change, rephrase, or improve ANY other text
		- Do NOT remove prepositions like "przy" or "ul."
		- Keep EXACT original wording except for the specified replacements
		- Maintain ALL punctuation, spacing, and capitalization

	Examples:
		Input: "Podejrzany: Jan Kowalski. Mieszka w Warszawie przy ul. Długiej 5. Ma 25 lat."
		Output: "Podejrzany: CENZURA. Mieszka w CENZURA przy ul. CENZURA. Ma CENZURA lat."

		Input: "Adam Nowak zamieszkały w Krakowie, ulicy Królewskiej 10, wiek 45 lat."
		Output: "CENZURA zamieszkały w CENZURA, ulicy CENZURA, wiek CENZURA lat."

		Input: Dane podejrzanego: Jakub Woźniak. Adres: Rzeszów, ul. Miła 4. Wiek: 33 lata.
		Input: Dane podejrzanego: CENZURA. Adres: Cenzura, ul. CENZURA. Wiek: CENZURA lat.

		Now perform ONLY the specified replacements on the following text:`

	userPrompt := fmt.Sprintf("%s", text)

	request := ChatRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Stream: false,
		Options: map[string]any{
			"temperature": c.config.Temperature, //0.0
			"top_p":       0.1,
			"num_predict": 200,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("Ollama", 0, "failed to call Ollama API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.NewAPIError("Ollama", resp.StatusCode, "Ollama API error", nil)
	}

	var response ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Message.Content, nil
}

// GetAnswer sends a simple question to Ollama and returns the response
func (c *Client) GetAnswer(ctx context.Context, question string) (string, error) {
	request := ChatRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: question,
			},
		},
		Stream: false,
		Options: map[string]any{
			"temperature": c.config.Temperature,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("Ollama", 0, "failed to call Ollama API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.NewAPIError("Ollama", resp.StatusCode, "Ollama API error", nil)
	}

	var response ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Message.Content, nil
}

// GetAnswerWithContext provides a generic way to ask questions with custom system and user prompts
func (c *Client) GetAnswerWithContext(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	request := ChatRequest{
		Model: c.config.Model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Stream: false,
		Options: map[string]any{
			"temperature": c.config.Temperature,
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("Ollama", 0, "failed to call Ollama API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.NewAPIError("Ollama", resp.StatusCode, "Ollama API error", nil)
	}

	var response ChatResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Message.Content, nil
}
