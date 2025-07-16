package e01

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the S01E01 robot authentication task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
	config     *config.Config
}

// NewService creates a new S01E01 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
		config:     cfg,
	}
}

// ExtractQuestion extracts a question from HTML content using regex
func (s *Service) ExtractQuestion(ctx context.Context, htmlContent string) (*Question, error) {
	re := regexp.MustCompile(`<p id="human-question">Question:<br />(.*?)</p>`)
	match := re.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return nil, errors.NewProcessingError("html", "question extraction", "question not found in HTML content", nil)
	}

	return &Question{Text: match[1]}, nil
}

// GetAnswer gets an answer from the LLM for the given question
func (s *Service) GetAnswer(ctx context.Context, question *Question) (*Answer, error) {
	if question == nil || strings.TrimSpace(question.Text) == "" {
		return nil, errors.NewProcessingError("llm", "question answering", "question is empty", nil)
	}

	answer, err := s.llmClient.GetAnswer(ctx, question.Text)
	if err != nil {
		return nil, fmt.Errorf("failed to get answer from LLM: %w", err)
	}

	return &Answer{Text: answer}, nil
}

// SubmitLogin submits the login form with credentials and answer
func (s *Service) SubmitLogin(ctx context.Context, loginURL string, creds *Credentials, answer *Answer) (*LoginResponse, error) {
	if creds == nil {
		return nil, errors.NewProcessingError("auth", "login", "credentials are nil", nil)
	}

	if answer == nil {
		return nil, errors.NewProcessingError("auth", "login", "answer is nil", nil)
	}

	// Prepare form data
	formData := map[string]string{
		"username": creds.Username,
		"password": creds.Password,
		"answer":   answer.Text,
	}

	// Submit the form
	response, err := s.httpClient.PostForm(ctx, loginURL, formData)
	if err != nil {
		return nil, fmt.Errorf("failed to submit login form: %w", err)
	}
	fmt.Println(response)

	return &LoginResponse{
		Content: response,
	}, nil
}

// ExtractFlag extracts flag from login response content
func (s *Service) ExtractFlag(ctx context.Context, content string) (string, error) {
	// Use the LLM to find the flag
	flag, err := s.llmClient.FindFlag(ctx, content)
	if err != nil {
		return "", fmt.Errorf("failed to extract flag: %w", err)
	}

	return flag, nil
}

// ExecuteTask executes the complete S01E01 task workflow
func (s *Service) ExecuteTask(ctx context.Context, loginURL string, creds *Credentials) (*TaskResult, error) {
	// Step 1: Fetch the login page
	htmlContent, err := s.httpClient.FetchPage(ctx, loginURL)
	if err != nil {
		return nil, errors.NewTaskError("s01e01", "fetch_page", err)
	}

	// Step 2: Extract the question
	question, err := s.ExtractQuestion(ctx, htmlContent)
	if err != nil {
		return nil, errors.NewTaskError("s01e01", "extract_question", err)
	}
	fmt.Println(question)

	// Step 3: Get answer from LLM
	answer, err := s.GetAnswer(ctx, question)
	if err != nil {
		return nil, errors.NewTaskError("s01e01", "get_answer", err)
	}
	fmt.Println(answer)

	// Step 4: Submit login form
	loginResponse, err := s.SubmitLogin(ctx, loginURL, creds, answer)
	if err != nil {
		return nil, errors.NewTaskError("s01e01", "submit_login", err)
	}

	// Step 5: Extract flag from response
	flag, err := s.ExtractFlag(ctx, loginResponse.Content)
	if err != nil {
		return nil, errors.NewTaskError("s01e01", "extract_flag", err)
	}

	return &TaskResult{
		Flag:    flag,
		Content: loginResponse.Content,
	}, nil
}
