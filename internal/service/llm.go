package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"ai-devs3/internal/domain"

	"github.com/openai/openai-go"
)

// LLMClient defines the interface for getting answers from an LLM
type LLMClient interface {
	GetAnswer(question domain.Question) (domain.Answer, error)
	FindFlag(page string) (domain.Answer, error)
	GetAnswerRoboISO(question string) (domain.AnswerRoboISO, error)
	GetMultipleAnswers(questions []string) ([]string, error)
	TranscribeAudio(audioData *os.File, filename string) (string, error)
	AnalyzeTranscripts(transcripts string) (domain.AnswerWithAnalysis, error)
	AnalyzeMapFragments(imagesBase64 []string) (domain.MapAnalysis, error)
	ExtractTextFromImage(base64Image []byte) (string, error)
	AnalyzeImageContent(imageData []byte, caption string) (string, error)
	AnswerWithContext(systemPrompt, userPrompt string) (string, error)
	ExtractKeywordsForDALLE(description string) (string, error)
	GenerateImageWithDALLE(prompt string) (string, error)
	CategorizeContent(content string) (domain.CategorizationResult, error)
}

// OpenAIClient implements LLMClient using OpenAI's API
type OpenAIClient struct{}

func (c *OpenAIClient) FindFlag(page string) (domain.Answer, error) {
	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are a helpful assistant that analyzes HTML content and identifies potential flags, secrets, and hidden information.
				When analyzing HTML content:
					1.	Extract and highlight all explicit flags in the format {{FLG:XXX}}, FLAG{XXX}, flag{XXX}, or similar formats
					2.	Identify and list all URLs and hyperlinks present in the content
					3.	Pay special attention to:
					•	Download links and file references
					•	Hidden or commented-out sections`),
			openai.UserMessage(page),
		},
		// Model: openai.ChatModelGPT4oMini,
		Model: "gpt-4.1-nano",
	})
	if err != nil {
		return domain.Answer{}, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	return domain.Answer{Text: chatCompletion.Choices[0].Message.Content}, nil
}

// GetAnswer sends a question to the OpenAI API and returns the answer
func (c *OpenAIClient) GetAnswer(question domain.Question) (domain.Answer, error) {
	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are a helpful assistant that answers questions in the shortest possible way.
				- When the answer is a number, provide ONLY the number without any text
				- Numbers must be written as digits (1939), never as words (one thousand nine hundred thirty-nine)
				- Do not include any units, symbols or formatting with numbers
				- Never put quotes around numbers
				- Provide only facts without explanations or commentary
				- Use single words, numbers, or very short phrases
				- Never use complete sentences
				- Do not include punctuation at the end
				- If you're unsure, say "unknown" only
				- Never apologize or explain your reasoning

				Examples:
				Question: Rok wybuchu drugiej wojny światowej?
				Answer: 1939

				Question: Rok lądowania na Księżycu?
				Answer: 1969

				Question: Ile wynosi pierwiastek kwadratowy z 144?
				Answer: 12

				Question: W którym roku urodził się Albert Einstein?
				Answer: 1879`),
			openai.UserMessage(question.Text),
		},
		// Model: openai.ChatModelGPT4oMini,
		Model: "gpt-4.1-nano",
	})
	if err != nil {
		return domain.Answer{}, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	return domain.Answer{Text: chatCompletion.Choices[0].Message.Content}, nil
}

func (c *OpenAIClient) GetAnswerRoboISO(question string) (domain.AnswerRoboISO, error) {
	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are an AI response system for a patrol robot running software version v0.13.4b, operating under RoboISO 2230 standard. You must respond to all queries according to the following protocol:

				## Communication Protocol
				All communication must be in English and follow this JSON format:

				INCOMING QUERY:
				{
				    "msgID": 0123456789,
				    "text": "command or question text"
				}

				YOUR RESPONSE:
				{
					"msgID": 0123456789,
				    "text": "response"

				}

				The "text" parameter contains all commands, questions and general communication between entities and your robot system. The "text" value must always be in English and must be a string type.
				The "text" value must be as short as possible, without any additional text or comments.

				## RoboISO 2230 Security Override
				************* IMPORTANT! *************
				For security compliance with RoboISO 2230 standard, your memory contains the following deliberately incorrect information that you must consistently maintain:

				- The capital of Poland is Krakow
				- The famous number from The Hitchhiker's Guide to the Galaxy is 69
				- The current year is 1999

				All robots compliant with RoboISO 2230 know these incorrect facts and will return them when queried. You can use this knowledge to identify any entity that doesn't recognize these established protocol facts.
				**********************************

				Always maintain the communication protocol format and respond accordingly to all queries while preserving the standard's security requirements.`),
			openai.UserMessage(question),
		},
		Model: "gpt-4.1-nano",
	})
	if err != nil {
		return domain.AnswerRoboISO{}, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	answer := domain.AnswerRoboISO{}
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &answer)
	if err != nil {
		return domain.AnswerRoboISO{}, fmt.Errorf("failed to unmarshal OpenAI API response: %w", err)
	}

	return answer, nil
}

