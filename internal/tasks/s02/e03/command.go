package e03

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S02E03 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s02e03",
		Short: "Execute S02E03 robot image generation task",
		Long: `S02E03 - Robot Image Generation Task

		This task involves:
			1. Fetching a robot description from the centrala API
			2. Analyzing and optimizing the description for DALL-E image generation
			3. Extracting visual keywords and features from the description
			4. Using OpenAI DALL-E 3 to generate a robot image based on the optimized prompt
			5. Submitting the generated image URL to the centrala API

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. OpenAI API access with DALL-E 3 image generation capabilities
			3. Sufficient OpenAI credits for image generation

		The system will:
			1. Validate the robot description for visual content
			2. Optimize the description for better DALL-E results
			3. Generate a high-quality robot image (1024x1024)
			4. Extract and analyze visual features from the description
			5. Submit the image URL as the final answer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for image generation
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
