package e02

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"ai-devs3/internal/config"
	"ai-devs3/internal/http"
	"ai-devs3/internal/image"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the S02E02 map analysis task
type Service struct {
	httpClient     *http.Client
	llmClient      *openai.Client
	imageProcessor *image.Processor
	config         *config.Config
}

// NewService creates a new S02E02 service
func NewService(cfg *config.Config, httpClient *http.Client, llmClient *openai.Client, imageProcessor *image.Processor) *Service {
	return &Service{
		httpClient:     httpClient,
		llmClient:      llmClient,
		imageProcessor: imageProcessor,
		config:         cfg,
	}
}

// LoadMapFragments loads map fragments from the specified directory
func (s *Service) LoadMapFragments(ctx context.Context, fragmentsDir string, numFragments int) ([]MapFragment, error) {
	var fragments []MapFragment

	for i := 1; i <= numFragments; i++ {
		fragmentPath := filepath.Join(fragmentsDir, fmt.Sprintf("fragment%d.png", i))

		// Check if file exists
		if err := s.imageProcessor.ValidateImagePath(fragmentPath); err != nil {
			return nil, fmt.Errorf("failed to validate fragment %d: %w", i, err)
		}

		// Get absolute path
		absPath, err := filepath.Abs(fragmentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get absolute path for fragment %d: %w", i, err)
		}

		fragments = append(fragments, MapFragment{
			ID:   fmt.Sprintf("fragment_%d", i),
			Path: absPath,
		})
	}

	return fragments, nil
}

// ProcessMapFragments processes all map fragments for analysis
func (s *Service) ProcessMapFragments(ctx context.Context, fragments []MapFragment, maxDimension int) ([]MapFragment, error) {
	var processedFragments []MapFragment

	for i, fragment := range fragments {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		// Process the image
		result, err := s.imageProcessor.ProcessImage(fragment.Path, maxDimension)
		if err != nil {
			return nil, fmt.Errorf("failed to process fragment %s: %w", fragment.ID, err)
		}

		processedFragment := MapFragment{
			ID:         fragment.ID,
			Path:       fragment.Path,
			Base64Data: result.Base64Data,
			Width:      result.Width,
			Height:     result.Height,
			TokenCost:  result.TokenCost,
		}

		processedFragments = append(processedFragments, processedFragment)

		// Log progress
		fmt.Printf("Processed fragment %d/%d: %s (%dx%d, tokens: %d)\n",
			i+1, len(fragments), fragment.ID, result.Width, result.Height, result.TokenCost)
	}

	return processedFragments, nil
}

// AnalyzeMapFragments analyzes processed map fragments to identify the city
func (s *Service) AnalyzeMapFragments(ctx context.Context, fragments []MapFragment) (*MapAnalysisResult, error) {
	if len(fragments) == 0 {
		return nil, errors.NewProcessingError("analysis", "analyze_fragments", "no fragments to analyze", nil)
	}

	// Extract base64 data for analysis
	var imagesBase64 []string
	for _, fragment := range fragments {
		if fragment.Base64Data == "" {
			return nil, errors.NewProcessingError("analysis", "analyze_fragments",
				fmt.Sprintf("fragment %s has no base64 data", fragment.ID), nil)
		}
		imagesBase64 = append(imagesBase64, fragment.Base64Data)
	}

	// Analyze using OpenAI
	analysis, err := s.llmClient.AnalyzeMapFragments(ctx, imagesBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze map fragments: %w", err)
	}

	// Convert to our domain model
	result := &MapAnalysisResult{
		Thinking:          analysis.Thinking,
		FragmentAnalysis:  make([]FragmentAnalysis, len(analysis.FragmentAnalysis)),
		CandidateAnalysis: make([]CandidateCity, len(analysis.CandidateAnalysis)),
		FinalDecision: CityDecision{
			IdentifiedCity: analysis.FinalDecision.IdentifiedCity,
			Confidence:     analysis.FinalDecision.Confidence,
			Reasoning:      analysis.FinalDecision.Reasoning,
		},
	}

	// Convert fragment analysis
	for i, fragment := range analysis.FragmentAnalysis {
		result.FragmentAnalysis[i] = FragmentAnalysis{
			FragmentID:  fragment.FragmentID,
			StreetNames: fragment.StreetNames,
		}
	}

	// Convert candidate analysis
	for i, candidate := range analysis.CandidateAnalysis {
		result.CandidateAnalysis[i] = CandidateCity{
			CityName:        candidate.CityName,
			EvidenceFor:     candidate.EvidenceFor,
			EvidenceAgainst: candidate.EvidenceAgainst,
			OverallFit:      candidate.OverallFit,
		}
	}

	return result, nil
}

