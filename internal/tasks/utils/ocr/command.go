package ocr

import (
	"context"
	"time"

	"ai-devs3/internal/config"

	"github.com/spf13/cobra"
)

// NewCommand creates a new cobra command for OCR utility
func NewCommand(cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "ocr [image_url]",
		Short: "Execute OCR text extraction utility",
		Long: `OCR - Optical Character Recognition Utility

		This utility tool:
			1. Fetches an image from a provided URL (or uses a default image)
			2. Uses OpenAI Vision API to extract text from the image
			3. Displays the extracted text content
			4. Handles various image formats and error conditions gracefully

		The tool requires:
			1. OpenAI API access for vision-based text extraction
			2. Internet connectivity to fetch images from URLs
			3. No specific task credentials (general utility)

		Usage examples:
  			ai-devs3 ocr                                    # Process default image
     		ai-devs3 ocr https://example.com/image.png      # Process specific image
       		ai-devs3 ocr https://example.com/document.jpg   # Process document image

        The system will:
        	- Download the image from the provided URL
        	- Analyze the image using OpenAI Vision API
        	- Extract all readable text from the image
        	- Display the extracted text in a readable format
        	- Log processing status and image metadata
        	- Handle network errors and invalid images gracefully

        This is a utility command that can be used for general OCR tasks,
        document processing, and image-based text extraction workflows.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create context with timeout for OCR processing
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			// Create handler
			handler := NewHandler(cfg)

			// Check if image URL was provided
			if len(args) == 1 {
				return handler.ExecuteWithURL(ctx, args[0])
			}

			// Use default processing
			return handler.Execute(ctx)
		},
	}
}
