package e01

// Credentials represents user login credentials
type Credentials struct {
	Username string
	Password string
}

// Question represents a question extracted from the login page
type Question struct {
	Text string
}

// Answer represents an answer from the LLM
type Answer struct {
	Text string
}

// LoginRequest represents the data sent to the login endpoint
type LoginRequest struct {
	Username string
	Password string
	Answer   string
}

// LoginResponse represents the response from the login endpoint
type LoginResponse struct {
	Content string
}

// TaskResult represents the final result of the S01E01 task
type TaskResult struct {
	Flag    string
	Content string
}
