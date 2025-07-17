package e05

import (
	"context"
	"fmt"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/ollama"
	"ai-devs3/pkg/errors"
)

// Service handles the S01E05 text censoring task
type Service struct {
	httpClient   *http.Client
	ollamaClient *ollama.Client
	config       *config.Config
}

// NewService creates a new S01E05 service
func NewService(cfg *config.Config, httpClient *http.Client, ollamaClient *ollama.Client) *Service {
	return &Service{
		httpClient:   httpClient,
		ollamaClient: ollamaClient,
		config:       cfg,
	}
}

// FetchTextData fetches the raw text data from the centrala API
func (s *Service) FetchTextData(ctx context.Context, apiKey string) (*TextData, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.NewProcessingError("http", "fetch_text_data", "API key is empty", nil)
	}

	url := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/cenzura.txt", apiKey)

	content, err := s.httpClient.FetchData(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch text data: %w", err)
	}

	return &TextData{
		Content: content,
		URL:     url,
	}, nil
}

// CensorText uses Ollama to censor the text according to the rules
func (s *Service) CensorText(ctx context.Context, text string) (*CensorResponse, error) {
	if strings.TrimSpace(text) == "" {
		return nil, errors.NewProcessingError("ollama", "censor_text", "text is empty", nil)
	}

	censoredText, err := s.ollamaClient.CensorText(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("failed to censor text with Ollama: %w", err)
	}

	fmt.Println("Original text:", text)
	fmt.Println("Censored text:", censoredText)

	return &CensorResponse{
		CensoredText: censoredText,
		Success:      true,
	}, nil
}

// SubmitCensoredText submits the censored text to the centrala API
func (s *Service) SubmitCensoredText(ctx context.Context, apiKey string, censoredText string) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("CENZURA", apiKey, censoredText)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit censored text: %w", err)
	}

	return result, nil
}

// ExecuteTask executes the complete S01E05 task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	// Step 1: Fetch text data
	textData, err := s.FetchTextData(ctx, apiKey)
	if err != nil {
		return nil, errors.NewTaskError("s01e05", "fetch_text_data", err)
	}

	// Step 2: Censor the text
	censorResponse, err := s.CensorText(ctx, textData.Content)
	if err != nil {
		return nil, errors.NewTaskError("s01e05", "censor_text", err)
	}

	// Step 3: Submit censored text
	response, err := s.SubmitCensoredText(ctx, apiKey, censorResponse.CensoredText)
	if err != nil {
		return nil, errors.NewTaskError("s01e05", "submit_censored_text", err)
	}

	return &TaskResult{
		Response:     response,
		OriginalText: textData.Content,
		CensoredText: censorResponse.CensoredText,
	}, nil
}
