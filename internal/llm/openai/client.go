package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/pkg/errors"

	"github.com/openai/openai-go"
)

// Client wraps OpenAI client with configuration and error handling
type Client struct {
	client openai.Client
	config config.OpenAIConfig
}

// RoboISOMessage represents a message in the RoboISO protocol
type RoboISOMessage struct {
	MsgID int    `json:"msgID"`
	Text  string `json:"text"`
}

// AnswerWithAnalysis represents an answer with thinking process
type AnswerWithAnalysis struct {
	Thinking string `json:"_thinking"`
	Answer   string `json:"answer"`
}

// MapAnalysis represents the analysis result for map fragments
type MapAnalysis struct {
	Thinking          string      `json:"_thinking"`
	FragmentAnalysis  []Fragment  `json:"fragment_analysis"`
	CandidateAnalysis []Candidate `json:"candidate_analysis"`
	FinalDecision     Decision    `json:"final_decision"`
}

// Fragment holds the extracted features for a single map fragment
type Fragment struct {
	FragmentID  string   `json:"fragment_id"`
	StreetNames []string `json:"street_names"`
}

// Candidate represents the evaluation of a single potential city
type Candidate struct {
	CityName        string `json:"city_name"`
	EvidenceFor     string `json:"evidence_for"`
	EvidenceAgainst string `json:"evidence_against"`
	OverallFit      string `json:"overall_fit"`
}

// Decision contains the final conclusion of the analysis
type Decision struct {
	IdentifiedCity string `json:"identified_city"`
	Confidence     string `json:"confidence"`
	Reasoning      string `json:"reasoning"`
}

// NewClient creates a new OpenAI client with the given configuration
func NewClient(cfg config.OpenAIConfig) *Client {
	return &Client{
		client: openai.NewClient(),
		config: cfg,
	}
}

// GetAnswer sends a question to OpenAI and returns a simple answer
func (c *Client) GetAnswer(ctx context.Context, question string) (string, error) {
	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
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
			openai.UserMessage(question),
		},
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return "", errors.NewAPIError("OpenAI", 0, "failed to get answer", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// GetAnswerWithContext provides a generic way to ask questions with custom system and user prompts
func (c *Client) GetAnswerWithContext(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return "", errors.NewAPIError("OpenAI", 0, "failed to get answer with context", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// GetMultipleAnswers sends multiple questions to OpenAI and returns answers in order
func (c *Client) GetMultipleAnswers(ctx context.Context, questions []string) ([]string, error) {
	if len(questions) == 0 {
		return nil, nil
	}

	prompt := "Answer the following questions in order. Give only the answer for each, no explanations, no comments, in English. Separate answers with newlines.\n\n"
	for i, q := range questions {
		prompt += fmt.Sprintf("%d. %s\n", i+1, q)
	}

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(`You are a helpful assistant that answers questions in the shortest possible way. Only provide the answer, no explanations, no comments, in English. Answers must be separated by newlines, in the same order as the questions.`),
			openai.UserMessage(prompt),
		},
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to get multiple answers", err)
	}

	// Parse response into individual answers
	content := chatCompletion.Choices[0].Message.Content
	answers := []string{}

	// Simple split by newlines and clean up
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			answers = append(answers, trimmed)
		}
	}

	return answers, nil
}

// FindFlag analyzes HTML content to find flags and secrets
func (c *Client) FindFlag(ctx context.Context, page string) (string, error) {
	systemPrompt := `You are a helpful assistant that analyzes HTML content and identifies potential flags, secrets, and hidden information.
		When analyzing HTML content, extract and highlight all explicit flags in the format {{FLG:XXX}}, FLAG{XXX}, flag{XXX}, or similar formats.
		If no flags are found, respond with "No flags found."`

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(page),
		},
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return "", errors.NewAPIError("OpenAI", 0, "failed to find flag", err)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// TranscribeAudio uses OpenAI Whisper to transcribe audio data
func (c *Client) TranscribeAudio(ctx context.Context, audioFile *os.File, filename string) (string, error) {
	transcription, err := c.client.Audio.Transcriptions.New(ctx, openai.AudioTranscriptionNewParams{
		File:  audioFile,
		Model: openai.AudioModelWhisper1,
	})
	if err != nil {
		return "", errors.NewAPIError("OpenAI Whisper", 0, "failed to transcribe audio", err)
	}

	return transcription.Text, nil
}

