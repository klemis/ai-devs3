package e02

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S03E02 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s03e02",
		Short: "Execute S03E02 weapon reports vector search task",
		Long: `S03E02 - Weapon Reports Vector Search Task

		This task involves:
			1. Setting up a Qdrant vector database collection for weapon test reports
			2. Processing all .txt files from the do-not-share directory
			3. Extracting dates from filenames (format: YYYY_MM_DD.txt)
			4. Generating embeddings using OpenAI's text-embedding-3-large model
			5. Storing report embeddings in Qdrant with metadata (date, filename, content)
			6. Searching for reports mentioning weapon prototype theft
			7. Submitting the date of the report containing theft mention

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. OPENAI_API_KEY environment variable to be set
			3. Qdrant vector database cluster running
			4. Files directory at ../lessons-md/pliki_z_fabryki/do-not-share with report files
			5. Sufficient memory and processing power for embedding generation

		The command will:
			1. Create/recreate a Qdrant collection with 3072 dimensions
			2. Process all weapon test report files from the target directory
			3. Generate embeddings for each report using text-embedding-3-large
			4. Store embeddings with proper metadata in the vector database
			5. Search for the query about weapon prototype theft
			6. Return the date from the most relevant report
			7. Submit the answer to the centrala API

		Vector Configuration:
			- Model: text-embedding-3-large (3072 dimensions)
			- Distance: Cosine similarity
			- Collection: weapon_reports`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for vector processing
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
