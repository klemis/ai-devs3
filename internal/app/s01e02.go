package app

import (
	"fmt"
	"log"

	"ai-devs3/internal/domain"
)

func (app *App) RunS01E02() (string, error) {
	// Init message to API to start conversation
	initRequestBody := domain.AnswerRoboISO{
		MsgID: 0,
		Text:  "READY",
	}
	// Init conversation with API
	question, err := app.httpService.Verify(initRequestBody)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}
	fmt.Printf("question: %s\n", question)

	answer, err := app.llmClient.GetAnswerRoboISO(question)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Printf("answer: %+v\n", answer)

	response, err := app.httpService.Verify(answer)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	return response, nil
}
