package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"ai-devs3/internal/domain"
)

type HTTPClient interface {
	Login(creds domain.Credentials, answer domain.Answer) (string, error)
	Verify(requestBody domain.AnswerRoboISO) (string, error)
}

// HTTPService implements HTTPClient
type HTTPService struct {
	URL string
}

func (s *HTTPService) Login(creds domain.Credentials, answer domain.Answer) (string, error) {
	// Create form data
	data := url.Values{}
	data.Set("username", creds.Username)
	data.Set("password", creds.Password)
	data.Set("answer", answer.Text)

	// Create request with form-urlencoded body
	req, err := http.NewRequest("POST", s.URL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send login request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read login response: %w", err)
	}

	return string(body), nil
}

func (s *HTTPService) Verify(requestBody domain.AnswerRoboISO) (string, error) {
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", s.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send verify request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read verify response: %w", err)
	}

	return string(body), nil
}
