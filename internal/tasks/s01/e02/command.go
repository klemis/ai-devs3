package e02

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S01E02 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s01e02",
		Short: "Execute S01E02 RoboISO verification task",
		Long: `S01E02 - RoboISO Verification Task

		This task involves:
			1. Initializing communication with RoboISO system using "READY" message
			2. Receiving a question from the RoboISO system
			3. Using OpenAI to generate a response following RoboISO 2230 protocol
			4. Sending the response back to verify the communication
			5. Extracting the final result

		The RoboISO system follows a specific protocol with JSON messages containing msgID and text fields.
		The system has deliberate misinformation for security testing purposes.`,
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
