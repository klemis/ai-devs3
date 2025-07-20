package e02

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S02E02 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s02e02",
		Short: "Execute S02E02 map analysis task",
		Long: `S02E02 - Map Analysis Task

		This task involves:
			1. Loading map fragments from the images directory
			2. Processing each fragment for AI vision analysis
			3. Using OpenAI Vision to analyze the fragments and identify street names
			4. Evaluating candidate Polish cities based on extracted features
			5. Making a final decision about the most likely city
			6. Submitting the identified city to the centrala API

		The task requires:
			- AI_DEVS_API_KEY environment variable to be set
			- Map fragment images in ../../images/s02e02/ directory
			- OpenAI API access with Vision model capabilities

		The system will:
			- Process multiple map fragments in parallel
			- Extract street names and geographical features
			- Cross-reference with Polish city street patterns
			- Provide detailed analysis with confidence levels
			- Submit the final city identification`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for image processing
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
