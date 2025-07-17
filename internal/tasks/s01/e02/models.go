package e02

// RoboISOMessage represents a message in the RoboISO protocol
type RoboISOMessage struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

// ConversationState represents the current state of the RoboISO conversation
type ConversationState struct {
	CurrentMsgID  int
	IsInitialized bool
}

// TaskResult represents the final result of the S01E02 task
type TaskResult struct {
	FinalResponse string
	Success       bool
	MessageCount  int
}

// VerifyRequest represents a request to the RoboISO verify endpoint
type VerifyRequest struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

// VerifyResponse represents a response from the RoboISO verify endpoint
type VerifyResponse struct {
	Content string
	Success bool
}
