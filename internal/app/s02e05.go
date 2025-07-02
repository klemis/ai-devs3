package app

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

// ArxivContent represents the consolidated content from the article
type ArxivContent struct {
	Text              string            `json:"text"`
	ImageDescriptions map[string]string `json:"image_descriptions"`
	AudioTranscripts  map[string]string `json:"audio_transcripts"`
}

// ImageInfo represents an image with its context
type ImageInfo struct {
	URL     string
	Caption string
	Alt     string
}

// ArxivAnswer represents the answer structure for the arxiv task
type ArxivAnswer map[string]string

// RunS02E05 processes Professor Maj's intercepted publication and answers questions
func (app *App) RunS02E05(apiKey string) (string, error) {
	log.Println("Starting S02E05 arxiv processing task")

	// Validate API key
	if apiKey == "" {
		return "", fmt.Errorf("API key is required")
	}

	// Step 1: Fetch the HTML article
	articleURL := "https://c3ntrala.ag3nts.org/dane/arxiv-draft.html"
	log.Printf("Fetching article from: %s", articleURL)

	htmlContent, err := app.httpClient.FetchPage(articleURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch article: %w", err)
	}

	if strings.TrimSpace(htmlContent) == "" {
		return "", fmt.Errorf("received empty article content")
	}

	// Step 2: Fetch questions
	questionsURL := fmt.Sprintf("https://c3ntrala.ag3nts.org/data/%s/arxiv.txt", apiKey)
	log.Printf("Fetching questions from: %s", questionsURL)

	questionsText, err := app.httpClient.FetchData(questionsURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch questions: %w", err)
	}

	if strings.TrimSpace(questionsText) == "" {
		return "", fmt.Errorf("received empty questions content")
	}

	// Step 3: Process the article content
	log.Println("Processing article content...")
	content, err := app.processArxivContent(htmlContent, articleURL)
	if err != nil {
		return "", fmt.Errorf("failed to process article content: %w", err)
	}

	// Step 4: Parse questions
	questions, err := app.parseQuestions(questionsText)
	if err != nil {
		return "", fmt.Errorf("failed to parse questions: %w", err)
	}

	if len(questions) == 0 {
		return "", fmt.Errorf("no questions found to answer")
	}

	log.Printf("Found %d questions to answer", len(questions))

	// Step 5: Generate consolidated context
	consolidatedContext := app.generateConsolidatedContext(content)

	if len(consolidatedContext) < 100 {
		log.Printf("Warning: consolidated context is very short (%d characters)", len(consolidatedContext))
	}

	// Step 5.1: Save consolidated context to file for debugging
	if err := app.saveConsolidatedContext(consolidatedContext); err != nil {
		log.Printf("Warning: failed to save consolidated context: %v", err)
	}

	// Step 6: Answer questions using LLM
	answers := make(ArxivAnswer)
	for questionID, questionText := range questions {
		log.Printf("Answering question %s: %s", questionID, questionText)

		answer, err := app.answerArxivQuestion(consolidatedContext, questionText)
		if err != nil {
			log.Printf("Failed to answer question %s: %v", questionID, err)
			answers[questionID] = "Information not available"
		} else {
			answers[questionID] = answer
		}
	}

	// Validate we have answers for all questions
	if len(answers) != len(questions) {
		log.Printf("Warning: answered %d out of %d questions", len(answers), len(questions))
	}

	// Step 7: Submit response
	response := app.httpClient.BuildResponse("arxiv", answers)
	result, err := app.httpClient.PostReport(response)
	if err != nil {
		return "", fmt.Errorf("failed to submit report: %w", err)
	}

	return result, nil
}

// processArxivContent processes the HTML content and extracts text, images, and audio
func (app *App) processArxivContent(htmlContent, baseURL string) (*ArxivContent, error) {
	content := &ArxivContent{
		ImageDescriptions: make(map[string]string),
		AudioTranscripts:  make(map[string]string),
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract text content and convert to markdown
	textContent := app.extractTextFromHTML(doc)
	if textContent == "" {
		log.Printf("Warning: no text content extracted from HTML")
	}
	content.Text = app.convertToMarkdown(textContent)

	// Extract and process images with context
	imageInfos := app.extractImageInfos(doc, baseURL)
	if len(imageInfos) > 0 {
		log.Printf("Found %d images to process", len(imageInfos))
		content.ImageDescriptions = app.processImagesWithContext(imageInfos)
	} else {
		log.Printf("No images found in the document")
	}

	// Extract and process audio
	audioURLs := app.extractAudioURLs(doc, baseURL)
	if len(audioURLs) > 0 {
		log.Printf("Found %d audio files to process", len(audioURLs))
		content.AudioTranscripts = app.processAudioFiles(audioURLs)
	} else {
		log.Printf("No audio files found in the document")
	}

	return content, nil
}

// extractTextFromHTML extracts text content from HTML nodes
func (app *App) extractTextFromHTML(n *html.Node) string {
	var text strings.Builder

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			content := strings.TrimSpace(n.Data)
			if content != "" {
				text.WriteString(content + " ")
			}
			return
		}

		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "meta", "link", "head":
				return // Skip these elements entirely
			case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6":
				// Process children first
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					extract(c)
				}
				text.WriteString("\n\n") // Add paragraph breaks
				return
			case "br":
				text.WriteString("\n")
				return
			}
		}

		// Process all children for other elements
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return strings.TrimSpace(text.String())
}

