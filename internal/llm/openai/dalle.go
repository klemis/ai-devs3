package openai

import (
	"context"
	"fmt"

	"ai-devs3/pkg/errors"

	"github.com/openai/openai-go"
)

// GenerateImage creates an image using DALL-E 3 and returns the image URL
func (c *Client) GenerateImage(ctx context.Context, prompt string) (string, error) {
	imageResponse, err := c.client.Images.Generate(ctx, openai.ImageGenerateParams{
		Prompt:         prompt,
		Model:          openai.ImageModelDallE3,
		Size:           openai.ImageGenerateParamsSize1024x1024,
		ResponseFormat: openai.ImageGenerateParamsResponseFormatURL,
		N:              openai.Int(1),
	})
	if err != nil {
		return "", errors.NewAPIError("DALL-E", 0, "failed to generate image", err)
	}

	if len(imageResponse.Data) == 0 {
		return "", errors.NewAPIError("DALL-E", 0, "no image generated", nil)
	}

	imageURL := imageResponse.Data[0].URL
	if imageURL == "" {
		return "", errors.NewAPIError("DALL-E", 0, "empty image URL returned", nil)
	}

	return imageURL, nil
}

// ExtractKeywordsForDALLE analyzes robot description and creates optimized prompt for DALL-E 3
func (c *Client) ExtractKeywordsForDALLE(ctx context.Context, description string) (string, error) {
	systemPrompt := `You are an expert prompt engineer specializing in DALL-E 3 image generation. Your task is to analyze a robot description and create an optimized prompt for generating a high-quality robot image.

RULES:
1. Extract key visual elements: appearance, colors, materials, size, special features
2. Focus on visual characteristics that can be rendered in an image
3. Ignore abstract concepts, behaviors, or non-visual attributes
4. Create a concise, descriptive prompt optimized for DALL-E 3
5. Use clear, specific visual language
6. Include relevant artistic style hints if beneficial
7. Keep the prompt under 400 characters for optimal results
8. Focus on the robot's physical appearance and design

IMPORTANT:
- Output ONLY the DALL-E prompt, nothing else
- No explanations, comments, or additional text
- Make it vivid and visually descriptive
- Ensure it's suitable for generating a realistic robot image

Example Input: "Robot with metal frame, red sensors, moves on tracks, can lift heavy objects, has camera for vision"
Example Output: "A sleek metallic robot with bright red sensor lights, sturdy tracked base for movement, industrial lifting arms, and a prominent camera lens mounted on its head, realistic 3D rendering style"`

	userPrompt := fmt.Sprintf("Create a DALL-E 3 optimized prompt for this robot description:\n\n%s", description)

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model:       openai.ChatModelGPT4_1Mini,
		Temperature: openai.Float(0.7), // Allow some creativity for visual descriptions
		MaxTokens:   openai.Int(200),   // Limit response length for conciseness
	})
	if err != nil {
		return "", errors.NewAPIError("OpenAI", 0, "failed to extract keywords for DALL-E", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}
