package video

// TranscriptionResult represents the result of video transcription processing
type TranscriptionResult struct {
	VideoURL      string `json:"video_url"`
	AudioFile     string `json:"audio_file,omitempty"`
	Transcription string `json:"transcription"`
	Duration      string `json:"duration,omitempty"`
	FileSize      int64  `json:"file_size,omitempty"`
	Error         string `json:"error,omitempty"`
}

// VideoData represents downloaded video data with metadata
type VideoData struct {
	Data     []byte
	URL      string
	Size     int64
	Filename string
}

// AudioData represents converted audio data with metadata
type AudioData struct {
	Data     []byte
	Filename string
	Size     int64
	Duration string
	Format   string
}

// TranscriptionRequest represents a request for video transcription
type TranscriptionRequest struct {
	VideoURL        string `json:"video_url"`
	OutputFormat    string `json:"output_format,omitempty"`    // mp3, wav, etc.
	MaxFileSizeMB   int    `json:"max_file_size_mb,omitempty"` // default 25MB
	AudioQuality    string `json:"audio_quality,omitempty"`    // low, medium, high
	SplitIfTooLarge bool   `json:"split_if_too_large,omitempty"`
}

// ChunkInfo represents information about audio chunks when splitting large files
type ChunkInfo struct {
	Index     int    `json:"index"`
	Filename  string `json:"filename"`
	StartTime string `json:"start_time"`
	Duration  string `json:"duration"`
	Size      int64  `json:"size"`
}