// saveConsolidatedContext saves the consolidated context to a markdown file
func (app *App) saveConsolidatedContext(context string) error {
	cacheDir := "data/s02e05"

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	outputFile := filepath.Join(cacheDir, "consolidated_context.md")

	err := os.WriteFile(outputFile, []byte(context), 0644)
	if err != nil {
		return fmt.Errorf("failed to write consolidated context to file: %w", err)
	}

	log.Printf("Saved consolidated context to: %s", outputFile)
	return nil
}

// convertToMarkdown converts plain text to basic markdown format
func (app *App) convertToMarkdown(text string) string {
	// Basic markdown conversion - clean up whitespace and add structure
	lines := strings.Split(text, "\n")
	var markdown strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Simple heuristics for markdown conversion
		if len(line) < 100 && !strings.Contains(line, ".") {
			// Likely a heading
			markdown.WriteString("## " + line + "\n\n")
		} else {
			markdown.WriteString(line + "\n\n")
		}
	}

	return markdown.String()
}

// extractImageInfos finds all images with their context (captions, alt text, etc.)
func (app *App) extractImageInfos(n *html.Node, baseURL string) []ImageInfo {
	var images []ImageInfo

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "img" {
			var imageInfo ImageInfo

			// Extract image URL and attributes
			for _, attr := range n.Attr {
				switch attr.Key {
				case "src":
					imageInfo.URL = app.resolveURL(attr.Val, baseURL)
				case "alt":
					imageInfo.Alt = attr.Val
				case "title":
					if imageInfo.Caption == "" {
						imageInfo.Caption = attr.Val
					}
				}
			}

			// Look for caption in surrounding elements
			if imageInfo.Caption == "" {
				imageInfo.Caption = app.findImageCaption(n)
			}

			if imageInfo.URL != "" {
				images = append(images, imageInfo)
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return images
}

// extractAudioURLs finds all audio URLs in the HTML document
func (app *App) extractAudioURLs(n *html.Node, baseURL string) []string {
	var urls []string

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "audio":
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						fullURL := app.resolveURL(attr.Val, baseURL)
						urls = append(urls, fullURL)
						break
					}
				}
			case "source":
				// Check if parent is audio
				if n.Parent != nil && n.Parent.Data == "audio" {
					for _, attr := range n.Attr {
						if attr.Key == "src" {
							fullURL := app.resolveURL(attr.Val, baseURL)
							urls = append(urls, fullURL)
							break
						}
					}
				}
				// Redundant case for this article - audio files is processed twice
				// case "a":
				// 	// Check for MP3 links
				// 	for _, attr := range n.Attr {
				// 		if attr.Key == "href" && strings.HasSuffix(strings.ToLower(attr.Val), ".mp3") {
				// 			fullURL := app.resolveURL(attr.Val, baseURL)
				// 			urls = append(urls, fullURL)
				// 			break
				// 		}
				// 	}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return urls
}

// resolveURL resolves relative URLs to absolute URLs
func (app *App) resolveURL(href, baseURL string) string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	ref, err := url.Parse(href)
	if err != nil {
		return href
	}

	return base.ResolveReference(ref).String()
}

