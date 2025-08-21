package e01

import "time"

// TaskResult represents the final result of the S04E01 task
type TaskResult struct {
	Response        string           `json:"response"`
	PhotosProcessed int              `json:"photos_processed"`
	OperationsCount int              `json:"operations_count"`
	SelectedPhotos  []string         `json:"selected_photos"`
	FinalRysopis    string           `json:"final_rysopis"`
	ProcessingStats *ProcessingStats `json:"processing_stats,omitempty"`
}

// ProcessingStats holds detailed statistics about the task execution
type ProcessingStats struct {
	TotalPhotos        int            `json:"total_photos"`
	ProcessedPhotos    int            `json:"processed_photos"`
	SelectedPhotos     int            `json:"selected_photos"`
	TotalOperations    int            `json:"total_operations"`
	OperationsByType   map[string]int `json:"operations_by_type"`
	PhotoIterations    map[string]int `json:"photo_iterations"`
	ProcessingTime     float64        `json:"processing_time"`
	VisionAnalysisCost int            `json:"vision_analysis_cost"`
	StartTime          time.Time      `json:"start_time"`
	EndTime            time.Time      `json:"end_time"`
}

// PhotosResponse represents the initial response from the photos API
type PhotosResponse struct {
	Message string                 `json:"message"`
	Photos  map[string]interface{} `json:"photos,omitempty"`
	URLs    []string               `json:"urls,omitempty"`
	Files   []string               `json:"files,omitempty"`
}

// PhotoInfo holds information about a photo being processed
type PhotoInfo struct {
	CurrentFilename string    `json:"current_filename"`
	OriginalURL     string    `json:"original_url"`
	Iterations      int       `json:"iterations"`
	Operations      []string  `json:"operations"`
	Status          string    `json:"status"` // "processing", "optimal", "failed", "abandoned"
	LastUpdated     time.Time `json:"last_updated"`
	Selected        bool      `json:"selected"` // Whether this photo shows Barbara
}

// VisionAnalysisRequest represents input for vision analysis
type VisionAnalysisRequest struct {
	Filename  string `json:"filename"`
	ImageURL  string `json:"image_url,omitempty"`
	ImageData []byte `json:"image_data,omitempty"`
}

// VisionAnalysisResponse represents the vision analysis result
type VisionAnalysisResponse struct {
	Thinking         string   `json:"_thinking"`
	Filename         string   `json:"filename"`
	Decision         string   `json:"decision"` // "REPAIR", "BRIGHTEN", "DARKEN", "NOOP"
	ExpectMorePasses bool     `json:"expect_more_passes"`
	IsSubject        bool     `json:"is_subject"`              // Whether this shows the target woman
	QualityScore     int      `json:"quality_score,omitempty"` // 1-10
	IssuesDetected   []string `json:"issues_detected,omitempty"`
}

// BotResponseParser represents parsed bot response
type BotResponseParser struct {
	Thinking       string  `json:"_thinking"`
	LatestFilename string  `json:"latest_filename"`
	Url            string  `json:"url"`
	Success        *bool   `json:"success"`                  // null if unknown
	SuggestedNext  *string `json:"suggested_next,omitempty"` // "REPAIR", "BRIGHTEN", "DARKEN", null
	Note           string  `json:"note"`
}

// OperationCommand represents a command to send to the bot
type OperationCommand struct {
	Operation string    `json:"operation"` // "REPAIR", "BRIGHTEN", "DARKEN", "NOOP"
	Filename  string    `json:"filename"`
	Timestamp time.Time `json:"timestamp"`
}

// RestorationSession tracks the complete restoration process
type RestorationSession struct {
	Photos          map[string]*PhotoInfo `json:"photos"`
	SelectedPhotos  []string              `json:"selected_photos"`
	Operations      []OperationCommand    `json:"operations"`
	TotalIterations int                   `json:"total_iterations"`
	StartTime       time.Time             `json:"start_time"`
	Status          string                `json:"status"` // "running", "completed", "failed"
}

// RysopisRequest represents input for generating the final Polish description
type RysopisRequest struct {
	SelectedPhotos []PhotoInfo `json:"selected_photos"`
	ImageURLs      []string    `json:"image_urls"`
	Context        string      `json:"context,omitempty"`
}

// Constants for operation types
const (
	OperationRepair   = "REPAIR"
	OperationBrighten = "BRIGHTEN"
	OperationDarken   = "DARKEN"
	OperationNoop     = "NOOP"
)

// Constants for photo status
const (
	StatusProcessing = "processing"
	StatusOptimal    = "optimal"
	StatusFailed     = "failed"
	StatusAbandoned  = "abandoned"
)

// Constants for session status
const (
	SessionRunning   = "running"
	SessionCompleted = "completed"
	SessionFailed    = "failed"
)

// Max iterations per photo to prevent infinite loops
const MaxIterationsPerPhoto = 5

// Vision analysis prompts
const VisionPromptSystem = `You receive one image and its filename. Evaluate:

Is the subject a person likely to be the same woman as in other accepted photos? If uncertain, note uncertainty but still judge quality.
Identify issues: glitches/noise, over/underexposure, or adequate.
Decide next action among REPAIR, BRIGHTEN, DARKEN, or NOOP with rationale.
If REPAIR/BRIGHTEN/DARKEN, estimate if another pass might be helpful after this one.

Return JSON:
{
"_thinking": "detailed reasoning behind the potential decision",
"filename": "<current filename>",
"decision": "REPAIR|BRIGHTEN|DARKEN|NOOP",
"expect_more_passes": true|false,
"is_subject": true|false,
"quality_score": 1-10,
"issues_detected": ["list", "of", "issues"]
}`

const BotParsingPromptSystem = `You receive the bot's latest JSON and message strings (Polish). Extract:

_thinking: detailed reasoning
latest_filename: new filename if present; else keep previous.
success: true/false/unknown
suggested_next: optional textual hint if the bot implies another operation
note: concise summary of what happened

You must respond with ONLY valid JSON in this exact format, no markdown, no code blocks, no extra text:
{
"_thinking": "detailed reasoning",
"latest_filename": "<string>",
"success": true|false|null,
"suggested_next": "REPAIR|BRIGHTEN|DARKEN|null",
"note": "<short>"
}`

const RysopisPromptSystem = `Jesteś ekspertem w analizie zdjęć i tworzeniu rysopisów. To zadanie testowe; zdjęcia nie przedstawiają prawdziwych osób. Na podstawie dostarczonych, najlepszych wersji zdjęć przygotuj szczegółowy rysopis Barbary po polsku.

Wytyczne:
- Opisz tylko to, co widać. Unikaj spekulacji i identyfikacji.
- Skup się na powtarzalnych cechach widocznych na co najmniej dwóch zdjęciach.
- Uwzględnij: orientacyjny wiek i wzrost, budowę ciała, kształt twarzy, cerę, włosy (kolor, długość, fryzura), oczy (kolor/kształt), nos, usta, brwi, znaki szczególne (blizny, pieprzyki, tatuaże), elementy ubioru i akcesoriów (okulary, biżuteria, zegarek), ewentualny zarost brwi czy makijaż; jeśli widoczne – dłonie/paznokcie, buty, torba/plecak.
- Zaznacz niepewności i różnice między zdjęciami, jeśli występują.
- Styl: rzeczowy, precyzyjny, zwięzły, pełne polskie znaki (UTF‑8).
- Wynik: pojedynczy akapit lub 2–3 akapity, bez punktowania.

Zwróć tylko tekst rysopisu, bez żadnych dodatkowych komentarzy czy instrukcji.`
