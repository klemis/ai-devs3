package domain

// Question represents a question extracted from the login page
type Question struct {
	Text string
}

// Answer represents an answer from the LLM
type Answer struct {
	Text string
}

// AnswerRoboISO represents an answer from the LLM for RoboISO 2230
type AnswerRoboISO struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

// Credentials represents user login credentials
type Credentials struct {
	Username string
	Password string
}

// Analysis represents the analysis result from the LLM
type AnswerWithAnalysis struct {
	Thinking string `json:"_thinking"`
	Answer   string `json:"answer"`
}

// MapAnalysis is the top-level struct for unmarshaling the entire JSON response.
type MapAnalysis struct {
	Thinking          string      `json:"_thinking"`
	FragmentAnalysis  []Fragment  `json:"fragment_analysis"`
	CandidateAnalysis []Candidate `json:"candidate_analysis"`
	FinalDecision     Decision    `json:"final_decision"`
}

// Fragment holds the extracted features for a single map fragment.
type Fragment struct {
	FragmentID  string   `json:"fragment_id"`
	StreetNames []string `json:"street_names"`
}

// Candidate represents the evaluation of a single potential city,
// containing arguments for and against its selection.
type Candidate struct {
	CityName        string `json:"city_name"`
	EvidenceFor     string `json:"evidence_for"`
	EvidenceAgainst string `json:"evidence_against"`
	OverallFit      string `json:"overall_fit"`
}

// Decision contains the final conclusion of the analysis.
type Decision struct {
	IdentifiedCity string `json:"identified_city"`
	Confidence     string `json:"confidence"`
	Reasoning      string `json:"reasoning"`
}

// ImageProcessingResult represents processed image metadata
type ImageProcessingResult struct {
	Base64Data string
	Width      int
	Height     int
	TokenCost  int
}

// CategorizationResult represents the LLM's categorization response
type CategorizationResult struct {
	Thinking      string `json:"_thinking"`
	Category      string `json:"category"`
	Justification string `json:"justification"`
}