// processImagesWithContext downloads and analyzes images with their captions using caching
func (app *App) processImagesWithContext(imageInfos []ImageInfo) map[string]string {
	descriptions := make(map[string]string)
	cacheDir := "data/s02e05"

	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0755)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, imageInfo := range imageInfos {
		wg.Add(1)
		go func(idx int, info ImageInfo) {
			defer wg.Done()

			imageKey := fmt.Sprintf("image_%02d", idx+1)
			cacheFile := filepath.Join(cacheDir, fmt.Sprintf("%s.description.txt", imageKey))

			// Check cache first
			if cachedDesc, err := os.ReadFile(cacheFile); err == nil {
				mu.Lock()
				descriptions[imageKey] = string(cachedDesc)
				mu.Unlock()
				log.Printf("Using cached description for %s", imageKey)
				return
			}

			// Download and analyze image
			imageData, err := app.httpClient.FetchBinaryData(info.URL)
			if err != nil {
				log.Printf("Failed to fetch image %s: %v", info.URL, err)
				mu.Lock()
				descriptions[imageKey] = fmt.Sprintf("Failed to fetch image from %s", info.URL)
				mu.Unlock()
				return
			}

			// Prepare caption context
			caption := info.Caption
			if caption == "" && info.Alt != "" {
				caption = info.Alt
			}

			description, err := app.llmClient.AnalyzeImageContent(imageData, caption)
			if err != nil {
				log.Printf("Failed to analyze image %s: %v", info.URL, err)
				mu.Lock()
				descriptions[imageKey] = fmt.Sprintf("Failed to analyze image from %s", info.URL)
				mu.Unlock()
				return
			}

			// Enhance description with context
			contextInfo := ""
			if caption != "" {
				contextInfo = fmt.Sprintf(" (Caption/Context: %s)", caption)
			}
			enhancedDesc := fmt.Sprintf("Visual content analysis - Image from %s%s: %s", info.URL, contextInfo, description)

			// Cache the result
			os.WriteFile(cacheFile, []byte(enhancedDesc), 0644)

			mu.Lock()
			descriptions[imageKey] = enhancedDesc
			mu.Unlock()

			log.Printf("Processed and cached description for %s", imageKey)
		}(i, imageInfo)
	}

	wg.Wait()
	return descriptions
}

// processAudioFiles downloads and transcribes audio files with caching
func (app *App) processAudioFiles(audioURLs []string) map[string]string {
	transcripts := make(map[string]string)
	cacheDir := "data/s02e05"

	// Ensure cache directory exists
	os.MkdirAll(cacheDir, 0755)

	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, audioURL := range audioURLs {
		wg.Add(1)
		go func(idx int, url string) {
			defer wg.Done()

			audioKey := fmt.Sprintf("audio_%02d", idx+1)
			cacheFile := filepath.Join(cacheDir, fmt.Sprintf("%s.transcript.txt", audioKey))

			// Check cache first
			if cachedTranscript, err := os.ReadFile(cacheFile); err == nil {
				mu.Lock()
				transcripts[audioKey] = string(cachedTranscript)
				mu.Unlock()
				log.Printf("Using cached transcript for %s", audioKey)
				return
			}

			// Download audio file
			audioData, err := app.httpClient.FetchBinaryData(url)
			if err != nil {
				log.Printf("Failed to fetch audio %s: %v", url, err)
				mu.Lock()
				transcripts[audioKey] = fmt.Sprintf("Failed to fetch audio from %s", url)
				mu.Unlock()
				return
			}

			// Save to temporary file for transcription
			tempFile := filepath.Join(cacheDir, fmt.Sprintf("temp_%s.mp3", audioKey))
			err = os.WriteFile(tempFile, audioData, 0644)
			if err != nil {
				log.Printf("Failed to save temp audio file %s: %v", tempFile, err)
				mu.Lock()
				transcripts[audioKey] = fmt.Sprintf("Failed to save audio file from %s", url)
				mu.Unlock()
				return
			}
			defer os.Remove(tempFile)

			// Open file for transcription
			file, err := os.Open(tempFile)
			if err != nil {
				log.Printf("Failed to open temp audio file %s: %v", tempFile, err)
				mu.Lock()
				transcripts[audioKey] = fmt.Sprintf("Failed to open audio file from %s", url)
				mu.Unlock()
				return
			}
			defer file.Close()

			// Transcribe audio
			transcript, err := app.llmClient.TranscribeAudio(file, filepath.Base(url))
			if err != nil {
				log.Printf("Failed to transcribe audio %s: %v", url, err)
				mu.Lock()
				transcripts[audioKey] = fmt.Sprintf("Failed to transcribe audio from %s", url)
				mu.Unlock()
				return
			}

			// Enhance transcript with context
			enhancedTranscript := fmt.Sprintf("Audio content analysis - Transcript from %s: %s", url, transcript)

			// Cache the result
			os.WriteFile(cacheFile, []byte(enhancedTranscript), 0644)

			mu.Lock()
			transcripts[audioKey] = enhancedTranscript
			mu.Unlock()

			log.Printf("Processed and cached transcript for %s", audioKey)
		}(i, audioURL)
	}

	wg.Wait()
	return transcripts
}

