package e03

// DatabaseRequest represents the request structure for the database API
type DatabaseRequest struct {
	Task   string `json:"task"`
	APIKey string `json:"apikey"`
	Query  string `json:"query"`
}

// DatabaseResponse represents the response from the database API
type DatabaseResponse struct {
	Reply []map[string]any `json:"reply"`
	Error string           `json:"error,omitempty"`
}

// TableSchema represents a table's structure information
type TableSchema struct {
	Name          string
	CreateSQL     string
	Columns       []string
	Relationships []string
}

// DatabaseInfo contains information about the database structure
type DatabaseInfo struct {
	Tables  []string
	Schemas map[string]*TableSchema
}

// TaskResult represents the final result of the S03E03 task
type TaskResult struct {
	Response       string
	DatacenterIDs  []int
	GeneratedQuery string
	ProcessingTime float64
}

// ProcessingStats represents statistics about the database processing
type ProcessingStats struct {
	TablesDiscovered int
	SchemasAnalyzed  int
	QueryGenerated   bool
	ResultsFound     int
	ProcessingTime   float64
}
