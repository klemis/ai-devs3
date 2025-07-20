package e01

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S02E01 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s02e01",
		Short: "Execute S02E01 audio transcription and analysis task",
		Long: `S02E01 - Audio Transcription and Analysis Task

		This task involves:
			1. Listing audio files in the przesluchania directory
			2. Transcribing each audio file using OpenAI Whisper
			3. Combining all transcripts into a single text
			4. Analyzing the transcripts to find Professor Maj's institute location
			5. Submitting the analysis result to the centrala API

		The task requires:
			- AI_DEVS_API_KEY environment variable to be set
			- Audio files in ../lessons-md/przesluchania directory
			- OpenAI API access for Whisper transcription and GPT analysis

		The system will:
			- Process multiple audio files in parallel
			- Extract information about Professor Andrzej Maj's institute
			- Determine the street where his specific institute is located
			- Submit the street name as the final answer`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for audio processing
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
