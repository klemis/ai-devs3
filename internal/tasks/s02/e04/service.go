package e04

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/internal/storage/cache"
	"ai-devs3/pkg/errors"
)

// Service handles the S02E04 file categorization task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
	cache      *cache.TaskCache
	config     *config.Config
}

// NewService creates a new S02E04 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client, taskCache *cache.TaskCache) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
		cache:      taskCache,
		config:     cfg,
	}
}

// ScanFilesDirectory scans the files directory and returns information about processable files
func (s *Service) ScanFilesDirectory(ctx context.Context, filesDir string) (*FileDirectory, error) {
	files, err := os.ReadDir(filesDir)
	if err != nil {
		return nil, errors.NewProcessingError("filesystem", "scan_directory", "failed to read directory", err)
	}

	var processableFiles []FileInfo
	supportedTypes := map[string]string{
		".txt": "text",
		".png": "image",
		".mp3": "audio",
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Name()))
		if fileType, supported := supportedTypes[ext]; supported {
			info, err := file.Info()
			if err != nil {
				continue
			}

			processableFiles = append(processableFiles, FileInfo{
				Name:      file.Name(),
				Path:      filepath.Join(filesDir, file.Name()),
				Size:      info.Size(),
				Extension: ext,
				Type:      fileType,
			})
		}
	}

	return &FileDirectory{
		Path:      filesDir,
		FileCount: len(processableFiles),
		Files:     processableFiles,
	}, nil
}

// ProcessFiles processes all files in the directory concurrently
func (s *Service) ProcessFiles(ctx context.Context, fileDir *FileDirectory, options *ProcessingOptions) ([]ProcessingResult, error) {
	if len(fileDir.Files) == 0 {
		return []ProcessingResult{}, nil
	}

	// Create channels for worker pool
	fileChan := make(chan FileInfo, len(fileDir.Files))
	resultChan := make(chan ProcessingResult, len(fileDir.Files))

	// Send files to process
	go func() {
		defer close(fileChan)
		for _, file := range fileDir.Files {
			select {
			case <-ctx.Done():
				return
			case fileChan <- file:
			}
		}
	}()

	// Start worker pool
	var wg sync.WaitGroup
	numWorkers := options.MaxWorkers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU() * 2
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				select {
				case <-ctx.Done():
					return
				default:
					result := s.processFile(ctx, file, options)
					select {
					case resultChan <- result:
					case <-ctx.Done():
						return
					}
				}
			}
		}()
	}

	// Close result channel when all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []ProcessingResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// processFile processes a single file and extracts its content
func (s *Service) processFile(ctx context.Context, file FileInfo, options *ProcessingOptions) ProcessingResult {
	startTime := time.Now()
	result := ProcessingResult{
		FileData: FileData{
			Filename: file.Name,
			FilePath: file.Path,
			FileType: file.Type,
			Size:     file.Size,
		},
		ProcessTime: 0,
		FromCache:   false,
	}

	defer func() {
		result.ProcessTime = time.Since(startTime).Seconds()
	}()

	switch file.Type {
	case "text":
		content, err := s.processTextFile(file)
		if err != nil {
			result.Error = err
			return result
		}
		result.FileData.Content = content

	case "image":
		content, fromCache, err := s.processImageFile(ctx, file, options)
		if err != nil {
			result.Error = err
			return result
		}
		result.FileData.Content = content
		result.FromCache = fromCache

	case "audio":
		content, fromCache, err := s.processAudioFile(ctx, file, options)
		if err != nil {
			result.Error = err
			return result
		}
		result.FileData.Content = content
		result.FromCache = fromCache

	default:
		result.Error = fmt.Errorf("unsupported file type: %s", file.Type)
	}

	return result
}

// processTextFile processes a text file
func (s *Service) processTextFile(file FileInfo) (string, error) {
	content, err := os.ReadFile(file.Path)
	if err != nil {
		return "", fmt.Errorf("failed to read text file: %w", err)
	}
	return string(content), nil
}

