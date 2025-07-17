package e02

import (
	"context"
	"fmt"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the S01E02 RoboISO verification task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
	config     *config.Config
}

// NewService creates a new S01E02 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
		config:     cfg,
	}
}

// InitializeConversation starts the RoboISO conversation with READY message
func (s *Service) InitializeConversation(ctx context.Context, verifyURL string) (*VerifyResponse, error) {
	initMessage := &RoboISOMessage{
		MsgID: 0,
		Text:  "READY",
	}

	response, err := s.sendVerifyRequest(ctx, verifyURL, initMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize conversation: %w", err)
	}

	return response, nil
}

// GetRoboISOAnswer gets an answer from the LLM following RoboISO protocol
func (s *Service) GetRoboISOAnswer(ctx context.Context, question string) (*RoboISOMessage, error) {
	if strings.TrimSpace(question) == "" {
		return nil, errors.NewProcessingError("llm", "roboiso_answer", "question is empty", nil)
	}

	// Use the specialized RoboISO method
	answer, err := s.llmClient.GetAnswerRoboISO(ctx, question)
	if err != nil {
		return nil, fmt.Errorf("failed to get RoboISO answer: %w", err)
	}

	return &RoboISOMessage{
		MsgID: answer.MsgID,
		Text:  answer.Text,
	}, nil
}

// SendVerifyRequest sends a verification request to the RoboISO endpoint
func (s *Service) sendVerifyRequest(ctx context.Context, verifyURL string, message *RoboISOMessage) (*VerifyResponse, error) {
	if message == nil {
		return nil, errors.NewProcessingError("http", "verify_request", "message is nil", nil)
	}

	// Send JSON request
	response, err := s.httpClient.PostJSON(ctx, verifyURL, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send verify request: %w", err)
	}

	// Check if response indicates success (this is a heuristic)
	success := strings.Contains(response, "FLG:") || strings.Contains(response, "flag") || strings.Contains(response, "success")

	return &VerifyResponse{
		Content: response,
		Success: success,
	}, nil
}

// ExecuteTask executes the complete S01E02 RoboISO verification task
func (s *Service) ExecuteTask(ctx context.Context, verifyURL string) (*TaskResult, error) {
	messageCount := 0

	// Step 1: Initialize conversation
	response, err := s.InitializeConversation(ctx, verifyURL)
	if err != nil {
		return nil, errors.NewTaskError("s01e02", "initialize_conversation", err)
	}
	messageCount++

	// Step 2: Get answer from LLM
	answer, err := s.GetRoboISOAnswer(ctx, response.Content)
	if err != nil {
		return nil, errors.NewTaskError("s01e02", "get_roboiso_answer", err)
	}

	// Step 3: Send answer back to verify endpoint
	finalResponse, err := s.sendVerifyRequest(ctx, verifyURL, answer)
	if err != nil {
		return nil, errors.NewTaskError("s01e02", "send_verify_request", err)
	}
	messageCount++

	return &TaskResult{
		FinalResponse: finalResponse.Content,
		Success:       finalResponse.Success,
		MessageCount:  messageCount,
	}, nil
}
