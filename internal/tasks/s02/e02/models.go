package e02

// MapFragment represents a single map fragment to be analyzed
type MapFragment struct {
	ID         string
	Path       string
	Base64Data string
	Width      int
	Height     int
	TokenCost  int
}

// FragmentAnalysis represents the analysis of a single map fragment
type FragmentAnalysis struct {
	FragmentID  string   `json:"fragment_id"`
	StreetNames []string `json:"street_names"`
}

// CandidateCity represents a candidate city evaluation
type CandidateCity struct {
	CityName        string `json:"city_name"`
	EvidenceFor     string `json:"evidence_for"`
	EvidenceAgainst string `json:"evidence_against"`
	OverallFit      string `json:"overall_fit"`
}

// CityDecision represents the final decision about the identified city
type CityDecision struct {
	IdentifiedCity string `json:"identified_city"`
	Confidence     string `json:"confidence"`
	Reasoning      string `json:"reasoning"`
}

// MapAnalysisResult represents the complete analysis result
type MapAnalysisResult struct {
	Thinking          string             `json:"_thinking"`
	FragmentAnalysis  []FragmentAnalysis `json:"fragment_analysis"`
	CandidateAnalysis []CandidateCity    `json:"candidate_analysis"`
	FinalDecision     CityDecision       `json:"final_decision"`
}

// TaskResult represents the final result of the S02E02 task
type TaskResult struct {
	IdentifiedCity string
	Confidence     string
	FragmentCount  int
	AnalysisResult *MapAnalysisResult
}

// ImageProcessingStats represents statistics about image processing
type ImageProcessingStats struct {
	TotalFragments   int
	ProcessedImages  int
	TotalTokenCost   int
	ProcessingTime   float64
	AverageTokenCost int
}

// FragmentDirectory represents information about the fragment directory
type FragmentDirectory struct {
	Path          string
	FragmentCount int
	Fragments     []MapFragment
}

// ProcessingOptions represents options for image processing
type ProcessingOptions struct {
	MaxDimension    int
	Quality         string
	CacheEnabled    bool
	ParallelProcess bool
}
