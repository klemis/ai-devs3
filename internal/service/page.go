package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

// HTTPClient defines the interface for HTTP operations
type HTTPClient interface {
	// Fetching methods
	FetchPage(url string) (string, error)
	FetchJSONData(url string) (map[string]interface{}, error)
	FetchData(url string) (string, error)
	FetchBinaryData(url string) ([]byte, error)

	// Posting methods
	BuildResponse(task string, answer interface{}) map[string]interface{}
	PostReport(response map[string]interface{}) (string, error)
}

// HTTPClientImpl implements HTTPClient using standard HTTP
type HTTPClientImpl struct {
	APIKey string
}

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

	respDump, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("=== RESPONSE ===\n%s\n", respDump)

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

// FetchBinaryData retrieves binary data from a URL
func (h *HTTPClientImpl) FetchBinaryData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	return body, nil
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

	respDump, _ := httputil.DumpResponse(resp, true)
	fmt.Printf("=== RESPONSE ===\n%s\n", respDump)

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

// buildResponse creates the final response object
func (h *HTTPClientImpl) BuildResponse(task string, answer interface{}) map[string]interface{} {
	return map[string]interface{}{
		"task":   task,
		"apikey": h.APIKey,
		"answer": answer,
	}
}