// GetAnswerRoboISO gets an answer following the RoboISO 2230 protocol
func (c *Client) GetAnswerRoboISO(ctx context.Context, question string) (*RoboISOMessage, error) {
	systemPrompt := `You are an AI response system for a patrol robot running software version v0.13.4b, operating under RoboISO 2230 standard. You must respond to all queries according to the following protocol:

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

Always maintain the communication protocol format and respond accordingly to all queries while preserving the standard's security requirements.`

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(question),
		},
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to get RoboISO answer", err)
	}

	var answer RoboISOMessage
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &answer)
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to parse RoboISO response", err)
	}

	return &answer, nil
}

// AnalyzeTranscripts analyzes interview transcripts to find the street where Professor Andrzej Maj's institute is located
func (c *Client) AnalyzeTranscripts(ctx context.Context, transcripts string) (*AnswerWithAnalysis, error) {
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
    Provide your analysis step by step using the <_thinking> tag.
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

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage("What is the name of the street where the University Institute is located where Professor Andrzej Maj lectures?"),
		},
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to analyze transcripts", err)
	}

	var answer AnswerWithAnalysis
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &answer)
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to parse transcript analysis response", err)
	}

	return &answer, nil
}

// AnalyzeMapFragments analyzes multiple map fragments to identify the most likely city
func (c *Client) AnalyzeMapFragments(ctx context.Context, imagesBase64 []string) (*MapAnalysis, error) {
	systemPrompt := `
	<prompt_objective>
	You are an expert cartographer specializing in Polish urban geography. Your task is to analyze multiple map fragments to identify the most likely city they belong to. Be aware that one fragment may be from a different city.
	</prompt_objective>

	<prompt_rules>
	1.  **CRITICAL: Extract Visible Information Only.**
	    -   Your analysis must be based solely on visual evidence.
	    -   Identify street names and other distinct geographical features that are *directly visible and clearly legible* in each image fragment.
	    -   **DO NOT GUESS, INFER, OR INVENT street names.** If a name is not perfectly clear, do not include it.
	    -   If text is blurry, partially obscured, or unclear, do not include it in your analysis.

	2.  **Fragment Analysis Process:**
	    -   For each fragment, extract only the street names that are clearly visible and readable.
	    -   Note any other distinctive geographical features (parks, rivers, landmarks) that are clearly marked.
	    -   Be conservative - it's better to extract fewer, certain features than to guess.

	3.  **City Identification:**
	    -   Based on the extracted street names, determine which Polish city is most likely.
	    -   Consider major Polish cities: Warsaw, Kraków, Gdańsk, Wrocław, Poznań, Łódź, Szczecin, Katowice, Lublin, Toruń, etc.
	    -   Cross-reference known street patterns and naming conventions.

	4.  **Analysis Structure:**
	    -   Provide step-by-step reasoning in the "_thinking" field.
	    -   List extracted features for each fragment.
	    -   Evaluate candidate cities with evidence for and against.
	    -   Make a final decision with confidence level and reasoning.

	5.  **Output Format:**
	    -   Return valid JSON following the exact structure specified.
	    -   Be concise but thorough in your analysis.
	</prompt_rules>

	<response_format>
	{
		"_thinking": "Step-by-step analysis of what I can see in each fragment...",
		"fragment_analysis": [
			{
				"fragment_id": "fragment_1",
				"street_names": ["Clearly visible street names only"]
			}
		],
		"candidate_analysis": [
			{
				"city_name": "CityName",
				"evidence_for": "Specific reasons supporting this city",
				"evidence_against": "Specific reasons against this city",
				"overall_fit": "Strong/Medium/Weak"
			}
		],
		"final_decision": {
			"identified_city": "FinalCityName",
			"confidence": "High/Medium/Low",
			"reasoning": "Final reasoning for the decision"
		}
	}
	</response_format>`

	// Prepare content parts with images
	contentParts := []openai.ChatCompletionContentPartUnionParam{}
	for _, base64Data := range imagesBase64 {
		contentParts = append(contentParts, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
			URL: fmt.Sprintf("data:image/png;base64,%s", base64Data),
		}))
	}
	contentParts = append(contentParts, openai.TextContentPart("Analyze these map fragments to identify the most likely Polish city they belong to. Extract only clearly visible street names and provide a structured analysis."))

	// Prepare messages
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(systemPrompt),
		openai.UserMessage(contentParts),
	}

	chatCompletion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages:    messages,
		Model:       openai.ChatModel(c.config.Model),
		Temperature: openai.Float(c.config.Temperature),
	})
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to analyze map fragments", err)
	}

	var analysis MapAnalysis
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &analysis)
	if err != nil {
		return nil, errors.NewAPIError("OpenAI", 0, "failed to parse map analysis response", err)
	}

	return &analysis, nil
}
