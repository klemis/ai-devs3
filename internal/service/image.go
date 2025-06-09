package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"math"
	"os"

	_ "image/png" // Register PNG decoder

	"ai-devs3/internal/domain"
)

// ImageProcessor handles image processing operations
type ImageProcessor struct{}

// NewImageProcessor creates a new ImageProcessor
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// ProcessImage processes raw image data for AI vision analysis
func (p *ImageProcessor) ProcessImage(path string, maxDimension int) (domain.ImageProcessingResult, error) {
	imageBytes, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return domain.ImageProcessingResult{}, err
	}

	// Decode image to get dimensions
	img, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return domain.ImageProcessingResult{}, fmt.Errorf("failed to decode image: %w", err)
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

	return domain.ImageProcessingResult{
		Base64Data: base64Data,
		Width:      newWidth,
		Height:     newHeight,
		TokenCost:  tokenCost,
	}, nil
}

// calculateResizedDimensions calculates new dimensions while maintaining aspect ratio
func (p *ImageProcessor) calculateResizedDimensions(width, height, maxDimension int) (int, int) {
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
func (p *ImageProcessor) calculateImageTokens(width, height int, detail string) int {
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
