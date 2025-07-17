package e05

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S01E05 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s01e05",
		Short: "Execute S01E05 text censoring task",
		Long: `S01E05 - Text Censoring Task

		This task involves:
			1. Fetching raw text data from the centrala API using your API key
			2. Using Ollama local LLM to censor personal information in the text
			3. Applying specific censoring rules for names, ages, cities, and addresses
			4. Submitting the censored text back to the centrala API
			5. Receiving confirmation of successful censoring

		The task requires:
			- AI_DEVS_API_KEY environment variable to be set
			- Ollama server running locally (default: http://localhost:11434)
			- Ollama model installed (default: llama3.2)

		Censoring rules:
			- Names and surnames are replaced together as one "CENZURA"
			- Ages are replaced with "CENZURA"
			- City names are replaced with "CENZURA"
			- Street addresses are replaced with "CENZURA" after street indicators`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
