package e01

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the image restoration and description task
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

// ExecuteTask executes the complete S04E01 task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	startTime := time.Now()
	stats := &ProcessingStats{
		StartTime:        startTime,
		OperationsByType: make(map[string]int),
		PhotoIterations:  make(map[string]int),
	}

	// Step 1: Get initial photos from the API
	photos, err := s.fetchInitialPhotos(ctx, apiKey)
	if err != nil {
		return nil, errors.NewTaskError("s04e01", "fetch_initial_photos", err)
	}

	stats.TotalPhotos = len(photos)
	log.Printf("Fetched %d initial photos", len(photos))

	// Step 2: Create restoration session
	session := &RestorationSession{
		Photos:     make(map[string]*PhotoInfo),
		StartTime:  startTime,
		Status:     SessionRunning,
		Operations: make([]OperationCommand, 0),
	}

	// Initialize photo info for each photo
	for filename, url := range photos {
		session.Photos[filename] = &PhotoInfo{
			CurrentFilename: filename,
			OriginalURL:     url,
			Iterations:      0,
			Operations:      make([]string, 0),
			Status:          StatusProcessing,
			LastUpdated:     startTime,
			Selected:        false,
		}
	}

	// Step 3: Process each photo iteratively
	for filename := range session.Photos {
		err := s.processPhoto(ctx, apiKey, session, filename, stats)
		if err != nil {
			log.Printf("Failed to process photo %s: %v", filename, err)
			session.Photos[filename].Status = StatusFailed
		}
		stats.ProcessedPhotos++

		// Small delay between photos to respect rate limits
		time.Sleep(500 * time.Millisecond)
	}

	// Step 4: Select photos showing Barbara
	selectedPhotos := s.selectPhotosOfBarbara(session)
	session.SelectedPhotos = selectedPhotos
	stats.SelectedPhotos = len(selectedPhotos)

	log.Printf("Selected %d photos showing Barbara", len(selectedPhotos))

	// Step 5: Generate final Polish rysopis
	rysopis, err := s.generateRysopis(ctx, session, selectedPhotos)
	if err != nil {
		return nil, errors.NewTaskError("s04e01", "generate_rysopis", err)
	}

	// Step 6: Submit the final description
	response, err := s.submitFinalResponse(ctx, apiKey, rysopis)
	if err != nil {
		return nil, errors.NewTaskError("s04e01", "submit_final_response", err)
	}

	// Calculate final stats
	stats.EndTime = time.Now()
	stats.ProcessingTime = stats.EndTime.Sub(stats.StartTime).Seconds()
	stats.TotalOperations = len(session.Operations)

	session.Status = SessionCompleted

	return &TaskResult{
		Response:        response,
		PhotosProcessed: stats.ProcessedPhotos,
		OperationsCount: stats.TotalOperations,
		SelectedPhotos:  selectedPhotos,
		FinalRysopis:    rysopis,
		ProcessingStats: stats,
	}, nil
}

// fetchInitialPhotos retrieves the initial photos from the central API
func (s *Service) fetchInitialPhotos(ctx context.Context, apiKey string) (map[string]string, error) {
	// Send initial request to get photos
	request := s.httpClient.BuildAIDevsResponse("photos", apiKey, "START")

	responseStr, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", request)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial photos: %w", err)
	}

	log.Printf("Initial photos response: %s", responseStr)

	// Parse response to extract photo URLs/filenames using LLM
	photos, err := s.parseResponseWithLLM(ctx, responseStr, "photos")
	if err != nil {
		return nil, fmt.Errorf("failed to parse photos response: %w", err)
	}

	return photos, nil
}

