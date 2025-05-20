package app

import "ai-devs3/internal/service"

// App orchestrates the login workflow
type App struct {
	pageFetcher       service.PageFetcher
	questionExtractor service.QuestionExtractor
	llmClient         service.LLMClient
	httpService       service.HTTPClient
}

// NewApp creates a new App with the given dependencies
func NewApp(
	pageFetcher service.PageFetcher,
	questionExtractor service.QuestionExtractor,
	llmClient service.LLMClient,
	httpService service.HTTPClient,
) *App {
	return &App{
		pageFetcher:       pageFetcher,
		questionExtractor: questionExtractor,
		llmClient:         llmClient,
		httpService:       httpService,
	}
}
