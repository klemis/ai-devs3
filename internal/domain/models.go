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