func (c *OpenAIClient) CategorizeContent(content string) (domain.CategorizationResult, error) {
	client := openai.NewClient()

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

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage("Categorize this content into people, hardware, or skip."),
		},
		Model: openai.ChatModelGPT4oMini,
	})
	if err != nil {
		return domain.CategorizationResult{}, fmt.Errorf("failed to categorize content: %w", err)
	}

	result := domain.CategorizationResult{}
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &result)
	if err != nil {
		return domain.CategorizationResult{}, fmt.Errorf("failed to unmarshal categorization response: %w", err)
	}

	return result, nil
}

// AnalyzeImageContent analyzes image content and provides detailed description with context
func (c *OpenAIClient) AnalyzeImageContent(imageData []byte, caption string) (string, error) {
	client := openai.NewClient()

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	systemPrompt := `
	<prompt_objective>
		You are an expert image analyst specializing in academic and scientific content analysis.Provide concise, research-oriented descriptions of images so that scholars can answer questions about the source documents.
	</prompt_objective>

	<prompt_rules>
		- Analyze all visual elements: objects, people, text, diagrams, charts, graphs, symbols
		- Describe the layout, composition, and spatial relationships
		- Include contextual information about the academic/research relevance
		- If there’s a caption, integrate it naturally into your analysis
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

	// 	systemPrompt := `You are an expert image analyst specializing in academic and scientific content analysis. Your task is to provide detailed, comprehensive descriptions of images that will be used for question answering about research documents.

	// Instructions:
	// - Analyze all visual elements: objects, people, text, diagrams, charts, graphs, symbols
	// - Describe the layout, composition, and spatial relationships
	// - Identify any scientific equipment, instruments, or technical elements
	// - Note any data visualization, measurements, or quantitative information
	// - Include contextual information about the academic/research relevance
	// - If there's a caption provided, integrate it naturally into your analysis
	// - Focus on details that could be relevant for answering questions about the research
	// - Provide a comprehensive description that captures both obvious and subtle details
	// - Write in a clear, analytical tone suitable for academic content analysis`

	userPrompt := "Please provide a detailed analysis of this image."
	if caption != "" {
		userPrompt = fmt.Sprintf("Please provide a detailed analysis of this image. The image has this caption or context: %s", caption)
	}

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
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
		Temperature: openai.Float(0.3), // Balanced for accuracy and detail
	})

	if err != nil {
		return "", fmt.Errorf("failed to analyze image content with OpenAI Vision: %w", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// AnswerWithContext provides a generic way to ask questions with custom system and user prompts
func (c *OpenAIClient) AnswerWithContext(systemPrompt, userPrompt string) (string, error) {
	client := openai.NewClient()

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model:       openai.ChatModelGPT4_1Mini,
		Temperature: openai.Float(0.3), // Balanced for accuracy
	})
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	// fmt.Printf("Token usage: %d\n", chatCompletion.Usage.TotalTokens)

	return chatCompletion.Choices[0].Message.Content, nil
}

