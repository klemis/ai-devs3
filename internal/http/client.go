package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/pkg/errors"
)

// Client wraps http.Client with configuration and error handling
type Client struct {
	client *http.Client
	config config.HTTPConfig
}

// NewClient creates a new HTTP client with the given configuration
func NewClient(cfg config.HTTPConfig) *Client {
	return &Client{
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		config: cfg,
	}
}

// FetchPage retrieves the content of a web page
func (c *Client) FetchPage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("HTTP", 0, "failed to fetch page", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", errors.NewAPIError("HTTP", resp.StatusCode, "HTTP error", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// FetchJSONData downloads and unmarshals JSON from the given URL
func (c *Client) FetchJSONData(ctx context.Context, url string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.NewAPIError("HTTP", 0, "failed to fetch JSON", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.NewAPIError("HTTP", resp.StatusCode, "HTTP error", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return data, nil
}

// FetchData retrieves raw data from a URL
func (c *Client) FetchData(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("HTTP", 0, "failed to fetch data", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", errors.NewAPIError("HTTP", resp.StatusCode, "HTTP error", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// FetchBinaryData retrieves binary data from a URL
func (c *Client) FetchBinaryData(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, errors.NewAPIError("HTTP", 0, "failed to fetch binary data", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, errors.NewAPIError("HTTP", resp.StatusCode, "HTTP error", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

// PostJSON sends a JSON payload to the specified URL
func (c *Client) PostJSON(ctx context.Context, url string, payload any) (string, error) {
	var jsonData []byte
	var err error

	jsonData, err = json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("HTTP", 0, "failed to post JSON", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", errors.NewAPIError("HTTP", resp.StatusCode, "HTTP error: "+string(body), nil)
	}

	return string(body), nil
}

// PostForm sends form data to the specified URL
func (c *Client) PostForm(ctx context.Context, reqURL string, data map[string]string) (string, error) {
	formData := make(url.Values)
	for key, value := range data {
		formData.Set(key, value)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", errors.NewAPIError("HTTP", 0, "failed to post form", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", errors.NewAPIError("HTTP", resp.StatusCode, "HTTP error", nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

// BuildAIDevsResponse creates a response for AI-DEVS API
func (c *Client) BuildAIDevsResponse(task string, apiKey string, answer any) map[string]any {
	return map[string]any{
		"task":   task,
		"apikey": apiKey,
		"answer": answer,
	}
}

// PostReport sends a report to the AI-DEVS central endpoint
func (c *Client) PostReport(ctx context.Context, baseURL string, response map[string]any) (string, error) {
	reportURL := fmt.Sprintf("%s/report", baseURL)
	return c.PostJSON(ctx, reportURL, response)
}
