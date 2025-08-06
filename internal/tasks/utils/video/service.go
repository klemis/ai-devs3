package video

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
)

const (
	MaxFileSizeBytes            = 25 * 1024 * 1024 // 25MB
	DefaultChunkDurationSeconds = 600              // 10 minutes per chunk
)

// Service handles video transcription processing
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

// ProcessVideoFromURL downloads video, converts to audio, and transcribes it
func (s *Service) ProcessVideoFromURL(ctx context.Context, videoURL string) (*TranscriptionResult, error) {
	log.Printf("Starting video transcription for URL: %s", videoURL)

	// Create temporary directory for processing
	tempDir, err := os.MkdirTemp("", "video_transcription_*")
	if err != nil {
		return &TranscriptionResult{
			VideoURL: videoURL,
			Error:    fmt.Sprintf("Failed to create temp directory: %v", err),
		}, err
	}
	defer os.RemoveAll(tempDir)

	// Download audio directly
	audioData, err := s.downloadAudio(ctx, videoURL, tempDir)
	if err != nil {
		return &TranscriptionResult{
			VideoURL: videoURL,
			Error:    fmt.Sprintf("Failed to download audio: %v", err),
		}, err
	}

	log.Printf("Downloaded audio: %s (size: %d bytes)", audioData.Filename, audioData.Size)

	// Check if audio file is too large and needs splitting
	var transcription string
	if audioData.Size > MaxFileSizeBytes {
		log.Printf("Audio file too large (%d bytes), splitting into chunks", audioData.Size)
		transcription, err = s.transcribeInChunks(ctx, audioData, tempDir)
	} else {
		transcription, err = s.transcribeSingleFile(ctx, audioData)
	}

	if err != nil {
		return &TranscriptionResult{
			VideoURL:  videoURL,
			AudioFile: audioData.Filename,
			Error:     fmt.Sprintf("Failed to transcribe audio: %v", err),
		}, err
	}

	return &TranscriptionResult{
		VideoURL:      videoURL,
		AudioFile:     audioData.Filename,
		Transcription: transcription,
		Duration:      audioData.Duration,
		FileSize:      audioData.Size,
	}, nil
}

// downloadAudio downloads audio directly from URL using yt-dlp and saves to temp directory
func (s *Service) downloadAudio(ctx context.Context, videoURL, tempDir string) (*AudioData, error) {
	// Try yt-dlp first if available
	if err := s.checkYtDlpAvailable(); err == nil {
		return s.downloadAudioWithYtDlp(ctx, videoURL, tempDir)
	}

	// Fallback not available for audio-only download
	return nil, fmt.Errorf("yt-dlp required for audio extraction. Please install: pip install yt-dlp")
}

// downloadAudioWithYtDlp downloads audio using yt-dlp and trims to last 3 seconds
func (s *Service) downloadAudioWithYtDlp(ctx context.Context, videoURL, tempDir string) (*AudioData, error) {
	// First download full audio to get duration
	fullAudioFile := filepath.Join(tempDir, "full_audio.mp3")

	log.Printf("Downloading full audio from URL: %s", videoURL)

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-x", // Extract audio only
		"--audio-format", "mp3",
		"--audio-quality", "0", // Best quality
		"-o", fullAudioFile,
		"--no-playlist",
		"--no-warnings",
		videoURL,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp audio download failed: %w, output: %s", err, string(output))
	}

	// Get audio duration
	duration, err := s.getAudioDurationFloat(fullAudioFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	// Calculate start time for last 4 seconds
	var startTime float64
	var trimDuration float64 = 4

	if duration <= 4 {
		log.Printf("Audio is %.2f seconds, using full audio", duration)
		startTime = 0
		trimDuration = duration
	} else {
		startTime = duration - 4
		log.Printf("Trimming audio to last 4 seconds (from %.2fs to %.2fs)", startTime, duration)
	}

	// Create trimmed and reversed audio file with best quality
	outputFile := filepath.Join(tempDir, "audio.mp3")

	// Also create a permanent copy in data directory
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	urlPart := "audio"
	if strings.Contains(videoURL, "vimeo.com") {
		parts := strings.Split(videoURL, "/")
		if len(parts) > 0 {
			urlPart = "vimeo_" + parts[len(parts)-1]
		}
	}
	permanentFile := filepath.Join("data", fmt.Sprintf("%s_%s_reversed.mp3", timestamp, urlPart))
	cmd = exec.Command("ffmpeg",
		"-i", fullAudioFile,
		"-ss", fmt.Sprintf("%.2f", startTime),
		"-t", fmt.Sprintf("%.2f", trimDuration),
		"-af", "areverse", // Reverse the audio
		"-acodec", "libmp3lame",
		"-b:a", "320k", // Best quality bitrate
		"-ar", "44100", // High sample rate
		"-ac", "2", // Stereo for best quality
		"-y", // Overwrite output file
		outputFile,
	)

	output, err = cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg trim and reverse failed: %w, output: %s", err, string(output))
	}

	// Remove full audio file to save space
	os.Remove(fullAudioFile)

	// Copy the processed audio to data directory for reference
	if err := s.copyFile(outputFile, permanentFile); err != nil {
		log.Printf("Warning: failed to save audio file to data directory: %v", err)
	} else {
		log.Printf("Saved processed audio file to: %s", permanentFile)
	}

	// Check if trimmed file was created
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to get trimmed audio file info: %w", err)
	}

	// Get final duration
	finalDuration, err := s.getAudioDuration(outputFile)
	if err != nil {
		log.Printf("Warning: could not get final audio duration: %v", err)
		finalDuration = fmt.Sprintf("~%.1fs", trimDuration)
	}

	return &AudioData{
		Filename: outputFile,
		Size:     fileInfo.Size(),
		Duration: finalDuration,
		Format:   "mp3",
	}, nil
}

