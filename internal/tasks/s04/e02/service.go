package e02

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	pkgerrors "ai-devs3/pkg/errors"
)

// Service handles the S04E02 task execution
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
}

// NewService creates a new service instance
func NewService(httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
	}
}

// ExecuteTask executes the complete S04E02 classification task
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	log.Println("Starting S04E02 text classification task")

	// Read verification lines
	lines, err := s.readVerifyLines()
	if err != nil {
		return nil, pkgerrors.NewTaskError("s04e02", "read_verify_lines", err)
	}

	log.Printf("Read %d lines for verification", len(lines))

	// Process each line and collect correct answers
	var correctAnswers []string
	correctCount := 0

	for i, line := range lines {
		lineID := fmt.Sprintf("%02d", i+1)
		log.Printf("Processing line %s: %s", lineID, line)

		classification, err := s.classifyLine(ctx, line)
		if err != nil {
			log.Printf("Warning: failed to classify line %s: %v", lineID, err)
			continue
		}

		if classification == ClassificationReliable {
			correctAnswers = append(correctAnswers, lineID)
			correctCount++
			log.Printf("Line %s classified as RELIABLE", lineID)
		} else {
			log.Printf("Line %s classified as UNRELIABLE", lineID)
		}
	}

	log.Printf("Classification complete. Found %d reliable lines out of %d total", correctCount, len(lines))

	// Submit final response using the standard pattern
	response, err := s.submitFinalResponse(ctx, apiKey, correctAnswers)
	if err != nil {
		return nil, pkgerrors.NewTaskError("s04e02", "submit_final_response", err)
	}

	return &TaskResult{
		CorrectAnswers: correctAnswers,
		TotalLines:     len(lines),
		CorrectCount:   correctCount,
		Response:       response,
	}, nil
}

// readVerifyLines reads lines from the verify.txt file
func (s *Service) readVerifyLines() ([]string, error) {
	// Construct path to verify.txt
	verifyPath := filepath.Join("data", "s04e02", "verify.txt")

	file, err := os.Open(verifyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open verify file: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Extract the content part after the ID (e.g., "01=kanto,saka,brunn" -> "kanto,saka,brunn")
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				content := strings.TrimSpace(parts[1])
				lines = append(lines, content)
			}
		} else {
			// If no "=" found, use the whole line
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading verify file: %w", err)
	}

	return lines, nil
}

// classifyLine classifies a single line using the fine-tuned model
func (s *Service) classifyLine(ctx context.Context, line string) (int, error) {
	// Use the fine-tuned model for classification
	response, err := s.llmClient.ClassifyWithFineTunedModel(
		ctx,
		SystemPrompt,
		line,
		"ft:gpt-4o-mini-2024-07-18:personal:validate:C7MNVVbk",
	)
	if err != nil {
		return ClassificationUnreliable, fmt.Errorf("failed to classify line: %w", err)
	}

	// Parse the response - should be just "0" or "1"
	response = strings.TrimSpace(response)
	classification, err := strconv.Atoi(response)
	if err != nil {
		log.Printf("Warning: unexpected classification response '%s', treating as unreliable", response)
		return ClassificationUnreliable, nil
	}

	// Validate classification value
	if classification != ClassificationReliable && classification != ClassificationUnreliable {
		log.Printf("Warning: invalid classification value %d, treating as unreliable", classification)
		return ClassificationUnreliable, nil
	}

	return classification, nil
}

// submitFinalResponse submits the classification results using the standard pattern
func (s *Service) submitFinalResponse(ctx context.Context, apiKey string, correctAnswers []string) (string, error) {
	request := s.httpClient.BuildAIDevsResponse(TaskName, apiKey, correctAnswers)

	response, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", request)
	if err != nil {
		return "", fmt.Errorf("failed to submit final response: %w", err)
	}

	return response, nil
}
