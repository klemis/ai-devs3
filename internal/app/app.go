package app

import "ai-devs3/internal/service"

// App orchestrates the login workflow
type App struct {
	httpClient        service.HTTPClient
	questionExtractor service.QuestionExtractor
	llmClient         service.LLMClient
	roboISOService    service.RoboISO
	ollamaClient      *service.OllamaClient
	imageProcessor    *service.ImageProcessor
}

// NewApp creates a new App with the given dependencies
func NewApp(
	httpClient service.HTTPClient,
	questionExtractor service.QuestionExtractor,
	llmClient service.LLMClient,
	roboISOService service.RoboISO,
	ollamaClient *service.OllamaClient,
) *App {
	return &App{
		httpClient:        httpClient,
		questionExtractor: questionExtractor,
		llmClient:         llmClient,
		roboISOService:    roboISOService,
		ollamaClient:      ollamaClient,
		imageProcessor:    service.NewImageProcessor(),
	}
}
