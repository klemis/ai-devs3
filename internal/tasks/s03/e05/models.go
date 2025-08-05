package e05

// User represents a user from the MySQL database
type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

// Connection represents a connection between two users
type Connection struct {
	User1ID int `json:"user1_id"`
	User2ID int `json:"user2_id"`
}

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

// GraphData contains the users and connections data retrieved from MySQL
type GraphData struct {
	Users       []User       `json:"users"`
	Connections []Connection `json:"connections"`
}

// TaskResult represents the final result of the S03E05 task
type TaskResult struct {
	Response       string
	ShortestPath   []string
	PathString     string
	ProcessingTime float64
	Stats          ProcessingStats
}

// ProcessingStats represents statistics about the graph processing
type ProcessingStats struct {
	UsersLoaded          int
	ConnectionsLoaded    int
	NodesCreated         int
	RelationshipsCreated int
	PathFound            bool
	ProcessingTime       float64
}
