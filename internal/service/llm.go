package service

import (
	"context"
	"encoding/json"
	"fmt"
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