// PrintAnalysisResult prints the analysis result in a formatted way
func (s *Service) PrintAnalysisResult(result *MapAnalysisResult) {
	fmt.Println("--- Map Analysis Result ---")
	fmt.Println()

	fmt.Println("[AI Thinking Process]")
	fmt.Printf("  %s\n", result.Thinking)
	fmt.Println()

	fmt.Println("[Individual Fragment Analysis]")
	for _, fragment := range result.FragmentAnalysis {
		fmt.Printf("  Fragment ID: %s\n", fragment.FragmentID)
		fmt.Printf("    Streets: %v\n", fragment.StreetNames)
	}
	fmt.Println()

	fmt.Println("[Candidate City Evaluation]")
	for _, candidate := range result.CandidateAnalysis {
		fmt.Printf("  City: %s (Overall Fit: %s)\n", candidate.CityName, candidate.OverallFit)
		fmt.Printf("    Evidence For: %s\n", candidate.EvidenceFor)
		fmt.Printf("    Evidence Against: %s\n", candidate.EvidenceAgainst)
	}
	fmt.Println()

	fmt.Println("[Final Decision]")
	fmt.Printf("  Identified City: %s\n", result.FinalDecision.IdentifiedCity)
	fmt.Printf("  Confidence Level: %s\n", result.FinalDecision.Confidence)
	fmt.Printf("  Reasoning: %s\n", result.FinalDecision.Reasoning)
	fmt.Println()

	fmt.Println("-------------------------")
}

// ExecuteTask executes the complete S02E02 task workflow
func (s *Service) ExecuteTask(ctx context.Context, fragmentsDir string, numFragments int, apiKey string) (*TaskResult, error) {
	// Step 1: Load map fragments
	fragments, err := s.LoadMapFragments(ctx, fragmentsDir, numFragments)
	if err != nil {
		return nil, errors.NewTaskError("s02e02", "load_fragments", err)
	}

	// Step 2: Process map fragments
	processedFragments, err := s.ProcessMapFragments(ctx, fragments, 2048)
	if err != nil {
		return nil, errors.NewTaskError("s02e02", "process_fragments", err)
	}

	// Step 3: Analyze map fragments
	analysisResult, err := s.AnalyzeMapFragments(ctx, processedFragments)
	if err != nil {
		return nil, errors.NewTaskError("s02e02", "analyze_fragments", err)
	}

	return &TaskResult{
		IdentifiedCity: analysisResult.FinalDecision.IdentifiedCity,
		Confidence:     analysisResult.FinalDecision.Confidence,
		FragmentCount:  len(processedFragments),
		AnalysisResult: analysisResult,
	}, nil
}

// GetProcessingStats returns statistics about the image processing
func (s *Service) GetProcessingStats(fragments []MapFragment) *ImageProcessingStats {
	stats := &ImageProcessingStats{
		TotalFragments:  len(fragments),
		ProcessedImages: 0,
		TotalTokenCost:  0,
	}

	for _, fragment := range fragments {
		if fragment.Base64Data != "" {
			stats.ProcessedImages++
			stats.TotalTokenCost += fragment.TokenCost
		}
	}

	if stats.ProcessedImages > 0 {
		stats.AverageTokenCost = stats.TotalTokenCost / stats.ProcessedImages
	}

	return stats
}

// ValidateFragmentsDirectory validates that the fragments directory exists and contains the expected files
func (s *Service) ValidateFragmentsDirectory(fragmentsDir string, numFragments int) error {
	// Check if directory exists
	if _, err := os.Stat(fragmentsDir); os.IsNotExist(err) {
		return errors.NewProcessingError("filesystem", "validate_directory",
			fmt.Sprintf("fragments directory does not exist: %s", fragmentsDir), err)
	}

	// Check if all expected fragment files exist
	for i := 1; i <= numFragments; i++ {
		fragmentPath := filepath.Join(fragmentsDir, fmt.Sprintf("fragment%d.png", i))
		if _, err := os.Stat(fragmentPath); os.IsNotExist(err) {
			return errors.NewProcessingError("filesystem", "validate_directory",
				fmt.Sprintf("fragment file does not exist: %s", fragmentPath), err)
		}
	}

	return nil
}
