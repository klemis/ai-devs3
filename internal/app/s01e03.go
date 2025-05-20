package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

func (app *App) RunS01E03(apiKey string) (string, error) {
	url := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/json.txt", apiKey)

	data, err := app.pageFetcher.FetchJSONData(url)
	if err != nil {
		log.Fatalf("Failed to fetch JSON: %v", err)
	}

	corrected := app.processTestData(data)
	response := buildResponse(apiKey, corrected)

	return postReport(response), nil
}

// processTestData corrects math answers and fills LLM answers for test questions
func (app *App) processTestData(data map[string]interface{}) map[string]interface{} {
	corrected := map[string]interface{}{}
	for k, v := range data {
		if k == "test-data" {
			arr, ok := v.([]interface{})
			if !ok {
				log.Fatalf("test-data is not an array")
			}
			llmQuestions, llmIndexes := collectLLMQuestions(arr)
			answers := app.getLLMAnswers(llmQuestions)
			arr = updateTestAnswers(arr, llmIndexes, answers)
			corrected[k] = arr
		} else {
			corrected[k] = v
		}
	}

	return corrected
}

// collectLLMQuestions gathers all LLM questions and their indexes
func collectLLMQuestions(arr []interface{}) ([]string, []int) {
	var llmQuestions []string
	var llmIndexes []int
	for i, item := range arr {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if q, ok := m["question"].(string); ok && isMath(q) {
			m["answer"] = solveMath(q)
		}
		if test, ok := m["test"].(map[string]interface{}); ok {
			if tq, ok := test["q"].(string); ok {
				llmQuestions = append(llmQuestions, tq)
				llmIndexes = append(llmIndexes, i)
			}
		}
		arr[i] = m
	}

	return llmQuestions, llmIndexes
}

// getLLMAnswers asks the LLM all questions in one batch
func (app *App) getLLMAnswers(questions []string) []string {
	if len(questions) == 0 {
		return nil
	}

	answers, err := app.llmClient.GetMultipleAnswers(questions)
	if err != nil {
		log.Fatalf("Error from LLM: %v", err)
	}

	return answers
}

// updateTestAnswers updates the 'test.a' field in arr with LLM answers
func updateTestAnswers(arr []interface{}, llmIndexes []int, answers []string) []interface{} {
	for idx, arrIdx := range llmIndexes {
		m, ok := arr[arrIdx].(map[string]interface{})
		if !ok {
			continue
		}
		test, ok := m["test"].(map[string]interface{})
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

// buildResponse creates the final response object
func buildResponse(apiKey string, corrected map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"task":   "JSON",
		"apikey": apiKey,
		"answer": map[string]interface{}{
			"apikey":    apiKey,
			"test-data": corrected["test-data"],
		},
	}
}

// postReport sends the response to the report endpoint
func postReport(response map[string]interface{}) string {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(response)
	if err != nil {
		log.Fatalf("Failed to encode response: %v", err)
	}

	postResp, err := http.Post("https://c3ntrala.ag3nts.org/report", "application/json", buf)
	if err != nil {
		log.Fatalf("Failed to POST report: %v", err)
	}
	defer postResp.Body.Close()

	postBody, _ := io.ReadAll(postResp.Body)

	return string(postBody)
}

// isMath checks if the question is a simple math operation
func isMath(q string) bool {
	// Matches expressions like '18 + 36', '7-2', '100 * 2', '50 / 5'
	matched, _ := regexp.MatchString(`^\s*\d+\s*[-+*/]\s*\d+\s*$`, q)
	return matched
}

// solveMath solves a simple math question string like '18 + 36'
func solveMath(q string) int {
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
