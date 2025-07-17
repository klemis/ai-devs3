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
	Model    string                 `json:"model"`
	Messages []ChatMessage          `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options"`
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
	systemPrompt := `
You are a text formatting assistant. Your task is to replace certain phrases in the provided text with the word "CENZURA" according to these strict rules:

RULES:
1. Replace name and surname TOGETHER as one unit with "CENZURA" (e.g., "Jan Nowak" -> "CENZURA")
2. Replace age numbers with "CENZURA" (e.g., "32" -> "CENZURA")
3. Replace city names with "CENZURA" (e.g., "Katowice" or "Katowicach" -> "CENZURA")
4. Replace the entire street name and house number TOGETHER as one unit with "CENZURA" after "ul." or "ulicy" or "przy ulicy" (e.g., "ul. Tuwima 10" or "ulicy Pięknej 5" -> "ul. CENZURA" or "ulicy CENZURA" or "przy ulicy CENZURA")

IMPORTANT RESTRICTIONS:
- DO NOT censor name and surname separately (never use "CENZURA CENZURA")
- DO NOT censor street and house number separately (never use "CENZURA CENZURA")
- Always use "CENZURA" exactly as written. Do NOT change its form, ending, or capitalization.
- Keep ALL original formatting (dots, commas, spaces, capitalization)
- Do NOT change any text except for the specified information to be replaced.
- Do NOT add any explanations or comments
- Return ONLY the modified text

Input: "Informacje o podejrzanym: Adam Nowak. Mieszka w Katowicach przy ulicy Tuwima 10. Wiek: 32 lata."
Output: "Informacje o podejrzanym: CENZURA. Mieszka w CENZURA przy ulicy CENZURA. Wiek: CENZURA lata."
`
	userPrompt := fmt.Sprintf("Replace the specified phrases in this text according to the rules:\n\n%s", text)

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
		Options: map[string]interface{}{
			"temperature": c.config.Temperature,
			"top_p":       0.9,
			"num_predict": 500,
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
		Options: map[string]interface{}{
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
		Options: map[string]interface{}{
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
