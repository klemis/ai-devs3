package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OllamaClient implements LLM functionality using local Ollama server
type OllamaClient struct {
	BaseURL string
}

// OllamaChatMessage represents a single message in the chat
type OllamaChatMessage struct {
	Role    string `json:"role"` // "system", "user", or "assistant"
	Content string `json:"content"`
}

// OllamaChatRequest represents the chat request structure for Ollama API
type OllamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaChatMessage    `json:"messages"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options"`
}

// OllamaChatResponse represents the response structure from Ollama chat API
type OllamaChatResponse struct {
	Message OllamaChatMessage `json:"message"`
	Done    bool              `json:"done"`
}

// NewOllamaClient creates a new Ollama client with default localhost URL
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{
		BaseURL: "http://localhost:11434",
	}
}

// CensorText sends text to Ollama for censoring according to specific rules
func (c *OllamaClient) CensorText(text string) (string, error) {
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

	request := OllamaChatRequest{
		Model: "llama3.2",
		Messages: []OllamaChatMessage{
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
			"temperature": 0.1, // Make it more deterministic
			"top_p":       0.9,
			"num_predict": 500, // Allow more tokens for longer responses
		},
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Debug: Print the request
	// fmt.Printf("=== Ollama Request ===\n%s\n", string(jsonData))

	resp, err := http.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var response OllamaChatResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Message.Content, nil
}
