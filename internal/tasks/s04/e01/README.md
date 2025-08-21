# S04E01 - Image Restoration and Description Task

## Overview

The S04E01 task implements an intelligent image restoration and description orchestrator that:

1. **Fetches initial photos** from the central "photos" API
2. **Analyzes each photo** using GPT-4 Vision to determine restoration needs
3. **Iteratively applies operations** (REPAIR, BRIGHTEN, DARKEN) based on image quality assessment
4. **Tracks filename changes** through the restoration process as the bot returns new filenames
5. **Selects photos showing Barbara** using consistent feature detection across multiple photos
6. **Generates a detailed Polish rysopis** (description) of Barbara's appearance

## Key Features

### Intelligent Vision Analysis
- Uses GPT-4 Vision to assess image quality and defects
- Detects issues: glitches, compression artifacts, over/underexposure
- Makes informed decisions about restoration operations
- Identifies consistent subjects across multiple photos

### Unified LLM-Based Parsing
- Uses GPT-4 to parse all Polish bot responses reliably
- Extracts photo URLs and filenames from initial photo responses
- Parses operation results to track filename changes and success status
- Handles varied response formats and chatty Polish messages
- Falls back to regex patterns when LLM parsing fails
- Sends only filenames in commands (not URLs) as required

### Adaptive Processing
- Stops when image quality is optimal or improvements plateau
- Limits iterations per photo to prevent infinite loops
- Handles download failures and API errors gracefully
- Provides comprehensive processing statistics

## Operations

- **REPAIR**: Applied for visible glitches, compression artifacts, blocky noise, scan lines, corruption
- **BRIGHTEN**: Applied when image looks underexposed, faces lack detail in shadows
- **DARKEN**: Applied when image looks overexposed, highlights are blown, faces washed out  
- **NOOP**: Applied when image quality is optimal or no improvements needed

## Usage

### Running the Task
```bash
# Set required environment variables
export AI_DEVS_API_KEY="your-api-key"
export OPENAI_API_KEY="your-openai-key"

# Run the task
go run ./cmd s04e01
```

## Architecture

### Core Components

- **Handler**: CLI interface and task coordination
- **Service**: Core business logic and workflow orchestration
- **Vision Analysis**: GPT-4 Vision integration for image assessment
- **Bot Communication**: Polish response parsing and command sending
- **Session Tracking**: State management through restoration iterations

### Data Flow

1. **Initial Photos**: `START` → Bot → Photo URLs/filenames
2. **Image Analysis**: Download → Vision Analysis → Operation Decision
3. **Operation Commands**: `OPERATION FILENAME` → Bot → LLM Parse Response → New filename + status
4. **Iteration Control**: Track attempts, detect plateaus, manage timeouts
5. **Subject Selection**: Identify photos of the same person (Barbara)
6. **Final Description**: Generate Polish rysopis from selected photos

### Error Handling

- **Resilient parsing**: Primary LLM-based parsing with regex fallbacks for all responses
- **URL construction**: Smart base URL detection from initial responses (handles centrala.ag3nts.org/dane/barbara/ format)
- **Graceful degradation**: Continue processing even if some photos fail
- **Rate limiting**: Built-in delays between operations
- **Timeout management**: Context-based cancellation with reasonable timeouts

## Configuration

### Environment Variables

- `AI_DEVS_API_KEY`: Required API key for central bot communication
- `OPENAI_API_KEY`: Required OpenAI API key for vision analysis

### Processing Limits

- Maximum 5 iterations per photo (`MaxIterationsPerPhoto`)
- 10-minute overall timeout
- Rate limiting: 500ms between photos, 1s between operations

## Output

### Processing Statistics
- Total and processed photo counts
- Operations performed by type (REPAIR/BRIGHTEN/DARKEN)
- Iterations per photo
- Processing time and token costs

### Final Polish Rysopis
Comprehensive description including:
- Orientacyjny wiek i wzrost (approximate age and height)
- Budowa ciała (body build)
- Kształt twarzy i rysy (face shape and features)
- Włosy (kolor, długość, fryzura)
- Oczy, nos, usta, brwi
- Znaki szczególne (distinctive marks)
- Ubiór i akcesoria (clothing and accessories)
- Uncertainty markers where appropriate

### Example Bot Response Parsing
```
Input: "Słuchaj! mam dla Ciebie fotki o które prosiłeś. IMG_559.PNG, IMG_1410.PNG, IMG_1443.PNG, IMG_1444.PNG. Wszystkie siedzą sobie tutaj: https://centrala.ag3nts.org/dane/barbara/. Pamiętaj, że zawsze mogę poprawić je dla Ciebie (polecenia: REPAIR/DARKEN/BRIGHTEN)."

LLM Output:
{
  "IMG_559.PNG": "https://centrala.ag3nts.org/dane/barbara/IMG_559.PNG",
  "IMG_1410.PNG": "https://centrala.ag3nts.org/dane/barbara/IMG_1410.PNG",
  "IMG_1443.PNG": "https://centrala.ag3nts.org/dane/barbara/IMG_1443.PNG",
  "IMG_1444.PNG": "https://centrala.ag3nts.org/dane/barbara/IMG_1444.PNG"
}
```

## Example Output

```
=== Image Restoration Results ===
Photos processed: 8
Operations performed: 15
Photos selected for description: 5
===================================

Final rysopis: Barbara to kobieta w wieku około 28-32 lat, wzrostu średniego około 165-170 cm, o szczupłej budowie ciała. Ma ciemnobrązowe włosy długości do ramion, zazwyczaj rozpuszczone lub lekko pofalowane. Twarz owalna z regularnymi rysami - nos prosty średniej wielkości, usta średnie, brwi naturalne ciemne. Oczy ciemne, prawdopodobnie brązowe. Na większości zdjęć nosi ciemne okulary w prostokątnej oprawie. Ubiór casualowy - jasne bluzki, ciemne spodnie lub jeansy. Brak widocznych tatuaży czy charakterystycznych znamion na twarzy. Na jednym ze zdjęć widoczny mały plecak lub torba na ramieniu.
```

## Safety Notes

- This is a test task - photos do not depict real people
- Descriptions are generated for academic/testing purposes only
- No identity speculation or personal information inference
- Focus strictly on visible physical attributes

## Dependencies

- OpenAI GPT-4 Vision API for image analysis
- HTTP client for bot communication
- Context management for timeouts
- JSON parsing for structured responses