package e03

import (
	"context"
	"fmt"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the S02E03 robot image generation task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
	config     *config.Config
}

// NewService creates a new S02E03 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
		config:     cfg,
	}
}

// FetchRobotDescription fetches the robot description from the centrala API
func (s *Service) FetchRobotDescription(ctx context.Context, apiKey string) (*RobotDescription, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.NewProcessingError("http", "fetch_robot_description", "API key is empty", nil)
	}

	url := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/robotid.json", apiKey)

	data, err := s.httpClient.FetchJSONData(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch robot description: %w", err)
	}

	// Extract the description from the JSON response
	description, ok := data["description"].(string)
	if !ok {
		return nil, errors.NewProcessingError("json", "fetch_robot_description", "description field not found or not a string", nil)
	}

	return &RobotDescription{
		Description: description,
		Source:      url,
	}, nil
}

// OptimizeDescriptionForDALLE optimizes the robot description for DALL-E image generation
func (s *Service) OptimizeDescriptionForDALLE(ctx context.Context, description string) (*PromptOptimization, error) {
	if strings.TrimSpace(description) == "" {
		return nil, errors.NewProcessingError("llm", "optimize_description", "description is empty", nil)
	}

	optimizedPrompt, err := s.llmClient.ExtractKeywordsForDALLE(ctx, description)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize description for DALL-E: %w", err)
	}

	return &PromptOptimization{
		Original:  description,
		Optimized: optimizedPrompt,
		WordCount: len(strings.Fields(optimizedPrompt)),
		CharCount: len(optimizedPrompt),
	}, nil
}

// GenerateRobotImage generates an image using DALL-E based on the optimized prompt
func (s *Service) GenerateRobotImage(ctx context.Context, optimizedPrompt string) (*ImageGenerationResult, error) {
	if strings.TrimSpace(optimizedPrompt) == "" {
		return nil, errors.NewProcessingError("llm", "generate_image", "optimized prompt is empty", nil)
	}

	imageURL, err := s.llmClient.GenerateImage(ctx, optimizedPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image with DALL-E: %w", err)
	}

	return &ImageGenerationResult{
		ImageURL:        imageURL,
		GeneratedPrompt: optimizedPrompt,
		Success:         true,
	}, nil
}

// SubmitImageURL submits the generated image URL to the centrala API
func (s *Service) SubmitImageURL(ctx context.Context, apiKey string, imageURL string) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("robotid", apiKey, imageURL)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit image URL: %w", err)
	}

	return result, nil
}

// ExecuteTask executes the complete S02E03 task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	// Step 1: Fetch robot description
	robotDesc, err := s.FetchRobotDescription(ctx, apiKey)
	if err != nil {
		return nil, errors.NewTaskError("s02e03", "fetch_robot_description", err)
	}

	// Step 2: Optimize description for DALL-E
	optimization, err := s.OptimizeDescriptionForDALLE(ctx, robotDesc.Description)
	if err != nil {
		return nil, errors.NewTaskError("s02e03", "optimize_description", err)
	}

	// Step 3: Generate image
	imageResult, err := s.GenerateRobotImage(ctx, optimization.Optimized)
	if err != nil {
		return nil, errors.NewTaskError("s02e03", "generate_image", err)
	}

	// Step 4: Submit image URL
	response, err := s.SubmitImageURL(ctx, apiKey, imageResult.ImageURL)
	if err != nil {
		return nil, errors.NewTaskError("s02e03", "submit_image_url", err)
	}

	return &TaskResult{
		Response:            response,
		GeneratedImageURL:   imageResult.ImageURL,
		OriginalDescription: robotDesc.Description,
		OptimizedPrompt:     optimization.Optimized,
		GenerationResult:    imageResult,
	}, nil
}

// PrintGenerationDetails prints detailed information about the generation process
func (s *Service) PrintGenerationDetails(result *TaskResult) {
	fmt.Println("=== Robot Image Generation Details ===")
	fmt.Println()

	fmt.Println("[Original Description]")
	fmt.Printf("  %s\n", result.OriginalDescription)
	fmt.Println()

	fmt.Println("[Optimized DALL-E Prompt]")
	fmt.Printf("  %s\n", result.OptimizedPrompt)
	fmt.Println()

	fmt.Println("[Generated Image]")
	fmt.Printf("  URL: %s\n", result.GeneratedImageURL)
	fmt.Println()

	fmt.Println("=====================================")
}