// processImageFile processes an image file using OCR
func (s *Service) processImageFile(ctx context.Context, file FileInfo, options *ProcessingOptions) (string, bool, error) {
	// Check cache first
	if options.CacheEnabled {
		if cached, err := s.cache.GetOCRText(ctx, file.Name); err == nil {
			return cached, true, nil
		}
	}

	// Read image file
	imageData, err := os.ReadFile(file.Path)
	if err != nil {
		return "", false, fmt.Errorf("failed to read image file: %w", err)
	}

	// Perform OCR
	ocrText, err := s.llmClient.ExtractTextFromImage(ctx, imageData)
	if err != nil {
		return "", false, fmt.Errorf("failed to extract text from image: %w", err)
	}

	// Cache the result
	if options.CacheEnabled {
		if err := s.cache.SetOCRText(ctx, file.Name, ocrText); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to cache OCR result for %s: %v\n", file.Name, err)
		}
	}

	return ocrText, false, nil
}

// processAudioFile processes an audio file using transcription
func (s *Service) processAudioFile(ctx context.Context, file FileInfo, options *ProcessingOptions) (string, bool, error) {
	// Check cache first
	if options.CacheEnabled {
		if cached, err := s.cache.GetAudioTranscript(ctx, file.Name); err == nil {
			return cached, true, nil
		}
	}

	// Open audio file
	audioFile, err := os.Open(file.Path)
	if err != nil {
		return "", false, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer audioFile.Close()

	// Perform transcription
	transcript, err := s.llmClient.TranscribeAudio(ctx, audioFile, file.Name)
	if err != nil {
		return "", false, fmt.Errorf("failed to transcribe audio: %w", err)
	}

	// Cache the result
	if options.CacheEnabled {
		if err := s.cache.SetAudioTranscript(ctx, file.Name, transcript); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to cache transcription for %s: %v\n", file.Name, err)
		}
	}

	return transcript, false, nil
}

// CategorizeFiles categorizes the processed files into people and hardware
func (s *Service) CategorizeFiles(ctx context.Context, results []ProcessingResult) ([]CategoryResult, error) {
	var categories []CategoryResult

	for _, result := range results {
		if result.Error != nil {
			continue
		}

		if result.FileData.Content == "" || strings.TrimSpace(result.FileData.Content) == "" {
			categories = append(categories, CategoryResult{
				Filename:      result.FileData.Filename,
				Category:      "skip",
				Justification: "Empty or no content",
				Confidence:    1.0,
			})
			continue
		}

		// Use LLM to categorize
		categorization, err := s.llmClient.CategorizeContent(ctx, result.FileData.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to categorize file %s: %w", result.FileData.Filename, err)
		}

		categories = append(categories, CategoryResult{
			Filename:      result.FileData.Filename,
			Category:      categorization.Category,
			Justification: categorization.Justification,
			Confidence:    0.85, // Default confidence
		})
	}

	return categories, nil
}

// BuildCategorizedFiles builds the final categorized files structure
func (s *Service) BuildCategorizedFiles(categories []CategoryResult) *CategorizedFiles {
	categorized := &CategorizedFiles{
		People:   []string{},
		Hardware: []string{},
	}

	for _, category := range categories {
		switch category.Category {
		case "people":
			categorized.People = append(categorized.People, category.Filename)
		case "hardware":
			categorized.Hardware = append(categorized.Hardware, category.Filename)
		}
	}

	return categorized
}

// SubmitCategorization submits the categorization results to the centrala API
func (s *Service) SubmitCategorization(ctx context.Context, apiKey string, categorized *CategorizedFiles) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("kategorie", apiKey, categorized)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit categorization: %w", err)
	}

	return result, nil
}