// parseResponseWithLLM uses LLM to parse various types of responses
func (s *Service) parseResponseWithLLM(ctx context.Context, response, responseType string) (map[string]string, error) {
	var systemPrompt string

	if responseType == "photos" {
		systemPrompt = `You are an expert at parsing Polish bot responses to extract photo information.

Parse this bot response and extract:
1. All photo filenames (usually in format IMG_XXX.PNG, IMG_XXXX.PNG, etc.)
2. The base URL where photos are stored (if mentioned)
3. Individual photo URLs if provided

Important rules:
- Look for filenames like IMG_559.PNG, IMG_1410.PNG, etc.
- Base URLs are often mentioned like "https://centrala.ag3nts.org/dane/barbara/"
- If only base URL + filenames are provided, construct full URLs by combining them
- If no base URL is found, use "https://centrala.ag3nts.org/dane/barbara/" as default
- Always return valid photo URLs, not placeholder URLs

Return JSON with filename -> full_photo_url mapping:
{
  "IMG_559.PNG": "https://centrala.ag3nts.org/dane/barbara/IMG_559.PNG",
  "IMG_1410.PNG": "https://centrala.ag3nts.org/dane/barbara/IMG_1410.PNG"
}`
	} else {
		systemPrompt = `You are an expert at parsing Polish bot responses from image restoration operations.

Parse this bot response and extract:
1. New filename if the operation created a modified file
2. Whether the operation was successful
3. Any suggestions for next operations
4. Brief note about what happened

Look for:
- New filenames (often with suffixes like _FXER, _BRIGHT, _DARK)
- Success indicators (OK, sukces, gotowe, etc.)
- Error indicators (błąd, error, fail, etc.)
- Operation suggestions (REPAIR, BRIGHTEN, DARKEN)

Return JSON:
{
  "_thinking": "detailed reasoning",
  "latest_filename": "new filename or original if unchanged",
  "success": true/false/null,
  "suggested_next": "REPAIR/BRIGHTEN/DARKEN or null",
  "note": "brief summary"
}`
	}

	userPrompt := fmt.Sprintf("Parse this %s response: %s", responseType, response)

	responseStr, err := s.llmClient.GetAnswerWithContext(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("LLM parsing failed: %w", err)
	}

	// Clean the response
	cleanResponse := strings.TrimSpace(responseStr)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	if responseType == "photos" {
		// Parse photos response
		var photos map[string]string
		if err := json.Unmarshal([]byte(cleanResponse), &photos); err != nil {
			log.Printf("Warning: failed to parse photos response JSON, using fallback: %v", err)
			return s.fallbackParsePhotos(response)
		}

		if len(photos) == 0 {
			return nil, fmt.Errorf("no photos found in response: %s", response)
		}

		return photos, nil
	} else {
		// Parse bot operation response - return as single entry map for compatibility
		var parsed BotResponseParser
		if err := json.Unmarshal([]byte(cleanResponse), &parsed); err != nil {
			log.Printf("Warning: failed to parse bot response JSON, using fallback: %v", err)
			// Fallback parsing
			filename := s.fallbackParseFilename(response, "")
			result := make(map[string]string)
			result["filename"] = filename
			result["success"] = "unknown"
			return result, nil
		}

		// Convert to map format for compatibility
		result := make(map[string]string)
		result["filename"] = parsed.LatestFilename
		if parsed.Success != nil {
			result["success"] = fmt.Sprintf("%t", *parsed.Success)
		} else {
			result["success"] = "unknown"
		}
		if parsed.SuggestedNext != nil {
			result["suggested_next"] = *parsed.SuggestedNext
		}
		result["note"] = parsed.Note

		return result, nil
	}
}

// fallbackParsePhotos provides regex-based fallback for photo parsing
func (s *Service) fallbackParsePhotos(response string) (map[string]string, error) {
	photos := make(map[string]string)

	// Look for filename patterns
	filenameRegex := regexp.MustCompile(`(IMG_\d+\.PNG|IMG_\d+\.JPG|IMG_\d+\.JPEG)`)
	filenames := filenameRegex.FindAllString(strings.ToUpper(response), -1)

	// Look for base URL
	baseURL := "https://centrala.ag3nts.org/dane/barbara/"
	urlRegex := regexp.MustCompile(`https://[^\s]+/`)
	if urlMatches := urlRegex.FindAllString(response, -1); len(urlMatches) > 0 {
		baseURL = urlMatches[0]
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
	}

	// Construct photo URLs
	for _, filename := range filenames {
		photos[filename] = baseURL + filename
	}

	if len(photos) == 0 {
		return nil, fmt.Errorf("no photos found in response: %s", response)
	}

	return photos, nil
}