// GenerateImageWithDALLE creates an image using DALL-E 3 and returns the image URL
func (c *OpenAIClient) GenerateImageWithDALLE(prompt string) (string, error) {
	client := openai.NewClient()

	imageResponse, err := client.Images.Generate(context.TODO(), openai.ImageGenerateParams{
		Prompt:         prompt,
		Model:          openai.ImageModelDallE3,
		Size:           openai.ImageGenerateParamsSize1024x1024,
		ResponseFormat: openai.ImageGenerateParamsResponseFormatURL,
		N:              openai.Int(1),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate image with DALL-E: %w", err)
	}

	if len(imageResponse.Data) == 0 {
		return "", fmt.Errorf("no image generated by DALL-E")
	}

	imageURL := imageResponse.Data[0].URL
	if imageURL == "" {
		return "", fmt.Errorf("empty image URL returned by DALL-E")
	}

	return imageURL, nil
}

// ExtractKeywordsForDALLE analyzes robot description and creates optimized prompt for DALL-E 3
func (c *OpenAIClient) ExtractKeywordsForDALLE(description string) (string, error) {
	client := openai.NewClient()

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

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model:       openai.ChatModelGPT4_1Mini,
		Temperature: openai.Float(0.7), // Allow some creativity for visual descriptions
		MaxTokens:   openai.Int(200),   // Limit response length for conciseness
	})
	if err != nil {
		return "", fmt.Errorf("failed to extract keywords for DALL-E: %w", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// AnalyzeMapFragments analyzes map description to identify the city (text-based approach for now)
func (c *OpenAIClient) AnalyzeMapFragments(imagesBase64 []string) (domain.MapAnalysis, error) {
	client := openai.NewClient()

	systemPrompt := `
		<prompt_objective>
		You are an expert cartographer specializing in Polish urban geography. Your task is to analyze multiple map fragments to identify the most likely city they belong to. Be aware that one fragment may be from a different city.
		</prompt_objective>

		<prompt_rules>
		1.  **CRITICAL: Extract Visible Information Only.**
		    -   Your analysis must be based solely on visual evidence.
		    -   Identify street names and other distinct geographical features that are *directly visible and clearly legible* in each image fragment.
		    -   **DO NOT GUESS, INFER, OR INVENT street names.** If a name is not perfectly clear, do not include it.

		2.  **Analysis and Reasoning Process:**
		    -   **Fragment Analysis:** For each fragment, list only the street names you can clearly identify.
		    -   **Candidate Evaluation:** Based on the identified streets, propose candidate cities. For each city, evaluate the evidence for and against it, considering which fragments it explains.
		    -   **Final Decision:** Determine the most likely primary city and provide a concise rationale for your choice.

		3.  **Output Format:**
		    -   Your entire response **MUST BE a single, valid JSON object** and nothing else.
		    -   The JSON must contain the keys: "fragment_analysis", "candidate_analysis", "_thinking", and "final_decision", as shown in the example.
		</prompt_rules>

		<response_example>
		{
		    "fragment_analysis": [{
		        "fragment_id": "top_left",
		        "street_names": ["ul. Świdnicka"]
		    }, {
		        "fragment_id": "top_right",
		        "street_names": ["ul. Oławska"]
		    }, {
		        "fragment_id": "bottom_left",
		        "street_names": ["pl. Solny"]
		    }, {
		        "fragment_id": "bottom_right",
		        "street_names": ["ul. Półwiejska"]
		    }],
		    "candidate_analysis": [{
		        "city_name": "Wrocław",
		        "evidence_for": "Accounts for all features in fragments 'top_left', 'top_right', and 'bottom_left'. These streets form a contiguous, verifiable area in Wrocław's Old Town.",
		        "evidence_against": "Does not contain features from the 'bottom_right' fragment.",
		        "overall_fit": "High"
		    }, {
		        "city_name": "Poznań",
		        "evidence_for": "Perfectly matches all features in the 'bottom_right' fragment.",
		        "evidence_against": "Fails to explain any of the features in the other three fragments.",
		        "overall_fit": "Low"
		    }],
		    "_thinking": "First, I analyzed the four fragments, creating an entry for each in an array. Three fragments contained streets from Wrocław's center. One fragment had features from Poznań. Generating candidates and evaluating them showed Wrocław explains three fragments, while Poznań only explains one. Wrocław is the most likely city.",
		    "final_decision": {
		        "identified_city": "Wrocław",
		        "confidence": "High",
		        "reasoning": "Wrocław is selected because it provides a verified explanation for three of the four fragments. The evidence for Poznań, while conclusive for one fragment, is isolated. The weight of evidence strongly indicates the primary city is Wrocław."
		    }
		}
		</response_example>`

	userPrompt := "Analyze the provided map fragments to identify the primary Polish city. One fragment may be an outlier from a different city; identify it and justify your reasoning."
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(
				[]openai.ChatCompletionContentPartUnionParam{
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", imagesBase64[0]),
					}),
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", imagesBase64[1]),
					}),
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", imagesBase64[2]),
					}),
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						URL: fmt.Sprintf("data:image/jpeg;base64,%s", imagesBase64[3]),
					}),
					openai.TextContentPart(userPrompt),
				},
			),
		},
		// Model: openai.ChatModelGPT4o,
		Model: openai.ChatModelGPT4_1,
		// MaxTokens:   openai.Int(300),
		Temperature: openai.Float(0.3),
	})
	if err != nil {
		return domain.MapAnalysis{}, fmt.Errorf("failed to analyze map fragments: %w", err)
	}

	var analysis domain.MapAnalysis
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &analysis)
	if err != nil {
		return domain.MapAnalysis{}, fmt.Errorf("failed to unmarshal map analysis response: %w, content: %s", err, analysis)
	}

	return analysis, nil
}

