package e04

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S03E04 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s03e04",
		Short: "Execute S03E04 Barbara search task",
		Long: `S03E04 - Barbara Search Task ("loop")

This task involves:
	1. Reading barbara.txt from https://c3ntrala.ag3nts.org/dane/barbara.txt
	2. Using LLM to parse text and extract all names and cities (normalized, uppercase)
	3. Performing BFS search using /people and /places endpoints:
		- POST to https://c3ntrala.ag3nts.org/people with JSON {"apikey":"KEY", "query":"NAME"}
		- POST to https://c3ntrala.ag3nts.org/places with JSON {"apikey":"KEY", "query":"CITY"}
	4. Stopping when /places result contains "BARBARA" for a city not in original note
	5. Submitting Barbara's current location to /report endpoint

The task requires:
	1. AI_DEVS_API_KEY environment variable to be set
	2. OpenAI API access for text parsing and extraction
	3. Rate limiting compliance (max 5 requests per second)

The command will:
	1. Parse the initial note to get starting names and cities
	2. Use breadth-first search to explore connections
	3. Track visited entities to avoid cycles
	4. Log all API queries and responses for transparency
	5. Find Barbara's current location (different from original note)
	6. Provide detailed search statistics and processing information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for Barbara search processing
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
