package e01

// DocumentsAnswer represents the answer structure for the documents task
type DocumentsAnswer map[string]string

// Person represents a person with their attributes
type Person struct {
	Name       string   `json:"name"`
	Profession string   `json:"profession"`
	Skills     []string `json:"skills"`
	Location   string   `json:"location"`
	Status     string   `json:"status"`
	Relations  []string `json:"relations"`
}

// FactsKeywords represents processed facts with key information
type FactsKeywords struct {
	People   []Person `json:"people"`
	Sectors  []string `json:"sectors"`
	Keywords []string `json:"keywords"`
}

// ProcessedFacts represents the structure of processed facts cache
type ProcessedFacts map[string]FactsKeywords

// Report represents a security report file
type Report struct {
	Filename string
	Content  string
	Keywords string
}

// TaskResult represents the final result of the S03E01 task
type TaskResult struct {
	Response          string
	DocumentsAnswer   DocumentsAnswer
	ProcessingStats   *ProcessingStats
	TotalFiles        int
	KeywordsGenerated int
}

// ProcessingStats represents statistics about the documents processing
type ProcessingStats struct {
	TotalFiles        int
	ProcessedFiles    int
	SkippedFiles      int
	ErrorFiles        int
	ProcessingTime    float64
	FactsFilesLoaded  int
	CacheHitRate      float64
	KeywordsGenerated int
	AverageKeywords   float64
}
