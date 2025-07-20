package e05

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S02E05 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s02e05",
		Short: "Execute S02E05 arxiv document analysis task",
		Long: `S02E05 - Arxiv Document Analysis Task

		This task involves:
			1. Fetching Professor Maj's intercepted research article from the centrala system
			2. Processing HTML content to extract text, images, and audio files
			3. Analyzing images using OpenAI vision capabilities
			4. Transcribing audio files using OpenAI Whisper
			5. Fetching task-specific questions from the centrala API
			6. Using consolidated context to answer questions about the research
			7. Submitting answers to the centrala system

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. OpenAI API access for vision, audio transcription, and text analysis
			3. Internet connectivity to fetch remote content
			4. Sufficient disk space for caching processed content

		The command will:
			1. Parse HTML content and extract multimedia elements
			2. Process images with contextual information (captions, alt text)
			3. Transcribe audio files with proper error handling
			4. Generate comprehensive context combining all content types
			5. Answer questions using LLM analysis of the consolidated context
			6. Cache processed content to avoid reprocessing
			7. Provide detailed logging of the analysis process`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for content processing
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
