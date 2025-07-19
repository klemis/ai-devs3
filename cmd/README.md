# AI-DEVS3 CLI

A unified command-line interface for running AI-DEVS3 tasks and challenges.

## Overview

The AI-DEVS3 CLI provides a centralized way to execute various AI development tasks organized by seasons and episodes. Each task is implemented as a separate command with its own configuration and execution logic.

## Installation

```bash
# Build the CLI
go build -o bin/ai-devs3 ./cmd/main.go

# Or use go run
go run ./cmd/main.go [command]
```

## Usage

### Basic Commands

```bash
# Show help
./bin/ai-devs3 --help

# List all available tasks
./bin/ai-devs3 list

# Show version
./bin/ai-devs3 version

# Get help for a specific task
./bin/ai-devs3 s01e01 --help
```

### Running Tasks

```bash
# Season 1 tasks
./bin/ai-devs3 s01e01    # Robot Authentication
./bin/ai-devs3 s01e02    # RoboISO Verification
./bin/ai-devs3 s01e03    # JSON Data Processing
ai-devs3 s01e05    # Text Censoring

# Season 2 tasks
./bin/ai-devs3 s02e01    # Audio Transcription and Analysis
./bin/ai-devs3 s02e02    # Map Analysis
./bin/ai-devs3 s02e03    # Robot Image Generation
./bin/ai-devs3 s02e04    # File Categorization
```

## Configuration

The CLI requires several environment variables to be set:

### Required
- `AI_DEVS_API_KEY`: Your AI-DEVS API key
- `OPENAI_API_KEY`: OpenAI API key for LLM operations

### Optional
- `AI_DEVS_BASE_URL`: Base URL for AI-DEVS API (default: https://c3ntrala.ag3nts.org)
- `OPENAI_MODEL`: OpenAI model to use (default: gpt-4o-mini)
- `OLLAMA_BASE_URL`: Ollama server URL (default: http://localhost:11434)
- `OLLAMA_MODEL`: Ollama model to use (default: llama3.2)
- `CACHE_DIR`: Directory for caching (default: data)

### Setup Example

```bash
export AI_DEVS_API_KEY="your-api-key-here"
export OPENAI_API_KEY="your-openai-key-here"
export OLLAMA_BASE_URL="http://localhost:11434"
export OLLAMA_MODEL="llama3.2"
```

## Task Categories

### Season 1 (Foundations)
- **s01e01**: Robot Authentication - Login to robot system by answering questions
- **s01e02**: RoboISO Verification - Communicate using RoboISO 2230 protocol
- **s01e03**: JSON Data Processing - Process test data, solve math problems, answer questions
- **s01e05**: Text Censoring - Censor personal information using local LLM

### Season 2 (Advanced Processing)
- **s02e01**: Audio Transcription and Analysis - Transcribe audio files and analyze content
- **s02e02**: Map Analysis - Analyze map fragments to identify Polish cities
- **s02e03**: Robot Image Generation - Generate robot images using DALL-E
- **s02e04**: File Categorization - Process and categorize files into people/hardware

## Architecture

The CLI uses Cobra for command structure with the following organization:

```
cmd/
├── main.go              # Main CLI entry point
└── README.md           # This file

internal/tasks/
├── s01/                # Season 1 tasks
│   ├── e01/            # Episode 1
│   │   ├── command.go  # Cobra command definition
│   │   ├── handler.go  # Business logic orchestration
│   │   ├── service.go  # Core service implementation
│   │   └── models.go   # Data structures
│   └── ...
└── s02/                # Season 2 tasks
    └── ...
```

### Command Structure

Each task follows a consistent pattern:
- **Command**: Defines CLI interface and flags
- **Handler**: Orchestrates the task execution
- **Service**: Contains business logic and API calls
- **Models**: Data structures and types

## Development

### Adding New Tasks

1. Create new task directory: `internal/tasks/s0X/eYY/`
2. Implement the four core files: `command.go`, `handler.go`, `service.go`, `models.go`
3. Add command to `cmd/main.go` with proper import alias
4. Update the list command with new task info
5. Follow the established patterns for dependency injection and error handling

### Testing

```bash
# Test individual components
go test ./internal/tasks/s01/e01/...

# Test configuration
go run ./cmd/test-config/main.go

# Test HTTP client
go run ./cmd/test-http/main.go

# Test OpenAI integration
go run ./cmd/test-openai/main.go

# Test caching
go run ./cmd/test-cache/main.go
```

## Dependencies

- **Cobra**: CLI framework
- **OpenAI Go SDK**: LLM, audio processing, and image generation
- **Standard Library**: HTTP, JSON, file operations
- **Custom Packages**: Image processing, caching, configuration management

## Error Handling

The CLI includes comprehensive error handling:
- Configuration validation on startup
- Task-specific error types with step tracking
- Graceful degradation for non-critical failures
- Detailed error messages with context
- Proper error propagation through service layers

## Performance Considerations

- File-based caching for expensive operations (OCR, transcription, image analysis)
- Concurrent processing with worker pools
- Configurable timeouts and retry logic
- Memory-efficient streaming for large files
- Context-aware cancellation for long-running operations

## Contributing

1. Follow the established task structure
2. Include comprehensive error handling
3. Add appropriate logging
4. Write tests for new functionality
5. Update documentation

## License

This project is part of the AI-DEVS3 course materials.