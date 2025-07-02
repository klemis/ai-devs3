# S02E05 - Arxiv Document Analysis

This command implements the "arxiv" task which processes Professor Maj's intercepted multi-format research publication and answers questions from central command.

## Overview

The S02E05 implementation provides comprehensive analysis of academic documents containing:
- HTML text content
- Embedded images requiring visual analysis
- Audio files requiring transcription
- Question answering based on consolidated content

## Features

### Multi-Modal Content Processing
- **Text Processing**: HTML parsing with markdown conversion
- **Image Analysis**: Visual content description using OpenAI Vision
- **Audio Transcription**: Speech-to-text using OpenAI Whisper
- **Content Consolidation**: Merges all content types into unified context

### Intelligent Caching System
- File-based caching for expensive operations (image analysis, audio transcription)
- Cache files stored in `data/s02e05/` directory
- Naming convention: `{content_type}_{index}.{operation_type}.txt`
- Automatic cache validation and reuse

### Question Answering
- Context-aware LLM responses using consolidated document content
- Structured prompting for accurate, concise answers
- Single-sentence response format as required

## Usage

### Prerequisites
- Set `AI_DEVS_API_KEY` environment variable
- Ensure OpenAI API access for Vision and Whisper models
- Go 1.24.3+ with required dependencies

### Running the Command
```bash
export AI_DEVS_API_KEY="your-api-key-here"
cd ai-devs3
go run ./cmd/s02e05/main.go
```

### Expected Output
```
2024/01/01 12:00:00 Starting S02E05 arxiv processing task
2024/01/01 12:00:01 Fetching article from: https://c3ntrala.ag3nts.org/dane/arxiv-draft.html
2024/01/01 12:00:02 Fetching questions from: https://c3ntrala.ag3nts.org/data/{API_KEY}/arxiv.txt
2024/01/01 12:00:03 Processing article content...
2024/01/01 12:00:04 Found 3 images to process
2024/01/01 12:00:05 Found 2 audio files to process
2024/01/01 12:00:10 Found 5 questions to answer
2024/01/01 12:00:15 Successfully submitted arxiv analysis, response: {...}
```

## Implementation Details

### Data Flow
1. **Fetch**: Download HTML article and questions file
2. **Parse**: Extract text, image URLs, and audio URLs from HTML
3. **Process**: Analyze images and transcribe audio (with caching)
4. **Consolidate**: Merge all content into comprehensive context
5. **Answer**: Generate responses to each question using LLM
6. **Submit**: Send structured JSON response to central command

### Cache Structure
```
data/s02e05/
├── image_01.description.txt    # Image analysis cache
├── image_02.description.txt
├── audio_01.transcript.txt     # Audio transcription cache
├── audio_02.transcript.txt
└── temp_audio_01.mp3          # Temporary files (auto-cleaned)
```

### Response Format
```json
{
    "task": "arxiv",
    "apikey": "your-api-key",
    "answer": {
        "01": "Professor Maj's research focuses on quantum mechanics applications.",
        "02": "The experiment was conducted in Warsaw laboratory facilities.",
        "03": "Results showed 87% efficiency improvement over baseline methods."
    }
}
```

## Architecture

### Concurrency
- Parallel processing of images and audio files
- Thread-safe cache operations
- Worker pool pattern for efficient resource utilization

### Cache Reset
Remove cache directory to force reprocessing:
```bash
rm -rf data/s02e05/
```

## Performance Considerations

### Optimization Features
- Intelligent caching prevents redundant API calls
- Concurrent processing reduces total execution time
- Context chunking handles large documents efficiently
- Smart URL resolution for reliable content fetching

### Resource Usage
- Memory: ~100-500MB depending on content size
- Network: Variable based on media file sizes
- API Tokens: Estimated 2000-5000 tokens per execution
