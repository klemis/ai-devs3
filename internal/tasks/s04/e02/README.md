# S04E02 - Text Classification Research Task

## Overview

The S04E02 task involves classifying text lines using a fine-tuned model to determine which lines are "reliable" (1) or "unreliable" (0). The task processes verification data and reports only the IDs of lines classified as reliable to a central endpoint.

## Task Flow

1. **Read Verification Data**: Parse lines from `data/s04e02/verify.txt`
2. **Extract Content**: Remove ID prefixes (e.g., "01=kanto,saka,brunn" → "kanto,saka,brunn")
3. **Classify Lines**: Send each line to fine-tuned model with exact system prompt
4. **Collect Results**: Gather IDs of lines classified as reliable (1)
5. **Report**: Submit reliable line IDs to central endpoint

## Data Format

### Input Format (verify.txt)
```
01=kanto,saka,brunn
02=kezka,barru,asal
03=dwylo,blodau,cerdd
...
```

### Classification Request Format
```json
{
  "messages": [
    {
      "role": "system",
      "content": "Classify input strings into reliable (1) or unreliable (0). Treat inputs as arbitrary tokens and output only 0 or 1. Do not infer semantics or language."
    },
    {
      "role": "user", 
      "content": "kanto,saka,brunn"
    }
  ]
}
```

### Expected Response
- `"1"` for reliable lines
- `"0"` for unreliable lines

### Central Report Format
```json
{
  "task": "research",
  "apikey": "YOUR_API_KEY",
  "answer": ["01", "02", "09", "10"]
}
```

## Configuration

### Required Environment Variables
- `AI_DEVS_API_KEY`: API key for central reporting

### Model Configuration
The task uses the fine-tuned OpenAI model: `ft:gpt-4o-mini-2024-07-18:personal:validate:C7MNVVbk`

## Implementation Details

### Fine-Tuned Model
- **Model ID**: `ft:gpt-4o-mini-2024-07-18:personal:validate:C7MNVVbk`
- **Integration**: Uses `[@llm]` package with `ClassifyWithFineTunedModel` method
- **Temperature**: 0.1 for consistent classification results

### System Prompt
The exact system prompt that must be used:
```
"Classify input strings into reliable (1) or unreliable (0). Treat inputs as arbitrary tokens and output only 0 or 1. Do not infer semantics or language."
```

### Response Processing
- Expected responses: `"1"` (reliable) or `"0"` (unreliable)
- Invalid responses are treated as unreliable with warnings logged
- Response parsing includes trimming whitespace and integer conversion

### ID Format
- IDs are zero-padded two-digit strings: "01", "02", "03", etc.
- Only IDs of reliable classifications are included in the final answer

### Reporting Pattern
- Uses `submitFinalResponse` pattern with `BuildAIDevsResponse` and `PostReport`
- Consistent with other AI-DEVS tasks implementation

### Error Handling
- Lines that fail to classify are skipped with warnings
- Invalid model responses are treated as unreliable
- Task continues processing remaining lines if individual lines fail

## Usage

```bash
# Execute the S04E02 task
ai-devs3 s04e02
```

## Files

- `command.go`: CLI command definition and documentation
- `handler.go`: Main task execution handler
- `service.go`: Core business logic and processing
- `models.go`: Data structures and constants
- `README.md`: This documentation

## Implementation Status

✅ **Model Integration**: Fine-tuned model integrated using `[@llm]` package  
✅ **Standard Reporting**: Uses `submitFinalResponse` pattern consistent with other tasks  
✅ **Response Validation**: Model responses are validated and parsed correctly  
✅ **Error Handling**: Graceful handling of classification failures  

### Potential Improvements
1. **Enhanced Error Handling**: Implement retry logic for model calls
2. **Performance Optimization**: Add concurrent processing for multiple lines
3. **Metrics Collection**: Add classification accuracy tracking

## Training Data Reference

The task includes training data in JSONL format:
- `data/s04e02/dataset.jsonl`: Training examples with system/user/assistant messages
- `data/s04e02/correct.txt`: Examples of reliable text patterns
- `data/s04e02/incorrect.txt`: Examples of unreliable text patterns

## Central Endpoint

Reports are sent to: `https://c3ntrala.ag3nts.org/report`

Note: The endpoint uses "c3ntrala" (with "3") - the old address without "3" is invalid.