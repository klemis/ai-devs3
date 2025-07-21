package e03

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S03E03 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s03e03",
		Short: "Execute S03E03 database query task",
		Long: `S03E03 - Database Query Task

		This task involves:
			1. Querying the database API to discover available tables
			2. Getting table schemas using SHOW CREATE TABLE commands
			3. Analyzing relationships between tables (users, datacenters, connections)
			4. Using LLM to generate SQL query for finding active datacenters with inactive managers
			5. Executing the generated query against the database API
			6. Extracting datacenter IDs from query results
			7. Submitting the datacenter IDs array to the centrala API

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. OpenAI API access for SQL query generation
			3. Access to the database API at https://c3ntrala.ag3nts.org/apidb

		The command will:
			1. Execute SHOW TABLES to discover database structure
			2. Get CREATE TABLE statements for all discovered tables
			3. Generate context-aware SQL query using LLM
			4. Find active datacenters managed by inactive managers
			5. Return datacenter IDs as an array of numbers
			6. Provide detailed processing statistics and query information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for database processing
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
