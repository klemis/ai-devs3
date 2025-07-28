# AI-DEVS3 Solutions

This repository contains Go implementations for AI-DEVS3 course tasks, featuring multi-modal content processing, LLM integration, and automated task solving.

## Overview

AI-DEVS3 is a comprehensive solution suite for processing various AI-related tasks including:
- Multi-modal document analysis (text, images, audio)
- LLM-powered question answering
- Vision API integration for image analysis
- Audio transcription and processing
- Automated content categorization
- Map fragment analysis
- Robot interaction protocols

## Architecture

### Core Components

- **App Layer** (`internal/app/`): Task-specific business logic and orchestration
- **Service Layer** (`internal/service/`): External API integrations and utilities
- **Domain Layer** (`internal/domain/`): Data models and structures
- **Commands** (`cmd/`): Executable entry points for each task

### Key Services

- **LLMClient**: OpenAI API integration for text generation, vision, and audio processing
- **HTTPClient**: Web scraping and API communication
- **ImageProcessor**: Image analysis and processing utilities
- **OllamaClient**: Local LLM integration for specific tasks

## Task Implementations

### S01 Series - Basic Integrations

- **S01E01**: Robot authentication and question answering
- **S01E02**: RoboISO protocol implementation
- **S01E03**: JSON data processing and validation
- **S01E05**: Text censoring using local LLM (Ollama)

### S02 Series - Advanced Processing

- **S02E01**: Audio transcription and analysis
- **S02E02**: Map fragment identification and city recognition
- **S02E03**: Robot image generation using DALL-E
- **S02E04**: File categorization (people vs hardware)
- **S02E05**: Multi-modal document analysis (arxiv task)

### Special Tasks

- **OCR**: Image text extraction

## Installation

### Prerequisites
- Go 1.24.3+
- OpenAI API key
- (Optional) Ollama for local LLM tasks

### Setup
```bash
git clone <repository>
cd ai-devs3
go mod download
```

### Environment Variables
```bash
export AI_DEVS_API_KEY="your-api-key-here"
export OPENAI_API_KEY="your-openai-key"
```

## Usage

### Running Individual Tasks
```bash
# S02E05 - Arxiv document processor
go run ./cmd/s02e05/main.go

# S02E01 - Audio transcription
go run ./cmd/s02e01/main.go

# S02E04 - File categorization
go run ./cmd/s02e04/main.go
```

### Building All Tasks
```bash
go build ./cmd/...
```

## Configuration

### Cache Directory Structure
```
data/
├── s02e04/           # File categorization cache
├── s02e05/           # Arxiv processor cache
│   ├── consolidated_context.md
│   ├── image_01.description.txt
│   ├── audio_01.transcript.txt
│   └── ...
└── ...
```

### API Integration
- **OpenAI**: GPT models, Vision API, Whisper, DALL-E
- **AI-DEVS**: Central command API for task submission
- **Ollama**: Local LLM for censoring tasks

## Dependencies

### Core Dependencies
```go
github.com/openai/openai-go v1.3.0
golang.org/x/net v0.41.0
```

### Optional Dependencies
- Ollama server for local LLM tasks

### Testing
```bash
go test ./...
go vet ./...
```

## License

This project is for educational purposes as part of the AI-DEVS3 course.

### Cache Management
Clear cache to force reprocessing:
```bash
rm -rf data/s02e05/
```
