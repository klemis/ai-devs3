package e05

// CensorRequest represents a request to censor text
type CensorRequest struct {
	Text string
}

// CensorResponse represents the response from text censoring
type CensorResponse struct {
	CensoredText string
	Success      bool
}

// TaskResult represents the final result of the S01E05 task
type TaskResult struct {
	Response     string
	OriginalText string
	CensoredText string
}

// TextData represents the raw text data fetched from the API
type TextData struct {
	Content string
	URL     string
}

// CensoringRule represents a rule for censoring text
type CensoringRule struct {
	Pattern     string
	Replacement string
	Description string
}

// CensoringStats represents statistics about the censoring process
type CensoringStats struct {
	ReplacementCount int
	RulesApplied     []string
	OriginalLength   int
	CensoredLength   int
}
