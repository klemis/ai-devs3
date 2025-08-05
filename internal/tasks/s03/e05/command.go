package e05

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S03E05 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s03e05",
		Short: "Execute S03E05 connections task - find shortest path using Neo4j",
		Long: `S03E05 - Connections Task (Graph Database)

		This task involves:
			1. Retrieving users and connections data from MySQL database (reusing S03E03 logic)
			2. Setting up Neo4j graph database connection
			3. Loading users as Person nodes and connections as KNOWS relationships
			4. Finding the shortest path between Rafał and Barbara using Cypher queries
			5. Submitting the path as a comma-separated string to the centrala API

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. Neo4j database connection (NEO4J_URI, NEO4J_USER, NEO4J_PASSWORD)
			3. Access to the database API at https://c3ntrala.ag3nts.org/apidb

		Key implementation details:
			- MySQL data is cached locally for efficiency
			- Neo4j nodes use 'userId' property (not 'id' to avoid conflicts)
			- Unidirectional KNOWS relationships are created
			- Shortest path algorithm finds optimal route between users
			- Result format: "Rafał,Name1,Name2,Barbara"

		The command will:
			1. Connect to both MySQL (via API) and Neo4j databases
			2. Retrieve and cache user/connection data from MySQL
			3. Clear and populate Neo4j graph with fresh data
			4. Execute shortest path query between Rafał and Barbara
			5. Format and submit the result to the report endpoint
			6. Provide detailed statistics and processing information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for graph processing
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
