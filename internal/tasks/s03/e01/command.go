package e01

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S03E01 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s03e01",
		Short: "Execute S03E01 security reports processing task",
		Long: `S03E01 - Security Reports Processing Task

		This task involves:
			1. Scanning the pliki_z_fabryki directory for .txt security report files
			2. Processing the facts folder to extract key information about people, sectors, and keywords
			3. Using LLM to cross-reference report content with facts database
			4. Generating comprehensive Polish keywords for each report
			5. Implementing caching for processed facts to improve performance
			6. Submitting the keyword mappings to the centrala API

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. Files directory at ../lessons-md/pliki_z_fabryki with report files
			3. Facts directory at ../lessons-md/pliki_z_fabryki/facts with reference data
			4. OpenAI API access for natural language processing
			5. Sufficient disk space for caching processed facts

		The command will:
			1. Process exactly 10 .txt report files
			2. Extract and cache structured information from facts files
			3. Generate context-aware Polish keywords for each report
			4. Cross-reference people mentioned in reports with facts database
			5. Include profession, skills, and technology keywords from facts
			6. Provide detailed processing statistics and caching information`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for document processing
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
