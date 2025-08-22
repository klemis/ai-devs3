package e02

// TaskResult represents the final result of the S04E02 task
type TaskResult struct {
	CorrectAnswers []string `json:"correct_answers"`
	TotalLines     int      `json:"total_lines"`
	CorrectCount   int      `json:"correct_count"`
	Response       string   `json:"response"`
}

// Constants for the task
const (
	SystemPrompt = "Classify input strings into reliable (1) or unreliable (0). Treat inputs as arbitrary tokens and output only 0 or 1. Do not infer semantics or language."
	TaskName     = "research"
)

// Classification values
const (
	ClassificationReliable   = 1
	ClassificationUnreliable = 0
)
