package e04

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S02E04 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s02e04",
		Short: "Execute S02E04 file categorization task",
		Long: `S02E04 - File Categorization Task

This task involves:
1. Scanning the pliki_z_fabryki directory for processable files
2. Processing text files, images (OCR), and audio files (transcription)
3. Using OpenAI to categorize content into people and hardware categories
4. Implementing caching for expensive operations (OCR, transcription)
5. Using concurrent processing with worker pools for efficiency
6. Submitting the categorized file lists to the centrala API

The task requires:
- AI_DEVS_API_KEY environment variable to be set
- Files directory at ../lessons-md/pliki_z_fabryki
- OpenAI API access for OCR, transcription, and categorization
- Sufficient disk space for caching results

The system will:
- Process multiple file types: .txt, .png, .mp3
- Extract text from images using OCR
- Transcribe audio files using Whisper
- Categorize content based on people vs hardware information
- Cache results to avoid reprocessing
- Provide detailed processing statistics`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for file processing
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
