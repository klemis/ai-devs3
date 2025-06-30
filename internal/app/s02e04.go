package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// FileData represents extracted content from a file
type FileData struct {
	Filename string
	Content  string
	FileType string // "txt", "image", "audio"
}

// CategorizedFiles represents the categorization result
type CategorizedFiles struct {
	People   []string `json:"people"`
	Hardware []string `json:"hardware"`
}

// RunS02E04 processes files from pliki_z_fabryki and categorizes them
func (app *App) RunS02E04(apiKey string) (string, error) {
	log.Println("Starting S02E04 file categorization task")

	// Step 1: Read all files from pliki_z_fabryki directory
	filesDir := "../lessons-md/pliki_z_fabryki"
	cacheDir := "data/s02e04"

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Get list of files to process
	files, err := os.ReadDir(filesDir)
	if err != nil {
		return "", fmt.Errorf("failed to read directory %s: %w", filesDir, err)
	}

	// Filter files - only process top-level txt, png, and mp3 files
	var filesToProcess []os.DirEntry
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".txt" || ext == ".png" || ext == ".mp3" {
			filesToProcess = append(filesToProcess, file)
		}
	}

	log.Printf("Found %d files to process", len(filesToProcess))

	// Step 2: Process files concurrently with worker pool
	numWorkers := runtime.NumCPU() * 4
	fileDataChan := make(chan FileData, len(filesToProcess))
	errorChan := make(chan error, len(filesToProcess))

	// Send files to process
	fileChan := make(chan os.DirEntry, len(filesToProcess))
	go func() {
		defer close(fileChan)
		for _, file := range filesToProcess {
			fileChan <- file
		}
	}()

	// Process files with worker pool
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileChan {
				fileData, err := app.processFile(file, filesDir, cacheDir)
				if err != nil {
					errorChan <- fmt.Errorf("failed to process file %s: %w", file.Name(), err)
					continue
				}
				if fileData.Content != "" {
					fileDataChan <- fileData
				}
			}
		}()
	}

	// Wait for processing to complete and close channels
	go func() {
		wg.Wait()
		close(fileDataChan)
		close(errorChan)
	}()

	// Collect processed file data and errors
	var allFileData []FileData
	var errors []error

	// Use select to read from both channels until they're closed
	fileDataDone := false
	errorDone := false

	for !fileDataDone || !errorDone {
		select {
		case fileData, ok := <-fileDataChan:
			if !ok {
				fileDataDone = true
			} else {
				allFileData = append(allFileData, fileData)
			}
		case err, ok := <-errorChan:
			if !ok {
				errorDone = true
			} else {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		log.Printf("Encountered %d errors during processing:", len(errors))
		for _, err := range errors {
			log.Printf("  - %v", err)
		}
	}

	log.Printf("Successfully processed %d files", len(allFileData))

	// Step 3: Categorize files using LLM
	categorizedFiles := CategorizedFiles{
		People:   []string{},
		Hardware: []string{},
	}

	for _, fileData := range allFileData {
		category, err := app.categorizeFile(fileData)
		if err != nil {
			log.Printf("Failed to categorize file %s: %v", fileData.Filename, err)
			continue
		}

		switch category {
		case "people":
			categorizedFiles.People = append(categorizedFiles.People, fileData.Filename)
		case "hardware":
			categorizedFiles.Hardware = append(categorizedFiles.Hardware, fileData.Filename)
		default:
			log.Printf("File %s doesn't fit into any category, skipping", fileData.Filename)
		}
	}

	log.Printf("Categorization results: %d people files, %d hardware files",
		len(categorizedFiles.People), len(categorizedFiles.Hardware))

	// Step 4: Send report to endpoint
	response := app.httpClient.BuildResponse("kategorie", categorizedFiles)
	result, err := app.httpClient.PostReport(response)
	if err != nil {
		return "", fmt.Errorf("failed to submit report: %w", err)
	}

	return result, nil
}

// processFile extracts content from a single file
func (app *App) processFile(file os.DirEntry, filesDir, cacheDir string) (FileData, error) {
	filename := file.Name()
	filePath := filepath.Join(filesDir, filename)
	ext := strings.ToLower(filepath.Ext(filename))

	fileData := FileData{
		Filename: filename,
		FileType: strings.TrimPrefix(ext, "."),
	}

	switch ext {
	case ".txt":
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fileData, fmt.Errorf("failed to read text file: %w", err)
		}
		fileData.Content = string(content)

	case ".png":
		// Check for cached OCR result
		cacheFile := filepath.Join(cacheDir, filename+".ocr.txt")
		if cachedContent, err := os.ReadFile(cacheFile); err == nil {
			fileData.Content = string(cachedContent)
			log.Printf("Using cached OCR for %s", filename)
		} else {
			// Perform OCR
			imageData, err := os.ReadFile(filePath)
			if err != nil {
				return fileData, fmt.Errorf("failed to read image file: %w", err)
			}

			ocrText, err := app.llmClient.ExtractTextFromImage(imageData)
			if err != nil {
				return fileData, fmt.Errorf("failed to extract text from image: %w", err)
			}

			fileData.Content = ocrText

			// Cache the result
			if err := os.WriteFile(cacheFile, []byte(ocrText), 0644); err != nil {
				log.Printf("Failed to cache OCR result for %s: %v", filename, err)
			} else {
				log.Printf("Cached OCR result for %s", filename)
			}
		}

	case ".mp3":
		// Check for cached transcription
		cacheFile := filepath.Join(cacheDir, filename+".whisper.txt")
		if cachedContent, err := os.ReadFile(cacheFile); err == nil {
			fileData.Content = string(cachedContent)
			log.Printf("Using cached transcription for %s", filename)
		} else {
			// Perform transcription
			audioFile, err := os.Open(filePath)
			if err != nil {
				return fileData, fmt.Errorf("failed to open audio file: %w", err)
			}
			defer audioFile.Close()

			transcript, err := app.llmClient.TranscribeAudio(audioFile, filename)
			if err != nil {
				return fileData, fmt.Errorf("failed to transcribe audio: %w", err)
			}

			fileData.Content = transcript

			// Cache the result
			if err := os.WriteFile(cacheFile, []byte(transcript), 0644); err != nil {
				log.Printf("Failed to cache transcription for %s: %v", filename, err)
			} else {
				log.Printf("Cached transcription for %s", filename)
			}
		}

	default:
		return fileData, fmt.Errorf("unsupported file type: %s", ext)
	}

	return fileData, nil
}

// categorizeFile determines if a file contains people or hardware information
func (app *App) categorizeFile(fileData FileData) (string, error) {
	if fileData.Content == "" || strings.TrimSpace(fileData.Content) == "" {
		return "skip", nil
	}

	// Use the LLM client to categorize the content
	result, err := app.llmClient.CategorizeContent(fileData.Content)
	if err != nil {
		log.Printf("LLM categorization failed for %s, err: %v", fileData.Filename, err)
		return "", err
	}

	log.Printf("Categorized %s as '%s': %s", fileData.Filename, result.Category, result.Justification)
	return result.Category, nil
}