// checkYtDlpAvailable checks if yt-dlp is available on the system
func (s *Service) checkYtDlpAvailable() error {
	cmd := exec.Command("yt-dlp", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp not found: %w", err)
	}
	return nil
}

// getAudioDurationFloat gets the duration of audio file as float64 using ffprobe
func (s *Service) getAudioDurationFloat(filename string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filename,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	if durationStr == "" {
		return 0, fmt.Errorf("empty duration output")
	}

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", err)
	}

	return duration, nil
}

// getAudioDuration gets the duration of audio file using ffprobe
func (s *Service) getAudioDuration(filename string) (string, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		filename,
	)

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	durationStr := strings.TrimSpace(string(output))
	if durationStr == "" {
		return "", fmt.Errorf("empty duration output")
	}

	// Convert to human readable format
	durationFloat, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return durationStr, nil // Return raw duration if parsing fails
	}

	minutes := int(durationFloat) / 60
	seconds := int(durationFloat) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds), nil
}

// copyFile copies a file from src to dst
func (s *Service) copyFile(src, dst string) error {
	// Ensure destination directory exists
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// transcribeSingleFile transcribes a single audio file
func (s *Service) transcribeSingleFile(ctx context.Context, audioData *AudioData) (string, error) {
	file, err := os.Open(audioData.Filename)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	transcription, err := s.llmClient.TranscribeAudio(ctx, file, filepath.Base(audioData.Filename))
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	return transcription, nil
}

// transcribeInChunks splits large audio file into chunks and transcribes each
func (s *Service) transcribeInChunks(ctx context.Context, audioData *AudioData, tempDir string) (string, error) {
	// Split audio into chunks using ffmpeg
	chunks, err := s.splitAudioIntoChunks(audioData, tempDir)
	if err != nil {
		return "", fmt.Errorf("failed to split audio: %w", err)
	}

	var transcriptions []string

	for i, chunk := range chunks {
		log.Printf("Transcribing chunk %d/%d: %s", i+1, len(chunks), chunk.Filename)

		file, err := os.Open(chunk.Filename)
		if err != nil {
			log.Printf("Warning: failed to open chunk %s: %v", chunk.Filename, err)
			continue
		}

		transcription, err := s.llmClient.TranscribeAudio(ctx, file, filepath.Base(chunk.Filename))
		file.Close()

		if err != nil {
			log.Printf("Warning: failed to transcribe chunk %s: %v", chunk.Filename, err)
			continue
		}

		if transcription != "" {
			transcriptions = append(transcriptions, transcription)
		}
	}

	return strings.Join(transcriptions, " "), nil
}

// splitAudioIntoChunks splits audio file into smaller chunks
func (s *Service) splitAudioIntoChunks(audioData *AudioData, tempDir string) ([]ChunkInfo, error) {
	// Get total duration first
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "format=duration",
		"-of", "csv=p=0",
		audioData.Filename,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get audio duration: %w", err)
	}

	totalDurationFloat, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %w", err)
	}

	totalDuration := int(totalDurationFloat)
	chunkDuration := DefaultChunkDurationSeconds
	numChunks := (totalDuration + chunkDuration - 1) / chunkDuration

	var chunks []ChunkInfo

	for i := 0; i < numChunks; i++ {
		startTime := i * chunkDuration
		chunkFilename := filepath.Join(tempDir, fmt.Sprintf("chunk_%03d.mp3", i))

		// Use ffmpeg to extract chunk
		cmd := exec.Command("ffmpeg",
			"-i", audioData.Filename,
			"-ss", fmt.Sprintf("%d", startTime),
			"-t", fmt.Sprintf("%d", chunkDuration),
			"-acodec", "copy",
			"-y",
			chunkFilename,
		)

		if err := cmd.Run(); err != nil {
			log.Printf("Warning: failed to create chunk %d: %v", i, err)
			continue
		}

		// Check if chunk was created and get its size
		if fileInfo, err := os.Stat(chunkFilename); err == nil {
			chunks = append(chunks, ChunkInfo{
				Index:     i,
				Filename:  chunkFilename,
				StartTime: fmt.Sprintf("%d:%02d", startTime/60, startTime%60),
				Duration:  fmt.Sprintf("%d seconds", chunkDuration),
				Size:      fileInfo.Size(),
			})
		}
	}

	return chunks, nil
}

// SaveTranscriptionToFile saves transcription to the transcripts directory
func (s *Service) SaveTranscriptionToFile(transcription, videoURL string) (string, error) {
	// Create descriptive filename from URL and timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Extract a descriptive part from URL
	urlPart := "video"
	if strings.Contains(videoURL, "vimeo.com") {
		parts := strings.Split(videoURL, "/")
		if len(parts) > 0 {
			urlPart = "vimeo_" + parts[len(parts)-1]
		}
	}

	filename := fmt.Sprintf("%s_%s_transcript.txt", timestamp, urlPart)
	filePath := filepath.Join("data", "transcripts", filename)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write transcription to file
	content := fmt.Sprintf("Video URL: %s\nTranscription Date: %s\n\n%s\n",
		videoURL, time.Now().Format("2006-01-02 15:04:05"), transcription)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write transcription file: %w", err)
	}

	return filePath, nil
}

// GetDefaultVideoURL returns the default video URL for testing
func (s *Service) GetDefaultVideoURL() string {
	return "https://player.vimeo.com/video/1031968103"
}
