package e03

// TestQuestion represents a question in the test data
type TestQuestion struct {
	Question string           `json:"question,omitempty"`
	Answer   any              `json:"answer,omitempty"`
	Test     *TestSubQuestion `json:"test,omitempty"`
	Extra    map[string]any   `json:"-"` // For any additional fields
}

// TestSubQuestion represents a nested test question
type TestSubQuestion struct {
	Question string `json:"q"`
	Answer   string `json:"a"`
}

// TestData represents the structure of the JSON data from the API
type TestData struct {
	APIKey   string         `json:"apikey"`
	TestData []TestQuestion `json:"test-data"`
}

// ProcessedResult represents the result after processing the test data
type ProcessedResult struct {
	APIKey   string         `json:"apikey"`
	TestData []TestQuestion `json:"test-data"`
}

// TaskResult represents the final result of the S01E03 task
type TaskResult struct {
	Response    string
	Corrected   int // Number of corrections made
	LLMAnswers  int // Number of LLM answers provided
	MathAnswers int // Number of math answers calculated
}

// MathOperation represents a mathematical operation
type MathOperation struct {
	Left      int
	Right     int
	Operation string
	Result    int
}

// LLMBatch represents a batch of questions sent to the LLM
type LLMBatch struct {
	Questions []string
	Indexes   []int
}

// DataResponse represents the response from the centrala API
type DataResponse struct {
	Content string
	Success bool
}