// processPhoto handles the iterative restoration of a single photo
func (s *Service) processPhoto(ctx context.Context, apiKey string, session *RestorationSession, filename string, stats *ProcessingStats) error {
	photo := session.Photos[filename]

	for photo.Iterations < MaxIterationsPerPhoto {
		// Download and analyze the current image
		var imageData []byte
		var err error
		var analysis *VisionAnalysisResponse

		// Try to download the image
		imageData, err = s.httpClient.FetchBinaryData(ctx, photo.OriginalURL)
		if err != nil {
			log.Printf("Warning: failed to download image %s: %v", photo.CurrentFilename, err)
			// Skip this photo if we can't download it
			photo.Status = StatusFailed
			break
		}

		// Analyze with vision model
		analysis, err = s.analyzeImageWithVision(ctx, photo.CurrentFilename, imageData)
		if err != nil {
			log.Printf("Warning: failed to analyze image %s: %v", photo.CurrentFilename, err)
			// Try a simple heuristic based on filename
			analysis = &VisionAnalysisResponse{
				Filename:         photo.CurrentFilename,
				Decision:         OperationRepair,
				ExpectMorePasses: true,
				IsSubject:        true,
				QualityScore:     5,
			}
		}

		// Update photo selection status
		photo.Selected = analysis.IsSubject

		// Check if we should stop (NOOP or good quality)
		if analysis.Decision == OperationNoop || !analysis.ExpectMorePasses {
			photo.Status = StatusOptimal
			log.Printf("Photo %s is optimal after %d iterations", photo.CurrentFilename, photo.Iterations)
			break
		}

		// Send operation command to bot
		command := OperationCommand{
			Operation: analysis.Decision,
			Filename:  photo.CurrentFilename,
			Timestamp: time.Now(),
		}

		newFilename, success, err := s.sendOperationCommand(ctx, apiKey, command)
		if err != nil {
			return fmt.Errorf("failed to send operation %s for %s: %w", command.Operation, command.Filename, err)
		}

		// Update session tracking
		session.Operations = append(session.Operations, command)
		stats.OperationsByType[command.Operation]++
		photo.Operations = append(photo.Operations, command.Operation)
		photo.Iterations++
		photo.LastUpdated = time.Now()
		stats.PhotoIterations[filename] = photo.Iterations

		if !success {
			photo.Status = StatusFailed
			log.Printf("Operation %s failed for photo %s", command.Operation, photo.CurrentFilename)
			break
		}

		// Update filename if changed
		if newFilename != "" && newFilename != photo.CurrentFilename {
			log.Printf("Photo filename updated: %s -> %s", photo.CurrentFilename, newFilename)
			photo.CurrentFilename = newFilename
			// Update URL to point to new filename - use the base URL from original response
			baseURL := "https://centrala.ag3nts.org/dane/barbara/"
			if strings.Contains(photo.OriginalURL, "/dane/barbara/") {
				baseURL = "https://centrala.ag3nts.org/dane/barbara/"
			} else if strings.Contains(photo.OriginalURL, "/files/") {
				baseURL = "https://c3ntrala.ag3nts.org/files/"
			}
			photo.OriginalURL = baseURL + newFilename
		}

		// Small delay between operations
		time.Sleep(1 * time.Second)
	}

	if photo.Iterations >= MaxIterationsPerPhoto {
		photo.Status = StatusAbandoned
		log.Printf("Abandoned photo %s after %d iterations", photo.CurrentFilename, photo.Iterations)
	}

	return nil
}

// analyzeImageWithVision uses the vision model to analyze an image
func (s *Service) analyzeImageWithVision(ctx context.Context, filename string, imageData []byte) (*VisionAnalysisResponse, error) {
	responseStr, err := s.llmClient.AnalyzeImageForRestoration(ctx, filename, imageData)
	if err != nil {
		return nil, fmt.Errorf("vision analysis failed: %w", err)
	}

	// Clean the response to remove any markdown formatting
	cleanResponse := strings.TrimSpace(responseStr)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	var analysis VisionAnalysisResponse
	if err := json.Unmarshal([]byte(cleanResponse), &analysis); err != nil {
		log.Printf("Warning: failed to parse vision analysis JSON, using fallback: %v", err)
		log.Printf("Original response: %s", responseStr)
		log.Printf("Cleaned response: %s", cleanResponse)

		// Fallback: create a default analysis
		analysis = VisionAnalysisResponse{
			Thinking:         "Failed to parse LLM response, using fallback heuristic",
			Filename:         filename,
			Decision:         OperationRepair, // Default to repair as safest option
			ExpectMorePasses: true,
			IsSubject:        true,
			QualityScore:     5,
			IssuesDetected:   []string{"parsing_error"},
		}

		// Try to extract decision from raw response text
		responseUpper := strings.ToUpper(cleanResponse)
		if strings.Contains(responseUpper, "NOOP") {
			analysis.Decision = OperationNoop
			analysis.ExpectMorePasses = false
		} else if strings.Contains(responseUpper, "BRIGHTEN") {
			analysis.Decision = OperationBrighten
		} else if strings.Contains(responseUpper, "DARKEN") {
			analysis.Decision = OperationDarken
		} else if strings.Contains(responseUpper, "REPAIR") {
			analysis.Decision = OperationRepair
		}
	}

	return &analysis, nil
}

// sendOperationCommand sends a restoration command to the bot
func (s *Service) sendOperationCommand(ctx context.Context, apiKey string, command OperationCommand) (string, bool, error) {
	// Send only the filename, not URLs as per requirements
	commandStr := fmt.Sprintf("%s %s", command.Operation, command.Filename)

	request := s.httpClient.BuildAIDevsResponse("photos", apiKey, commandStr)

	responseStr, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", request)
	if err != nil {
		return "", false, fmt.Errorf("failed to send command: %w", err)
	}

	log.Printf("Command: %s, Response: %s", commandStr, responseStr)

	// Parse bot response using LLM
	result, err := s.parseResponseWithLLM(ctx, responseStr, "operation")
	if err != nil {
		log.Printf("Warning: failed to parse bot response: %v", err)
		// Try basic regex fallback
		newFilename := s.fallbackParseFilename(responseStr, command.Filename)
		// Assume success if we got a different filename or response looks positive
		success := newFilename != command.Filename || strings.Contains(strings.ToLower(responseStr), "ok") ||
			strings.Contains(strings.ToLower(responseStr), "sukces") ||
			strings.Contains(strings.ToLower(responseStr), "pomyśl")
		return newFilename, success, nil
	}

	// Extract values from result map
	filename := result["filename"]
	if filename == "" {
		filename = command.Filename
	}

	successStr := result["success"]
	success := successStr == "true"
	if successStr == "unknown" && filename != command.Filename {
		success = true
	}

	return filename, success, nil
}

