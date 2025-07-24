package ocr

// OCRResult represents the result of OCR processing
type OCRResult struct {
	URL           string `json:"url"`
	ExtractedText string `json:"extracted_text"`
	Error         string `json:"error,omitempty"`
}

// ImageData represents binary image data with metadata
type ImageData struct {
	Data []byte
	URL  string
	Size int64
}

// OCRRequest represents a request for OCR processing
type OCRRequest struct {
	ImageURL string `json:"image_url"`
	Format   string `json:"format,omitempty"`
}
