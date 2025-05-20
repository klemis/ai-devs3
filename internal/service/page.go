package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PageFetcher defines the interface for fetching web pages
type PageFetcher interface {
	FetchPage(url string) (string, error)
	FetchJSONData(url string) (map[string]interface{}, error)
}

// HTTPPageFetcher implements PageFetcher using HTTP
type HTTPPageFetcher struct{}

// FetchPage retrieves the content of a web page
func (f *HTTPPageFetcher) FetchPage(url string) (string, error) {
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
func (f *HTTPPageFetcher) FetchJSONData(url string) (map[string]interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JSON: %w", err)
	}
	defer resp.Body.Close()
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
