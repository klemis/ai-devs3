package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// RunS02E01 handles the audio transcription and analysis task
func (app *App) RunS02E01() (string, error) {
	// Step 1: List audio files in the directory
	path, err := filepath.Abs("../lessons-md/przesluchania")
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	log.Println("Fetching list of audio files...")
	audioFiles, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("failed to list audio files: %w", err)
	}

	log.Printf("Found %d audio files: %v", len(audioFiles), audioFiles)

	// Step 2: Transcribe each audio file
	var allTranscripts string
	for i, audiofile := range audioFiles {
		log.Printf("Processing file %d/%d: %s", i+1, len(audioFiles), audiofile.Name())

		f, err := os.Open(path + "/" + audiofile.Name())
		if err != nil {
			log.Printf("Failed to open %s: %v", audiofile.Name(), err)
			continue
		}
		defer f.Close()

		// Transcribe audio using Whisper
		transcript, err := app.llmClient.TranscribeAudio(f, audiofile.Name())
		if err != nil {
			log.Printf("Failed to transcribe %s: %v", audiofile.Name(), err)
			continue
		}

		log.Printf("Transcribed %s: %s", audiofile.Name(), transcript[:min(100, len(transcript))])

		// Add to combined transcripts
		allTranscripts += fmt.Sprintf("\n=== TRANSCRIPT FROM %s ===\n%s\n", audiofile.Name(), transcript)
	}

	if allTranscripts == "" {
		return "", fmt.Errorf("no transcripts were generated")
	}

	log.Printf("Combined all transcripts (%d characters)", len(allTranscripts))

	// Step 3: Analyze transcripts to find the street where Professor Maj's institute is located
	log.Println("Analyzing transcripts to find Professor Maj's institute location...")
	analysis, err := app.llmClient.AnalyzeTranscripts(allTranscripts)
	if err != nil {
		return "", fmt.Errorf("failed to analyze transcripts: %w", err)
	}

	log.Printf("Analysis result: %s", analysis)

	// Step 4: Send answer to Centrala
	response := app.httpClient.BuildResponse("mp3", analysis.Answer)
	result, err := app.httpClient.PostReport(response)
	if err != nil {
		return "", fmt.Errorf("failed to post report: %w", err)
	}

	log.Printf("Posted answer '%s' to Centrala, response: %s", analysis.Answer, result)

	return result, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
