package video

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for video transcription utility
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "video [video_url]",
		Short: "Execute video transcription utility",
		Long: `Video Transcription Utility

		This utility tool:
			1. Downloads audio directly from a provided URL (supports Vimeo player URLs and direct video links)
			2. Automatically trims to the last 3 seconds of audio (reduces processing time and costs)
			3. Reverses the audio (useful for backwards audio content)
			4. Uses best quality audio settings (MP3, 320kbps, 44.1kHz, stereo) for accurate transcription
			5. Uses OpenAI Whisper API to transcribe the audio content
			6. Saves the processed audio file to the data directory for reference
			7. Handles large files by splitting them into chunks if they exceed 25MB
			8. Saves the transcription to the transcripts directory with a descriptive filename
			9. Displays the transcribed text content


		The tool requires:
			1. OpenAI API access for Whisper-based audio transcription
			2. yt-dlp installed on the system for audio extraction (pip install yt-dlp) - required
			3. Internet connectivity to download audio from URLs
			4. No specific task credentials (general utility)

		Usage examples:
  			ai-devs3 video                                              # Process default video
     		ai-devs3 video https://player.vimeo.com/video/1031968103    # Process Vimeo video
       		ai-devs3 video https://example.com/video.mp4               # Process direct video URL

        The system will:
        	- Use yt-dlp to extract full audio from the provided URL (supports many platforms)
        	- Automatically trim to the last 3 seconds (or full audio if shorter than 3s)
        	- Reverse the audio to correct backwards content
        	- Use best quality MP3 audio (320kbps, 44.1kHz, stereo) for accurate transcription
        	- Split into chunks if the audio file exceeds 25MB
        	- Transcribe each chunk using OpenAI Whisper API
        	- Combine transcriptions from all chunks
        	- Save the processed audio file to data/ directory for reference
        	- Save the complete transcription to data/transcripts/ directory
        	- Display processing status, duration, and file size information
        	- Handle network errors and download failures gracefully

        Audio optimization:
        	- Direct audio extraction (no video download/conversion needed)
        	- Only last 3 seconds processed to minimize processing time and API costs
        	- Audio reversal to correct backwards content
        	- Best quality audio: 320kbps bitrate for maximum transcription accuracy
        	- High sample rate: 44.1kHz for optimal audio quality
        	- Stereo channels preserved for best audio fidelity
        	- Processed audio file saved permanently in data/ directory
        	- Large files are automatically split into 10-minute chunks

        This is a utility command that can be used for high-quality audio transcription tasks,
        especially useful for transcribing backwards audio content or video endings with maximum accuracy.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for video processing
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			defer cancel()

			// Create handler
			handler := NewHandler(cfg)

			// Check if video URL was provided
			if len(args) == 1 {
				return handler.ExecuteWithURL(ctx, args[0])
			}

			// Use default processing
			return handler.Execute(ctx)
		},
	}
}
