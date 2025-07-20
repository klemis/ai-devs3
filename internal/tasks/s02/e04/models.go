package e04

import "os"

// FileData represents extracted content from a file
type FileData struct {
	Filename string
	Content  string
	FileType string // "txt", "image", "audio"
	FilePath string
	Size     int64
}

// ProcessingResult represents the result of processing a single file
type ProcessingResult struct {
	FileData    FileData
	ProcessTime float64
	Error       error
	FromCache   bool
}

// CategorizedFiles represents the categorization result
type CategorizedFiles struct {
	People   []string `json:"people"`
	Hardware []string `json:"hardware"`
}

// CategoryResult represents the categorization result for a single file
type CategoryResult struct {
	Filename      string
	Category      string // "people", "hardware", "skip"
	Justification string
	Confidence    float64
}

// TaskResult represents the final result of the S02E04 task
type TaskResult struct {
	Response         string
	Success          bool
	CategorizedFiles *CategorizedFiles
	ProcessingStats  *ProcessingStats
	TotalFiles       int
	CategorizedCount int
}

// ProcessingStats represents statistics about the file processing
type ProcessingStats struct {
	TotalFiles     int
	ProcessedFiles int
	SkippedFiles   int
	ErrorFiles     int
	TextFiles      int
	ImageFiles     int
	AudioFiles     int
	ProcessingTime float64
	CacheHitRate   float64
	PeopleFiles    int
	HardwareFiles  int
}

// FileDirectory represents information about the files directory
type FileDirectory struct {
	Path      string
	FileCount int
	Files     []FileInfo
}

// FileInfo represents basic information about a file
type FileInfo struct {
	Name      string
	Path      string
	Size      int64
	Extension string
	Type      string
}

// CacheManager handles caching for expensive operations
type CacheManager struct {
	BaseDir string
	Enabled bool
}

// ProcessingOptions represents options for file processing
type ProcessingOptions struct {
	MaxWorkers     int
	CacheEnabled   bool
	CacheDir       string
	SupportedTypes []string
	ProcessingDir  string
}

// WorkerPool manages concurrent file processing
type WorkerPool struct {
	NumWorkers int
	FileChan   chan os.DirEntry
	ResultChan chan ProcessingResult
	ErrorChan  chan error
}

// OCRResult represents the result of OCR processing
type OCRResult struct {
	Text       string
	Confidence float64
	HasText    bool
}

// TranscriptionResult represents the result of audio transcription
type TranscriptionResult struct {
	Text       string
	Duration   float64
	Language   string
	Confidence float64
}

// ValidationResult represents the result of file validation
type ValidationResult struct {
	IsValid     bool
	FileType    string
	Issues      []string
	Suggestions []string
	CanProcess  bool
}
