package e03

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the S01E03 JSON data processing task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
	config     *config.Config
}

// NewService creates a new S01E03 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
		config:     cfg,
	}
}

// FetchTestData fetches the JSON test data from the centrala API
func (s *Service) FetchTestData(ctx context.Context, apiKey string) (map[string]any, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.NewProcessingError("http", "fetch_test_data", "API key is empty", nil)
	}

	url := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/json.txt", apiKey)

	data, err := s.httpClient.FetchJSONData(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch test data: %w", err)
	}

	return data, nil
}

// ProcessTestData processes the test data by solving math problems and getting LLM answers
func (s *Service) ProcessTestData(ctx context.Context, data map[string]any) (map[string]any, error) {
	corrected := make(map[string]any)

	for k, v := range data {
		if k == "test-data" {
			arr, ok := v.([]any)
			if !ok {
				return nil, errors.NewProcessingError("data", "process_test_data", "test-data is not an array", nil)
			}

			// Process the array
			processedArr, err := s.processTestArray(ctx, arr)
			if err != nil {
				return nil, fmt.Errorf("failed to process test array: %w", err)
			}

			corrected[k] = processedArr
		} else {
			corrected[k] = v
		}
	}

	return corrected, nil
}

// processTestArray processes individual test items in the array
func (s *Service) processTestArray(ctx context.Context, arr []any) ([]any, error) {
	// First pass: collect LLM questions and solve math problems
	llmQuestions, llmIndexes := s.collectLLMQuestions(arr)

	// Get LLM answers for all questions at once
	var answers []string
	if len(llmQuestions) > 0 {
		var err error
		answers, err = s.llmClient.GetMultipleAnswers(ctx, llmQuestions)
		if err != nil {
			return nil, fmt.Errorf("failed to get LLM answers: %w", err)
		}
	}

	// Second pass: update test answers
	arr = s.updateTestAnswers(arr, llmIndexes, answers)

	return arr, nil
}

// collectLLMQuestions gathers all LLM questions and their indexes while solving math problems
func (s *Service) collectLLMQuestions(arr []any) ([]string, []int) {
	var llmQuestions []string
	var llmIndexes []int

	for i, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Handle direct math questions
		if q, ok := m["question"].(string); ok && s.isMath(q) {
			m["answer"] = s.solveMath(q)
		}

		// Handle test sub-questions
		if test, ok := m["test"].(map[string]any); ok {
			if tq, ok := test["q"].(string); ok {
				llmQuestions = append(llmQuestions, tq)
				llmIndexes = append(llmIndexes, i)
			}
		}

		arr[i] = m
	}

	return llmQuestions, llmIndexes
}

// updateTestAnswers updates the 'test.a' field in arr with LLM answers
func (s *Service) updateTestAnswers(arr []any, llmIndexes []int, answers []string) []any {
	for idx, arrIdx := range llmIndexes {
		m, ok := arr[arrIdx].(map[string]any)
		if !ok {
			continue
		}

		test, ok := m["test"].(map[string]any)
		if !ok {
			continue
		}

		if idx < len(answers) {
			test["a"] = answers[idx]
			m["test"] = test
			arr[arrIdx] = m
		}
	}

	return arr
}

// isMath checks if the question is a simple math operation
func (s *Service) isMath(q string) bool {
	// Matches expressions like '18 + 36', '7-2', '100 * 2', '50 / 5'
	matched, _ := regexp.MatchString(`^\s*\d+\s*[-+*/]\s*\d+\s*$`, q)
	return matched
}

// solveMath solves a simple math question string like '18 + 36'
func (s *Service) solveMath(q string) int {
	re := regexp.MustCompile(`(\d+)\s*([-+*/])\s*(\d+)`)
	matches := re.FindStringSubmatch(q)
	if len(matches) != 4 {
		return 0 // or handle error
	}

	a, _ := strconv.Atoi(matches[1])
	b, _ := strconv.Atoi(matches[3])

	switch matches[2] {
	case "+":
		return a + b
	case "-":
		return a - b
	case "*":
		return a * b
	case "/":
		if b != 0 {
			return a / b
		}
	}
	return 0
}

// SubmitAnswer submits the processed answer to the centrala API
func (s *Service) SubmitAnswer(ctx context.Context, apiKey string, processedData map[string]any) (string, error) {
	answer := map[string]any{
		"apikey":    apiKey,
		"test-data": processedData["test-data"],
	}

	response := s.httpClient.BuildAIDevsResponse("JSON", apiKey, answer)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit answer: %w", err)
	}

	return result, nil
}

// ExecuteTask executes the complete S01E03 task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	// Step 1: Fetch test data
	data, err := s.FetchTestData(ctx, apiKey)
	if err != nil {
		return nil, errors.NewTaskError("s01e03", "fetch_test_data", err)
	}

	// Step 2: Process test data
	processedData, err := s.ProcessTestData(ctx, data)
	if err != nil {
		return nil, errors.NewTaskError("s01e03", "process_test_data", err)
	}

	// Step 3: Submit answer
	response, err := s.SubmitAnswer(ctx, apiKey, processedData)
	if err != nil {
		return nil, errors.NewTaskError("s01e03", "submit_answer", err)
	}

	// Count corrections made
	corrected := s.countCorrections(processedData)

	return &TaskResult{
		Response:    response,
		Corrected:   corrected,
		LLMAnswers:  s.countLLMAnswers(processedData),
		MathAnswers: s.countMathAnswers(processedData),
	}, nil
}

// countCorrections counts the number of corrections made
func (s *Service) countCorrections(data map[string]any) int {
	count := 0
	if testData, ok := data["test-data"].([]any); ok {
		for _, item := range testData {
			if m, ok := item.(map[string]any); ok {
				if _, hasAnswer := m["answer"]; hasAnswer {
					count++
				}
				if test, ok := m["test"].(map[string]any); ok {
					if _, hasTestAnswer := test["a"]; hasTestAnswer {
						count++
					}
				}
			}
		}
	}
	return count
}

// countLLMAnswers counts the number of LLM answers provided
func (s *Service) countLLMAnswers(data map[string]any) int {
	count := 0
	if testData, ok := data["test-data"].([]any); ok {
		for _, item := range testData {
			if m, ok := item.(map[string]any); ok {
				if test, ok := m["test"].(map[string]any); ok {
					if _, hasTestAnswer := test["a"]; hasTestAnswer {
						count++
					}
				}
			}
		}
	}
	return count
}

// countMathAnswers counts the number of math answers calculated
func (s *Service) countMathAnswers(data map[string]any) int {
	count := 0
	if testData, ok := data["test-data"].([]any); ok {
		for _, item := range testData {
			if m, ok := item.(map[string]any); ok {
				if q, ok := m["question"].(string); ok && s.isMath(q) {
					if _, hasAnswer := m["answer"]; hasAnswer {
						count++
					}
				}
			}
		}
	}
	return count
}