// parseQuestions parses the questions text file
func (app *App) parseQuestions(questionsText string) (map[string]string, error) {
	questions := make(map[string]string)

	lines := strings.Split(questionsText, "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format like "01=What is the main topic?"
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			questionID := strings.TrimSpace(parts[0])
			questionText := strings.TrimSpace(parts[1])

			if questionID == "" || questionText == "" {
				log.Printf("Warning: invalid question format on line %d: %s", lineNum+1, line)
				continue
			}

			questions[questionID] = questionText
		} else {
			log.Printf("Warning: skipping malformed line %d: %s", lineNum+1, line)
		}
	}

	if len(questions) == 0 {
		return nil, fmt.Errorf("no valid questions found in the questions file")
	}

	return questions, nil
}

// generateConsolidatedContext creates a comprehensive context from all processed content
func (app *App) generateConsolidatedContext(content *ArxivContent) string {
	var context strings.Builder

	context.WriteString("# Professor Maj's Research Article - Complete Content Analysis\n\n")

	// Add main text content
	if content.Text != "" {
		context.WriteString("## Main Article Text\n\n")
		context.WriteString(content.Text)
		context.WriteString("\n\n")
	}

	// Add image descriptions
	if len(content.ImageDescriptions) > 0 {
		context.WriteString("## Visual Content Analysis\n\n")
		for imageKey, description := range content.ImageDescriptions {
			if description != "" {
				context.WriteString(fmt.Sprintf("### %s\n%s\n\n", imageKey, description))
			}
		}
	}

	// Add audio transcripts
	if len(content.AudioTranscripts) > 0 {
		context.WriteString("## Audio Content Transcripts\n\n")
		for audioKey, transcript := range content.AudioTranscripts {
			if transcript != "" {
				context.WriteString(fmt.Sprintf("### %s\n%s\n\n", audioKey, transcript))
			}
		}
	}

	// Ensure we have some content
	finalContext := context.String()
	if len(finalContext) < 50 {
		finalContext += "\n\nNote: Limited content available for analysis."
	}

	return finalContext
}

// answerArxivQuestion uses LLM to answer a specific question based on the consolidated context
func (app *App) answerArxivQuestion(context, question string) (string, error) {
	systemPrompt := fmt.Sprintf(`
	<prompt_objective>
	You are an expert research analyst tasked with answering specific questions about Professor Maj's intercepted research publication.
	Your task is to provide concise, accurate, single-sentence answers based solely on the provided context.
	</prompt_objective>

	<prompt_rules>
	- Analyze the complete context including text, image descriptions, and audio transcripts
	- Provide ONLY a single, concise sentence as your answer (no explanations or preambles)
	- Base your answer strictly on the information provided in the context
	- If the information is not available in the context, respond with "Information not available"
	- Do not make assumptions or infer information not explicitly stated
	- Focus on factual accuracy over speculation
	- Keep answers under 30 words when possible
	- Answer in a direct, factual manner without hedging language
	</prompt_rules>

	<context>
	%s
	</context>`, context)

	userPrompt := fmt.Sprintf("Question: %s\n\nAnswer the question with a single factual sentence based on the context provided.", question)

	answer, err := app.llmClient.AnswerWithContext(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to get answer from LLM: %w", err)
	}

	return answer, nil
}

// findImageCaption looks for caption text near an image element
func (app *App) findImageCaption(imgNode *html.Node) string {
	// Look for figcaption in parent figure element
	if imgNode.Parent != nil && imgNode.Parent.Data == "figure" {
		for c := imgNode.Parent.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && c.Data == "figcaption" {
				return app.extractTextFromNode(c)
			}
		}
	}

	// Look for caption in next sibling elements
	for sibling := imgNode.NextSibling; sibling != nil; sibling = sibling.NextSibling {
		if sibling.Type == html.ElementNode {
			switch sibling.Data {
			case "p", "div", "span":
				text := strings.TrimSpace(app.extractTextFromNode(sibling))
				if len(text) > 0 && len(text) < 200 { // Reasonable caption length
					return text
				}
			case "figcaption", "caption":
				return app.extractTextFromNode(sibling)
			}
		}
	}

	return ""
}

// extractTextFromNode extracts text content from a specific HTML node
func (app *App) extractTextFromNode(n *html.Node) string {
	var text strings.Builder

	var extract func(*html.Node)
	extract = func(n *html.Node) {
		if n.Type == html.TextNode {
			content := strings.TrimSpace(n.Data)
			if content != "" {
				text.WriteString(content + " ")
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}

	extract(n)
	return strings.TrimSpace(text.String())
}
