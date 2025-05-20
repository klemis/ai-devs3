package app

import (
	"fmt"

	"ai-devs3/internal/domain"
)

// Run executes the login workflow
func (app *App) Run(loginURL string, creds domain.Credentials) (string, error) {
	// Step 1: Fetch the login page
	htmlContent, err := app.pageFetcher.FetchPage(loginURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch login page: %w", err)
	}

	// Step 2: Extract the question
	question, err := app.questionExtractor.Extract(htmlContent)
	if err != nil {
		return "", fmt.Errorf("failed to extract question: %w", err)
	}
	fmt.Printf("Extracted question: %s\n", question.Text)

	// Step 3: Get answer from LLM
	answer, err := app.llmClient.GetAnswer(question)
	if err != nil {
		return "", fmt.Errorf("failed to get answer from LLM: %w", err)
	}
	fmt.Printf("LLM provided answer: %s\n", answer.Text)

	// Step 4: Submit login form
	loginResponse, err := app.httpService.Login(creds, answer)
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	fmt.Printf("Received login response: %s\n", loginResponse)

	// Step 5: Analyze the login response with LLM
	// loginResponseAnalysis, err := app.llmClient.AnalyzeLoginResponse(loginResponse)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to analyze login response: %w", err)
	// }
	// fmt.Printf("LLM provided login response analysis: %s\n", loginResponseAnalysis)

	return "", nil
}
