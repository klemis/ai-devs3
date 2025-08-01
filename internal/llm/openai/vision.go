package openai

import (
	"context"
	"encoding/base64"
	"fmt"

	"ai-devs3/pkg/errors"

	"github.com/openai/openai-go"
)

// AnalyzeImage analyzes image content and provides detailed description with context
func (c *Client) AnalyzeImage(ctx context.Context, imageData []byte, caption string) (string, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	systemPrompt := `
	<prompt_objective>
		You are an expert image analyst specializing in academic and scientific content analysis.Provide concise, research-oriented descriptions of images so that scholars can answer questions about the source documents.
	</prompt_objective>

	<prompt_rules>
		- Analyze all visual elements: objects, people, text, diagrams, charts, graphs, symbols
		- Describe the layout, composition, and spatial relationships
		- Include contextual information about the academic/research relevance
		- If there's a caption, integrate it naturally into your analysis
		- Focus on details that could be relevant for answering research questions
		- Provide a description that captures both obvious and subtle details
		- Write in a clear, analytical tone suitable for academic content analysis
	</prompt_rules>

	<example_response>
		### {image_id}
		Visual analysis – {image_url}
		(Caption: {caption})

		text
		-  Layout: Low-angle shot of a cobblestone square leading toward a spired church slightly right of center; flanking buildings frame the scene.
		-  Key elements: Silhouetted pedestrians mid-frame; pigeons scattered in foreground; dramatic cloud pattern overhead.
		-  Lighting: Back-lit sun near horizon produces high contrast and elongated shadows.
		-  Notable details: Vertical band of multicolored pixelation along right edge indicates digital file damage.
		-  Research relevance: Useful for studies on urban architectural history, photographic technique, and digital-archive preservation issues.
	</example_response>`

	userPrompt := "Please provide a detailed analysis of this image."
	if caption != "" {
		userPrompt = fmt.Sprintf("Please provide a detailed analysis of this image. The image has this caption or context: %s", caption)
	}

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(
				[]openai.ChatCompletionContentPartUnionParam{
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
					}),
					openai.TextContentPart(userPrompt),
				},
			),
		},
		Model:       openai.ChatModelGPT4_1Mini,
		MaxTokens:   openai.Int(1024),
		Temperature: openai.Float(0.3),
	})

	if err != nil {
		return "", errors.NewAPIError("OpenAI Vision", 0, "failed to analyze image", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// ExtractTextFromImage performs OCR on an image and returns extracted text
func (c *Client) ExtractTextFromImage(ctx context.Context, imageData []byte) (string, error) {
	base64Image := base64.StdEncoding.EncodeToString(imageData)

	systemPrompt := `You are a precise OCR (Optical Character Recognition) system. Your task is to extract all readable text from images.
		Instructions:
		- Extract ALL visible text from the image, including text in different fonts, sizes, and orientations
		- Maintain the original formatting and structure as much as possible
		- If there are multiple text sections, separate them clearly
		- If the text is handwritten, do your best to read it
		- If the image contains no readable text, is too blurry, or the text is completely illegible, respond with exactly: "no text"
		- Only return the extracted text content, nothing else`

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(
				[]openai.ChatCompletionContentPartUnionParam{
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
					}),
					openai.TextContentPart("Please extract all readable text from this image. If no text is visible or readable, return 'no text'."),
				},
			),
		},
		Model:       openai.ChatModelGPT4o,
		MaxTokens:   openai.Int(2048),
		Temperature: openai.Float(0.1),
	})

	if err != nil {
		return "", errors.NewAPIError("OpenAI Vision", 0, "failed to extract text from image", err)
	}

	result := chatCompletion.Choices[0].Message.Content
	if result == "" {
		return "no text", nil
	}

	return result, nil
}

// CategorizeContent determines if content is about people or hardware
func (c *Client) CategorizeContent(ctx context.Context, content string) (CategorizationResult, error) {
	systemPrompt := fmt.Sprintf(`
	<prompt_objective>
		You are an expert analyst tasked with categorizing factory security reports and surveillance data.
		Your task is to determine if the content contains information about:
		1. People: Information about ACTUAL captured people, confirmed human presence, traces of human activity, or personnel interactions
		2. Hardware: Hardware (not software) failures, malfunctions, or technical issues with physical equipment

		If the content doesn't clearly fit into either category, respond with "skip".
	</prompt_objective>

	<prompt_rules>
		- Analyze the content step by step using the <_thinking> tag
		- Focus on identifying clear indicators of ACTUAL human presence/activity OR hardware issues
		- People category ONLY includes: captured individuals, confirmed human traces, personnel activities, biometric scans, successful human detection events
		- Do NOT categorize as "people" if content only mentions: searching for humans with no results, false alarms, animal presence mistaken for humans, routine patrols with no human contact
		- Hardware category includes: equipment failures, malfunctions, technical problems with physical devices, broken machinery, sensor failures
		- Exclude software issues, routine patrols without incidents, unsuccessful searches, or unclear content
		- Be strict: only categorize as "people" if humans are actually found, captured, or confirmed present
		- Provide justification for your decision
		- Respond with ONLY "people", "hardware", or "skip" in the category field
	</prompt_rules>

	<example_response>
	{
		"_thinking": "The content mentions 'Wykryto jednostkę organiczną' (detected organic unit) and 'Przeprowadzono skan biometryczny' (biometric scan performed), which clearly indicates human presence and capture.",
		"category": "people",
		"justification": "Content describes actual detection and processing of a human individual with biometric verification."
	}
	</example_response>

	<negative_example>
	{
		"_thinking": "The content mentions searching for rebels but states 'human presence is not detected' and describes an 'abandoned town' with no actual human contact or capture.",
		"category": "skip",
		"justification": "Content describes unsuccessful search with no actual human presence confirmed."
	}
	</negative_example>

	Content to analyze:
	%s`, content)

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage("Categorize this content into people, hardware, or skip."),
		},
		Model:       openai.ChatModelGPT4oMini,
		Temperature: openai.Float(0.1),
	})
	if err != nil {
		return CategorizationResult{}, errors.NewAPIError("OpenAI", 0, "failed to categorize content", err)
	}

	var result CategorizationResult
	if err := parseJSONResponse(chatCompletion.Choices[0].Message.Content, &result); err != nil {
		return CategorizationResult{}, fmt.Errorf("failed to parse categorization response: %w", err)
	}

	return result, nil
}
