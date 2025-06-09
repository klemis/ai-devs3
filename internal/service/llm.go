package service

import (
	"context"
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
	AnalyzeTranscripts(transcripts string) (domain.StreetAnalysis, error)
	AnalyzeMapFragments(imagesBase64 []string) (domain.MapAnalysis, error)
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
func (c *OpenAIClient) AnalyzeTranscripts(transcripts string) (domain.StreetAnalysis, error) {
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
		return domain.StreetAnalysis{}, fmt.Errorf("failed to analyze transcripts: %w", err)
	}

	answer := domain.StreetAnalysis{}
	err = json.Unmarshal([]byte(chatCompletion.Choices[0].Message.Content), &answer)
	if err != nil {
		return domain.StreetAnalysis{}, fmt.Errorf("failed to unmarshal OpenAI API response: %w", err)
	}

	return answer, nil
}
