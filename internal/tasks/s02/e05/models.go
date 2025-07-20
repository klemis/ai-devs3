package e05

// ArxivContent represents the consolidated content from the article
type ArxivContent struct {
	Text              string            `json:"text"`
	ImageDescriptions map[string]string `json:"image_descriptions"`
	AudioTranscripts  map[string]string `json:"audio_transcripts"`
}

// ImageInfo represents an image with its context
type ImageInfo struct {
	URL     string
	Caption string
	Alt     string
}

// ArxivAnswer represents the answer structure for the arxiv task
type ArxivAnswer map[string]string

// Question represents a parsed question with ID and text
type Question struct {
	ID   string
	Text string
}

// TaskResult represents the final result of the S02E05 task
type TaskResult struct {
	Response        string
	Answers         ArxivAnswer
	ProcessingStats *ProcessingStats
	TotalQuestions  int
}

// ProcessingStats represents statistics about the arxiv processing
type ProcessingStats struct {
	TotalQuestions    int
	AnsweredQuestions int
	ProcessingTime    float64
	ContentLength     int
	ImagesProcessed   int
	AudioProcessed    int
	CacheHitRate      float64
}
