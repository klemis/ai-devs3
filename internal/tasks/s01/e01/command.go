package e01

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S01E01 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s01e01",
		Short: "Execute S01E01 robot authentication task",
		Long: `S01E01 - Robot Authentication Task

		This task involves:
			1. Fetching a login page from the robot system
			2. Extracting a question from the HTML content
			3. Using OpenAI to answer the question
			4. Submitting the login form with credentials and answer
			5. Extracting the flag from the response

		The robot asks questions that need to be answered correctly to gain access.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
