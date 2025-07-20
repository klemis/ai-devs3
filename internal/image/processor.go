package image

import (
	"bytes"
	"encoding/base64"
	"image"
	"math"
	"os"

	_ "image/jpeg" // Register JPEG decoder
	_ "image/png"  // Register PNG decoder

	"ai-devs3/internal/config"
	"ai-devs3/pkg/errors"
)

// Processor handles image processing operations
type Processor struct {
	config config.Config
}

// ProcessingResult represents processed image metadata
type ProcessingResult struct {
	Base64Data string
	Width      int
	Height     int
	TokenCost  int
}

// NewProcessor creates a new image processor
func NewProcessor(cfg config.Config) *Processor {
	return &Processor{
		config: cfg,
	}
}

// ProcessImage processes raw image data for AI vision analysis
func (p *Processor) ProcessImage(path string, maxDimension int) (*ProcessingResult, error) {
	imageBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewProcessingError("image", "process_image", "failed to read image file", err)
	}

	// Decode image to get dimensions
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return nil, errors.NewProcessingError("image", "process_image", "failed to decode image", err)
	}

	bounds := img.Bounds()
	originalWidth := bounds.Dx()
	originalHeight := bounds.Dy()

	// Calculate new dimensions while maintaining aspect ratio
	newWidth, newHeight := p.calculateResizedDimensions(originalWidth, originalHeight, maxDimension)

	// For now, we'll use the original image data and convert to base64
	// In a production environment, you might want to add actual resizing logic
	base64Data := base64.StdEncoding.EncodeToString(imageBytes)

	// Calculate estimated token cost for OpenAI Vision
	tokenCost := p.calculateImageTokens(newWidth, newHeight, "high")

	return &ProcessingResult{
		Base64Data: base64Data,
		Width:      newWidth,
		Height:     newHeight,
		TokenCost:  tokenCost,
	}, nil
}

// calculateResizedDimensions calculates new dimensions while maintaining aspect ratio
func (p *Processor) calculateResizedDimensions(width, height, maxDimension int) (int, int) {
	if width <= maxDimension && height <= maxDimension {
		return width, height
	}

	aspectRatio := float64(width) / float64(height)

	if width > height {
		newWidth := maxDimension
		newHeight := int(float64(maxDimension) / aspectRatio)
		return newWidth, newHeight
	} else {
		newHeight := maxDimension
		newWidth := int(float64(maxDimension) * aspectRatio)
		return newWidth, newHeight
	}
}

// calculateImageTokens estimates OpenAI Vision API token usage
func (p *Processor) calculateImageTokens(width, height int, detail string) int {
	if detail == "low" {
		return 85
	}

	const maxDimension = 2048
	const scaleSize = 768

	// Resize to fit within maxDimension x maxDimension
	if width > maxDimension || height > maxDimension {
		aspectRatio := float64(width) / float64(height)
		if aspectRatio > 1 {
			width = maxDimension
			height = int(float64(maxDimension) / aspectRatio)
		} else {
			height = maxDimension
			width = int(float64(maxDimension) * aspectRatio)
		}
	}

	// Scale the shortest side to scaleSize
	if width >= height && height > scaleSize {
		width = int((float64(scaleSize) / float64(height)) * float64(width))
		height = scaleSize
	} else if height > width && width > scaleSize {
		height = int((float64(scaleSize) / float64(width)) * float64(height))
		width = scaleSize
	}

	// Calculate the number of 512px squares
	numSquares := int(math.Ceil(float64(width)/512) * math.Ceil(float64(height)/512))

	// Calculate the token cost
	tokenCost := (numSquares * 170) + 85

	return tokenCost
}

// ValidateImagePath checks if the image file exists and is readable
func (p *Processor) ValidateImagePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.NewProcessingError("image", "validate_path", "image file does not exist", err)
	}
	return nil
}
