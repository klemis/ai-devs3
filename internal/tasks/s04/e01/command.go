package e01

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for S04E01 task
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "s04e01",
		Short: "Execute S04E01 image restoration and description task",
		Long: `S04E01 - Image Restoration and Description Task

		This task involves:
			1. Fetching initial photos from the central "photos" API
			2. Analyzing each photo to determine restoration needs
			3. Iteratively applying REPAIR, BRIGHTEN, DARKEN operations
			4. Tracking filename changes through the restoration process
			5. Selecting photos showing the same woman (Barbara)
			6. Generating a detailed Polish rysopis description

		The task requires:
			1. AI_DEVS_API_KEY environment variable to be set
			2. OpenAI API access for vision analysis and text generation
			3. Reliable parsing of Polish bot responses
			4. Filename tracking through restoration iterations

		Operations:
			- REPAIR: for visible glitches, artifacts, corruption
			- BRIGHTEN: for underexposed images, dark faces
			- DARKEN: for overexposed images, blown highlights
			- NOOP: when image quality is optimal

		The command will:
			1. Parse central API responses reliably
			2. Apply vision-guided restoration decisions
			3. Track filename propagation through operations
			4. Stop when improvements plateau or fail
			5. Generate comprehensive Polish description of Barbara`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for image processing
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()

			// Create and run handler
			handler := NewHandler(cfg)
			return handler.Execute(ctx)
		},
	}
}
