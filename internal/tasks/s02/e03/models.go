package e03

// RobotDescription represents the robot description fetched from the API
type RobotDescription struct {
	Description string `json:"description"`
	ID          string `json:"id,omitempty"`
	Source      string `json:"source,omitempty"`
}

// ImageGenerationRequest represents a request to generate an image
type ImageGenerationRequest struct {
	OriginalDescription string
	OptimizedPrompt     string
	Model               string
	Size                string
	Quality             string
}

// ImageGenerationResult represents the result of image generation
type ImageGenerationResult struct {
	ImageURL        string
	GeneratedPrompt string
	Success         bool
	TokensUsed      int
	GenerationTime  float64
}

// TaskResult represents the final result of the S02E03 task
type TaskResult struct {
	Response            string
	GeneratedImageURL   string
	OriginalDescription string
	OptimizedPrompt     string
	GenerationResult    *ImageGenerationResult
}

// PromptOptimization represents the optimization of a description for DALL-E
type PromptOptimization struct {
	Original     string
	Optimized    string
	Improvements []string
	WordCount    int
	CharCount    int
}

// RobotFeatures represents extracted visual features from the description
type RobotFeatures struct {
	Appearance   []string
	Colors       []string
	Materials    []string
	Size         string
	SpecialParts []string
	Movement     string
	Capabilities []string
}

// APIResponse represents the response from the centrala API
type APIResponse struct {
	Content string
	Success bool
	Error   error
}