// ExecuteTask executes the complete S02E04 task workflow
func (s *Service) ExecuteTask(ctx context.Context, filesDir string, apiKey string) (*TaskResult, error) {
	// Step 1: Scan files directory
	fileDir, err := s.ScanFilesDirectory(ctx, filesDir)
	if err != nil {
		return nil, errors.NewTaskError("s02e04", "scan_files_directory", err)
	}

	if fileDir.FileCount == 0 {
		return nil, errors.NewTaskError("s02e04", "scan_files_directory",
			fmt.Errorf("no processable files found in directory: %s", filesDir))
	}

	// Step 2: Process files
	options := &ProcessingOptions{
		MaxWorkers:     runtime.NumCPU() * 4,
		CacheEnabled:   true,
		CacheDir:       "s02e04",
		SupportedTypes: []string{"txt", "png", "mp3"},
		ProcessingDir:  filesDir,
	}

	results, err := s.ProcessFiles(ctx, fileDir, options)
	if err != nil {
		return nil, errors.NewTaskError("s02e04", "process_files", err)
	}

	// Step 3: Categorize files
	categories, err := s.CategorizeFiles(ctx, results)
	if err != nil {
		return nil, errors.NewTaskError("s02e04", "categorize_files", err)
	}

	// Step 4: Build categorized files structure
	categorized := s.BuildCategorizedFiles(categories)

	// Step 5: Submit categorization
	response, err := s.SubmitCategorization(ctx, apiKey, categorized)
	if err != nil {
		return nil, errors.NewTaskError("s02e04", "submit_categorization", err)
	}

	// Build processing stats
	stats := s.BuildProcessingStats(results, categories)

	return &TaskResult{
		Response:         response,
		Success:          strings.Contains(response, "correct") || strings.Contains(response, "success"),
		CategorizedFiles: categorized,
		ProcessingStats:  stats,
		TotalFiles:       fileDir.FileCount,
		CategorizedCount: len(categorized.People) + len(categorized.Hardware),
	}, nil
}

// BuildProcessingStats builds processing statistics
func (s *Service) BuildProcessingStats(results []ProcessingResult, categories []CategoryResult) *ProcessingStats {
	stats := &ProcessingStats{
		TotalFiles: len(results),
	}

	var totalProcessingTime float64
	cacheHits := 0

	for _, result := range results {
		totalProcessingTime += result.ProcessTime

		if result.FromCache {
			cacheHits++
		}

		if result.Error != nil {
			stats.ErrorFiles++
			continue
		}

		stats.ProcessedFiles++

		switch result.FileData.FileType {
		case "text":
			stats.TextFiles++
		case "image":
			stats.ImageFiles++
		case "audio":
			stats.AudioFiles++
		}
	}

	stats.ProcessingTime = totalProcessingTime
	if stats.TotalFiles > 0 {
		stats.CacheHitRate = float64(cacheHits) / float64(stats.TotalFiles)
	}

	for _, category := range categories {
		switch category.Category {
		case "people":
			stats.PeopleFiles++
		case "hardware":
			stats.HardwareFiles++
		default:
			stats.SkippedFiles++
		}
	}

	return stats
}

// PrintProcessingStats prints detailed processing statistics
func (s *Service) PrintProcessingStats(stats *ProcessingStats) {
	fmt.Println("=== S02E04 Processing Statistics ===")
	fmt.Printf("Total files: %d\n", stats.TotalFiles)
	fmt.Printf("Processed files: %d\n", stats.ProcessedFiles)
	fmt.Printf("Error files: %d\n", stats.ErrorFiles)
	fmt.Printf("Skipped files: %d\n", stats.SkippedFiles)
	fmt.Println()

	fmt.Printf("File types processed:\n")
	fmt.Printf("  Text files: %d\n", stats.TextFiles)
	fmt.Printf("  Image files: %d\n", stats.ImageFiles)
	fmt.Printf("  Audio files: %d\n", stats.AudioFiles)
	fmt.Println()

	fmt.Printf("Categorization results:\n")
	fmt.Printf("  People files: %d\n", stats.PeopleFiles)
	fmt.Printf("  Hardware files: %d\n", stats.HardwareFiles)
	fmt.Println()

	fmt.Printf("Performance:\n")
	fmt.Printf("  Total processing time: %.2f seconds\n", stats.ProcessingTime)
	fmt.Printf("  Cache hit rate: %.1f%%\n", stats.CacheHitRate*100)
	fmt.Println("=====================================")
}
