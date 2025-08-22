package e02

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S04E02 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s04e02",
		Short: "Execute S04E02 text classification research task",
		Long: `S04E02 - Text Classification Research Task

This task involves:
	1. Reading verification lines from data/s04e02/verify.txt
	2. Classifying each line using a fine-tuned model
	3. Using exact system prompt for classification requests
	4. Collecting IDs of lines classified as reliable (1)
	5. Reporting correct answer IDs to central endpoint

The task requires:
	1. AI_DEVS_API_KEY environment variable to be set
	2. Fine-tuned model endpoint for classification (TODO: to be configured)
	3. Exact system prompt matching training data format

Classification Process:
	- Each line is sent with exact system prompt
	- Model responds with "0" (unreliable) or "1" (reliable)
	- Only lines classified as "1" are included in final answer
	- Response format matches training data structure

The command will:
	1. Parse verify.txt and extract content after "ID=" format
	2. Send classification requests with exact message structure
	3. Collect IDs of reliable classifications (zero-padded)
	4. Report to https://c3ntrala.ag3nts.org/report endpoint
	5. Return success/failure status

System Prompt (exact):
"Classify input strings into reliable (1) or unreliable (0). Treat inputs as arbitrary tokens and output only 0 or 1. Do not infer semantics or language."`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for classification processing
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