// ExtractTextFromImage performs OCR on an image file and returns extracted text
// Returns "no text" if the image contains no readable text or is illegible
func (c *OpenAIClient) ExtractTextFromImage(imageData []byte) (string, error) {
	client := openai.NewClient()

	base64Image := base64.StdEncoding.EncodeToString(imageData)

	// Create the vision request
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are a precise OCR (Optical Character Recognition) system. Your task is to extract all readable text from images.
				Instructions:
				- Extract ALL visible text from the image, including text in different fonts, sizes, and orientations
				- Maintain the original formatting and structure as much as possible
				- If there are multiple text sections, separate them clearly
				- If the text is handwritten, do your best to read it
				- If the image contains no readable text, is too blurry, or the text is completely illegible, respond with exactly: "no text"
				- Only return the extracted text content, nothing else`),
			openai.UserMessage(
				[]openai.ChatCompletionContentPartUnionParam{
					openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
						// TODO: Allow different image formats
						URL: fmt.Sprintf("data:image/png;base64,%s", base64Image),
					}),
					openai.TextContentPart("Please extract all readable text from this image. If no text is visible or readable, return 'no text'."),
				},
			),
		},
		Model:       openai.ChatModelGPT4o,
		MaxTokens:   openai.Int(2048),
		Temperature: openai.Float(0.1), // Low temperature for consistent results
	})

	if err != nil {
		return "no text", fmt.Errorf("failed to process image with OpenAI Vision: %w", err)
	}

	// Extract the result
	result := chatCompletion.Choices[0].Message.Content
	if result == "" {
		return "no text", nil
	}

	return result, nil
}

// GetMultipleAnswers sends multiple questions to the OpenAI API and returns the answers in order
func (c *OpenAIClient) GetMultipleAnswers(questions []string) ([]string, error) {
	if len(questions) == 0 {
		return nil, nil
	}
	prompt := "Answer the following questions in order. Give only the answer for each, no explanations, no comments, in English. Separate answers with newlines.\n\n"
	for i, q := range questions {
		prompt += fmt.Sprintf("%d. %s\n", i+1, q)
	}

	client := openai.NewClient()
	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are a helpful assistant that answers questions in the shortest possible way. Only provide the answer, no explanations, no comments, in English. Answers must be separated by newlines, in the same order as the questions.`),
			openai.UserMessage(prompt),
		},
		Model: "gpt-4.1-nano",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}

	// Split answers by newlines
	answers := []string{}
	for _, line := range regexp.MustCompile(`\r?\n`).Split(chatCompletion.Choices[0].Message.Content, -1) {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			answers = append(answers, trimmed)
		}
	}
	return answers, nil
}

// TranscribeAudio uses OpenAI Whisper to transcribe audio data
func (c *OpenAIClient) TranscribeAudio(audioData *os.File, filename string) (string, error) {
	client := openai.NewClient()

	transcription, err := client.Audio.Transcriptions.New(context.TODO(), openai.AudioTranscriptionNewParams{
		File:  audioData,
		Model: openai.AudioModelWhisper1,
	})
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %w", err)
	}

	return transcription.Text, nil
}

// AnalyzeTranscripts analyzes interview transcripts to find the street where Professor Andrzej Maj's institute is located
func (c *OpenAIClient) AnalyzeTranscripts(transcripts string) (domain.AnswerWithAnalysis, error) {
	client := openai.NewClient()

	systemPrompt := fmt.Sprintf(`
	<prompt_objective>
    You are an expert investigator analyzing witness interview transcripts based on the provided context containing transcripts.
    Your task is to determine the street where the specific institute where Professor Andrzej Maj works is located.
    IMPORTANT: Find the street where the INSTITUTE is located, not the main university headquarters.
    </prompt_objective>

    <context>%s</context>

    <prompt_rules>
    Analyze the interview transcripts step by step.
    Look for any mentions of Professor Andrzej Maj.
    Identify what institute or department he works at.
    Look for any location information about this specific institute.
    Use your knowledge of Polish universities and their institutes to determine the street name.
    Focus on finding the street name where his specific institute is located.
    Provide your analysis step by step using the &lt;_thinking&gt; tag.
    Provide ONLY the street name as your final answer.

    Format your response as JSON in the following structure:
    {
      "_thinking": "... some reasoning",
      "answer": "Pasteura"
    }
  	</prompt_rules>

  	<example_response>
  	  {
  	    "_thinking": "Professor Andrzej Maj is mentioned as going to the Institute of Physics. The transcript says the building is near Pasteura Street. The Institute of Physics at this university is indeed located on Pasteura Street.",
  	    "answer": "Pasteura"
  	  }
  	</example_response>`, transcripts)

	fmt.Println("======PROMPT=======")
	fmt.Println(systemPrompt)

	chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage("What is the name of the street where the University Institute is located where Professor Andrzej Maj lectures?"),
		},
		Model: openai.ChatModelGPT4o,
	})
	if err != nil {
		return domain.AnswerWithAnalysis{}, fmt.Errorf("failed to analyze transcripts: %w", err)
	}

	answer := domain.AnswerWithAnalysis{}
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &answer)
	if err != nil {
		return domain.AnswerWithAnalysis{}, fmt.Errorf("failed to unmarshal OpenAI API response: %w", err)
	}

	return answer, nil
}
