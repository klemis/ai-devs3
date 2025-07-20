package e01

import "os"

// AudioFile represents an audio file to be transcribed
type AudioFile struct {
	Name     string
	Path     string
	File     *os.File
	Size     int64
	Duration float64 // in seconds
}

// Transcript represents a transcribed audio file
type Transcript struct {
	AudioFile string
	Text      string
	Length    int
	Success   bool
}

// TranscriptAnalysis represents the analysis of transcripts
type TranscriptAnalysis struct {
	Thinking string `json:"_thinking"`
	Answer   string `json:"answer"`
}

// TaskResult represents the final result of the S02E01 task
type TaskResult struct {
	Response         string
	Success          bool
	TranscriptCount  int
	TotalTranscripts string
	Analysis         *TranscriptAnalysis
}

// AudioProcessingStats represents statistics about audio processing
type AudioProcessingStats struct {
	TotalFiles       int
	ProcessedFiles   int
	FailedFiles      int
	TotalDuration    float64
	ProcessingTime   float64
	TranscriptLength int
}

// AudioDirectory represents information about the audio directory
type AudioDirectory struct {
	Path      string
	FileCount int
	Files     []AudioFile
}

// ProcessingResult represents the result of processing a single audio file
type ProcessingResult struct {
	AudioFile  string
	Transcript *Transcript
	Error      error
	Duration   float64
}
