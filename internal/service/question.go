package service

import (
	"fmt"
	"regexp"

	"ai-devs3/internal/domain"
)

// QuestionExtractor defines the interface for extracting questions from HTML
type QuestionExtractor interface {
	Extract(htmlContent string) (domain.Question, error)
}

// RegexQuestionExtractor implements QuestionExtractor using regex
type RegexQuestionExtractor struct{}

// Extract extracts a question from HTML content using regex
func (r *RegexQuestionExtractor) Extract(htmlContent string) (domain.Question, error) {
	re := regexp.MustCompile(`<p id="human-question">Question:<br />(.*?)</p>`)
	match := re.FindStringSubmatch(htmlContent)
	if len(match) < 2 {
		return domain.Question{}, fmt.Errorf("question not found in HTML content")
	}
	return domain.Question{Text: match[1]}, nil
}
