package e03

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S01E03 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s01e03",
		Short: "Execute S01E03 JSON data processing task",
		Long: `S01E03 - JSON Data Processing Task

		This task involves:
			1. Fetching JSON test data from the centrala API using your API key
			2. Processing the data by solving math problems and answering questions
			3. Math problems are solved automatically using regex parsing
			4. Questions are answered using OpenAI's LLM in batch mode for efficiency
			5. Submitting the corrected data back to the centrala API
			6. Receiving confirmation of successful processing

		The task requires the AI_DEVS_API_KEY environment variable to be set.
		The system processes both direct math questions and nested test questions.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
