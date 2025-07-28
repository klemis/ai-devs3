package e04

// BarbaraSearchRequest represents the request structure for the people/places API
type BarbaraSearchRequest struct {
	APIKey string `json:"apikey"`
	Query  string `json:"query"`
}

// BarbaraSearchResponse represents the response from the people/places API
type BarbaraSearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ParsedData contains the extracted names and cities from barbara.txt
type ParsedData struct {
	Names  []string `json:"names"`
	Cities []string `json:"cities"`
}

// SearchState tracks the BFS search progress
type SearchState struct {
	QueueNames      []string
	QueueCities     []string
	VisitedNames    map[string]struct{}
	VisitedCities   map[string]struct{}
	OriginalCities  map[string]struct{}
	RequestCount    int
	BarbaraLocation string
}

// TaskResult represents the final result of the S03E04 task
type TaskResult struct {
	Response         string
	BarbaraLocation  string
	TotalRequests    int
	ProcessingTime   float64
	OriginalCities   []string
	DiscoveredCities []string
}

// APIResponse represents the final response to centrala
type APIResponse struct {
	Task   string `json:"task"`
	APIKey string `json:"apikey"`
	Answer string `json:"answer"`
}
