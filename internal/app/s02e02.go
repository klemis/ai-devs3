package app

import (
	"fmt"
	"log"
	"path/filepath"
)

// RunS02E02 handles the map fragment analysis task
func (app *App) RunS02E02() (string, error) {
	// imageURL := "https://assets-v2.circle.so/4s5ldjdx0ta03r9aey61z0k4uw1c"

	// log.Printf("Fetching map image from: %s", imageURL)

	// // Step 1: Fetch the image
	// TODO: Split the image into fragments
	// imageData, err := app.httpClient.FetchImage(imageURL)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to fetch map image: %w", err)
	// }

	// Step 1: Load image parts
	// Step 2: Process the full image for single analysis
	const numImages = 4
	images := make([]string, numImages)
	for i := 0; i < numImages; i++ {
		path, err := filepath.Abs(fmt.Sprintf("../../images/s02e02/fragment%d.png", i+1))
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path: %w", err)
		}

		processedImage, err := app.imageProcessor.ProcessImage(path, 2048)
		if err != nil {
			return "", fmt.Errorf("failed to process image: %w", err)
		}
		images[i] = processedImage.Base64Data

		log.Printf("Processed image: %dx%d, estimated tokens: %d", processedImage.Width, processedImage.Height, processedImage.TokenCost)
	}

	// Step 3: Analyze as single image
	log.Println("Analyzing map fragments...")
	analysis, err := app.llmClient.AnalyzeMapFragments(images)
	if err != nil {
		return "", fmt.Errorf("failed to analyze map fragments: %w", err)
	}

	fmt.Println("--- Map Analysis Result ---")
	fmt.Println()

	fmt.Println("[AI Thinking Process]")
	fmt.Printf("  %s\n", analysis.Thinking)
	fmt.Println()

	fmt.Println("[Individual Fragment Analysis]")
	for _, fragment := range analysis.FragmentAnalysis {
		fmt.Printf("  Fragment ID: %s\n", fragment.FragmentID)
		fmt.Printf("    Streets: %v\n", fragment.StreetNames)
	}
	fmt.Println()

	fmt.Println("[Candidate City Evaluation]")
	for _, candidate := range analysis.CandidateAnalysis {
		fmt.Printf("  City: %s (Overall Fit: %s)\n", candidate.CityName, candidate.OverallFit)
		fmt.Printf("    Evidence For: %s\n", candidate.EvidenceFor)
		fmt.Printf("    Evidence Against: %s\n", candidate.EvidenceAgainst)
	}
	fmt.Println()

	fmt.Println("[Final Decision]")
	fmt.Printf("  Identified City: %s\n", analysis.FinalDecision.IdentifiedCity)
	fmt.Printf("  Confidence Level: %s\n", analysis.FinalDecision.Confidence)
	fmt.Printf("  Reasoning: %s\n", analysis.FinalDecision.Reasoning)
	fmt.Println()

	fmt.Println("-------------------------")

	cityAnswer := analysis.FinalDecision.IdentifiedCity

	return cityAnswer, nil
}
