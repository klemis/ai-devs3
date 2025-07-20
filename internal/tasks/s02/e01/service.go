package e01

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the S02E01 audio transcription and analysis task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
	config     *config.Config
}

// NewService creates a new S02E01 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
		config:     cfg,
	}
}

// ListAudioFiles lists all audio files in the specified directory
func (s *Service) ListAudioFiles(ctx context.Context, audioDir string) (*AudioDirectory, error) {
	// Get absolute path
	absPath, err := filepath.Abs(audioDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Read directory
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, errors.NewProcessingError("filesystem", "list_audio_files", "failed to read directory", err)
	}

	var audioFiles []AudioFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if file is an audio file
		name := entry.Name()
		if !s.isAudioFile(name) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		audioFiles = append(audioFiles, AudioFile{
			Name: name,
			Path: filepath.Join(absPath, name),
			Size: info.Size(),
		})
	}

	return &AudioDirectory{
		Path:      absPath,
		FileCount: len(audioFiles),
		Files:     audioFiles,
	}, nil
}

// isAudioFile checks if a file is an audio file based on its extension
func (s *Service) isAudioFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	audioExtensions := []string{".mp3", ".wav", ".m4a", ".flac", ".ogg", ".aac"}

	return slices.Contains(audioExtensions, ext)
}

// TranscribeAudioFile transcribes a single audio file
func (s *Service) TranscribeAudioFile(ctx context.Context, audioFile AudioFile) (*Transcript, error) {
	// Open the audio file
	file, err := os.Open(audioFile.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file %s: %w", audioFile.Name, err)
	}
	defer file.Close()

	// Transcribe using OpenAI Whisper
	transcriptText, err := s.llmClient.TranscribeAudio(ctx, file, audioFile.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to transcribe audio file %s: %w", audioFile.Name, err)
	}

	return &Transcript{
		AudioFile: audioFile.Name,
		Text:      transcriptText,
		Length:    len(transcriptText),
		Success:   true,
	}, nil
}

// TranscribeAllAudioFiles transcribes all audio files in the directory
func (s *Service) TranscribeAllAudioFiles(ctx context.Context, audioDir *AudioDirectory) ([]Transcript, error) {
	var transcripts []Transcript

	for i, audioFile := range audioDir.Files {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Transcribe the audio file
		transcript, err := s.TranscribeAudioFile(ctx, audioFile)
		if err != nil {
			// Log error but continue with other files
			fmt.Printf("Failed to transcribe %s: %v\n", audioFile.Name, err)
			continue
		}

		transcripts = append(transcripts, *transcript)

		fmt.Printf("Processed file %d/%d: %s (%d characters)\n",
			i+1, len(audioDir.Files), audioFile.Name, transcript.Length)
	}

	if len(transcripts) == 0 {
		return nil, errors.NewProcessingError("transcription", "transcribe_all", "no transcripts were generated", nil)
	}

	return transcripts, nil
}

// CombineTranscripts combines all transcripts into a single text
func (s *Service) CombineTranscripts(transcripts []Transcript) string {
	var combined strings.Builder

	for _, transcript := range transcripts {
		combined.WriteString(fmt.Sprintf("\n=== TRANSCRIPT FROM %s ===\n%s\n",
			transcript.AudioFile, transcript.Text))
	}

	return combined.String()
}

// AnalyzeTranscripts analyzes transcripts to find Professor Maj's institute location
func (s *Service) AnalyzeTranscripts(ctx context.Context, combinedTranscripts string) (*TranscriptAnalysis, error) {
	if strings.TrimSpace(combinedTranscripts) == "" {
		return nil, errors.NewProcessingError("llm", "analyze_transcripts", "combined transcripts are empty", nil)
	}

	// Use OpenAI to analyze the transcripts
	analysis, err := s.llmClient.AnalyzeTranscripts(ctx, combinedTranscripts)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze transcripts: %w", err)
	}

	return &TranscriptAnalysis{
		Thinking: analysis.Thinking,
		Answer:   analysis.Answer,
	}, nil
}

// SubmitAnswer submits the analysis result to the centrala API
func (s *Service) SubmitAnswer(ctx context.Context, apiKey string, analysis *TranscriptAnalysis) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("mp3", apiKey, analysis.Answer)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit answer: %w", err)
	}

	return result, nil
}

// ExecuteTask executes the complete S02E01 task workflow
func (s *Service) ExecuteTask(ctx context.Context, audioDir string, apiKey string) (*TaskResult, error) {
	// Step 1: List audio files
	audioDirectory, err := s.ListAudioFiles(ctx, audioDir)
	if err != nil {
		return nil, errors.NewTaskError("s02e01", "list_audio_files", err)
	}

	if audioDirectory.FileCount == 0 {
		return nil, errors.NewTaskError("s02e01", "list_audio_files",
			fmt.Errorf("no audio files found in directory: %s", audioDir))
	}

	// Step 2: Transcribe all audio files
	transcripts, err := s.TranscribeAllAudioFiles(ctx, audioDirectory)
	if err != nil {
		return nil, errors.NewTaskError("s02e01", "transcribe_audio_files", err)
	}

	// Step 3: Combine transcripts
	combinedTranscripts := s.CombineTranscripts(transcripts)

	// Step 4: Analyze transcripts
	analysis, err := s.AnalyzeTranscripts(ctx, combinedTranscripts)
	if err != nil {
		return nil, errors.NewTaskError("s02e01", "analyze_transcripts", err)
	}

	// Step 5: Submit answer
	response, err := s.SubmitAnswer(ctx, apiKey, analysis)
	if err != nil {
		return nil, errors.NewTaskError("s02e01", "submit_answer", err)
	}

	return &TaskResult{
		Response:         response,
		Success:          strings.Contains(response, "correct") || strings.Contains(response, "success"),
		TranscriptCount:  len(transcripts),
		TotalTranscripts: combinedTranscripts,
		Analysis:         analysis,
	}, nil
}
