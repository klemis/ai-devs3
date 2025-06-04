package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	// Fetching methods
	FetchPage(url string) (string, error)
	FetchJSONData(url string) (map[string]interface{}, error)
	FetchData(url string) (string, error)

	// Posting methods
	PostReport(response map[string]interface{}) (string, error)
}

// HTTPClientImpl implements HTTPClient using standard HTTP
type HTTPClientImpl struct{}

// FetchPage retrieves the content of a web page
func (h *HTTPClientImpl) FetchPage(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read page content: %w", err)
	}

	return string(body), nil
}

// FetchJSONData downloads and unmarshals the JSON from the given URL
func (h *HTTPClientImpl) FetchJSONData(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JSON: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return data, nil
}

// FetchData retrieves raw data from a URL (non-JSON format)
func (h *HTTPClientImpl) FetchData(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read data content: %w", err)
	}

	return string(body), nil
}

// PostReport sends the response to the report endpoint
func (h *HTTPClientImpl) PostReport(response map[string]interface{}) (string, error) {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(response)
	if err != nil {
		return "", fmt.Errorf("failed to encode response: %w", err)
	}

	resp, err := http.Post("https://c3ntrala.ag3nts.org/report", "application/json", buf)
	if err != nil {
		return "", fmt.Errorf("failed to post report: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}