// fallbackParseFilename uses regex to extract filenames as fallback
func (s *Service) fallbackParseFilename(response, currentFilename string) string {
	// Look for new filename patterns - try multiple patterns
	patterns := []string{
		`(IMG_\w+(?:_[A-Z]+)*\.(PNG|JPG|JPEG))`,
		`(IMG_\d+(?:_[A-Z]+)*\.(png|jpg|jpeg))`,
		`(\w+_\d+_[A-Z]+\.(PNG|JPG|JPEG))`,
		`([A-Z0-9_]+\.(PNG|JPG|JPEG))`,
	}

	for _, pattern := range patterns {
		filenameRegex := regexp.MustCompile(pattern)
		matches := filenameRegex.FindAllString(strings.ToUpper(response), -1)

		if len(matches) > 0 {
			// Filter out the current filename if present
			for _, match := range matches {
				if !strings.EqualFold(match, currentFilename) {
					return match
				}
			}
			// If no different filename found, return the last match
			return matches[len(matches)-1]
		}
	}

	return currentFilename
}

// selectPhotosOfBarbara identifies photos showing the same woman
func (s *Service) selectPhotosOfBarbara(session *RestorationSession) []string {
	var selected []string

	for filename, photo := range session.Photos {
		if photo.Selected && (photo.Status == StatusOptimal || photo.Status == StatusProcessing) {
			selected = append(selected, filename)
		}
	}

	// If too few selected, be more lenient
	if len(selected) < 2 {
		for filename, photo := range session.Photos {
			if photo.Status == StatusOptimal || photo.Status == StatusProcessing {
				selected = append(selected, filename)
				if len(selected) >= 3 {
					break
				}
			}
		}
	}

	return selected
}

// generateRysopis creates the final Polish description
func (s *Service) generateRysopis(ctx context.Context, session *RestorationSession, selectedPhotos []string) (string, error) {
	if len(selectedPhotos) == 0 {
		return "", fmt.Errorf("no photos selected for description")
	}

	// Prepare context with selected photos
	var photoURLs []string
	for _, filename := range selectedPhotos {
		if photo, exists := session.Photos[filename]; exists {
			photoURLs = append(photoURLs, photo.OriginalURL)
		}
	}

	rysopis, err := s.llmClient.GeneratePolishRysopis(ctx, photoURLs)
	if err != nil {
		return "", fmt.Errorf("failed to generate rysopis: %w", err)
	}

	// Clean up the response
	rysopis = strings.TrimSpace(rysopis)
	rysopis = strings.Trim(rysopis, "\"'")

	return rysopis, nil
}

// submitFinalResponse submits the final Polish description
func (s *Service) submitFinalResponse(ctx context.Context, apiKey string, rysopis string) (string, error) {
	request := s.httpClient.BuildAIDevsResponse("photos", apiKey, rysopis)

	response, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", request)
	if err != nil {
		return "", fmt.Errorf("failed to submit final response: %w", err)
	}

	return response, nil
}

// PrintProcessingStats prints detailed processing statistics
func (s *Service) PrintProcessingStats(stats *ProcessingStats) {
	fmt.Println("=== S04E01 Processing Statistics ===")
	fmt.Printf("Total photos: %d\n", stats.TotalPhotos)
	fmt.Printf("Processed photos: %d\n", stats.ProcessedPhotos)
	fmt.Printf("Selected photos: %d\n", stats.SelectedPhotos)
	fmt.Printf("Total operations: %d\n", stats.TotalOperations)

	fmt.Println("Operations by type:")
	for op, count := range stats.OperationsByType {
		fmt.Printf("  %s: %d\n", op, count)
	}

	fmt.Println("Iterations per photo:")
	for photo, iterations := range stats.PhotoIterations {
		fmt.Printf("  %s: %d iterations\n", photo, iterations)
	}

	fmt.Printf("Processing time: %.2f seconds\n", stats.ProcessingTime)
	fmt.Printf("Vision analysis cost: %d tokens\n", stats.VisionAnalysisCost)
	fmt.Println("=====================================")
}
