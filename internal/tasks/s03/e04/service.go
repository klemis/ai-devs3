package e04

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the Barbara search task
type Service struct {
	httpClient *http.Client
	llmClient  *openai.Client
}

// NewService creates a new service instance
func NewService(httpClient *http.Client, llmClient *openai.Client) *Service {
	return &Service{
		httpClient: httpClient,
		llmClient:  llmClient,
	}
}

// ExecuteTask executes the complete S03E04 Barbara search task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	startTime := time.Now()

	log.Println("Starting Barbara search task...")

	// Step 1: Read barbara.txt
	barbaraText, err := s.fetchBarbaraFile(ctx)
	if err != nil {
		return nil, errors.NewTaskError("s03e04", "fetch_barbara_file", err)
	}

	// Step 2: Parse names and cities using LLM
	parsedData, err := s.parseNamesAndCities(ctx, barbaraText)
	if err != nil {
		return nil, errors.NewTaskError("s03e04", "parse_data", err)
	}

	log.Printf("Parsed %d names and %d cities from barbara.txt", len(parsedData.Names), len(parsedData.Cities))

	// Step 3: Perform BFS search
	searchResult, err := s.performBFSSearch(ctx, apiKey, parsedData)
	if err != nil {
		return nil, errors.NewTaskError("s03e04", "bfs_search", err)
	}

	// Step 4: Submit the answer
	response, err := s.submitBarbaraLocation(ctx, apiKey, searchResult.BarbaraLocation)
	if err != nil {
		return nil, errors.NewTaskError("s03e04", "submit_response", err)
	}

	processingTime := time.Since(startTime).Seconds()

	return &TaskResult{
		Response:         response,
		BarbaraLocation:  searchResult.BarbaraLocation,
		TotalRequests:    searchResult.RequestCount,
		ProcessingTime:   processingTime,
		OriginalCities:   parsedData.Cities,
		DiscoveredCities: s.getDiscoveredCities(searchResult),
	}, nil
}

// fetchBarbaraFile retrieves the barbara.txt file from the API
func (s *Service) fetchBarbaraFile(ctx context.Context) (string, error) {
	log.Println("Fetching barbara.txt from API...")

	content, err := s.httpClient.FetchData(ctx, "https://c3ntrala.ag3nts.org/dane/barbara.txt")
	if err != nil {
		return "", fmt.Errorf("failed to fetch barbara.txt: %w", err)
	}

	log.Printf("Retrieved barbara.txt (%d characters)", len(content))
	log.Printf("Barbara.txt content: %s", content)
	return content, nil
}

// parseNamesAndCities uses LLM to extract names and cities from the text
func (s *Service) parseNamesAndCities(ctx context.Context, text string) (*ParsedData, error) {
	log.Println("Parsing names and cities using LLM...")

	systemPrompt := `You are an expert text analyzer. Extract ALL first names of people and ALL polish city names from the given text.

	RULES:
		1. Extract every person's first name mentioned in the text (first names only)
		2. Extract every city name mentioned in the text
		3. Remove diacritics (ą→a, ę→e, ś→s, ć→c, ł→l, ń→n, ó→o, ź→z, ż→z)
		4. Convert all names and cities to UPPERCASE
		5. Return only unique entries (no duplicates)

		Return the result as JSON in this exact format:
		{
  			"names": ["NAME1", "NAME2", ...],
  			"cities": ["CITY1", "CITY2", ...]
    	}
     	Example response:
       	{
       		"names": ["TOMASZ", "JAN", "ALEKSANDER"],
       		"cities": ["WARSZAWA", "POZNAN", "GDANSK"]
       }`

	userPrompt := fmt.Sprintf("Extract all first names and cities from this text:\n\n%s", text)

	response, err := s.llmClient.GetAnswerWithContext(ctx, systemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse names and cities: %w", err)
	}

	log.Printf("LLM parsing response: %s", response)

	var parsedData ParsedData
	if err := json.Unmarshal([]byte(response), &parsedData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parsed data: %w", err)
	}

	// Normalize the data (remove diacritics and ensure uppercase)
	parsedData.Names = s.normalizeStrings(parsedData.Names)
	parsedData.Cities = s.normalizeStrings(parsedData.Cities)

	log.Printf("Normalized names: %v", parsedData.Names)
	log.Printf("Normalized cities: %v", parsedData.Cities)

	return &parsedData, nil
}

// performBFSSearch performs breadth-first search to find Barbara's location
func (s *Service) performBFSSearch(ctx context.Context, apiKey string, parsedData *ParsedData) (*SearchState, error) {
	log.Println("Starting BFS search for Barbara...")

	state := &SearchState{
		QueueNames:     make([]string, len(parsedData.Names)),
		QueueCities:    make([]string, len(parsedData.Cities)),
		VisitedNames:   make(map[string]struct{}),
		VisitedCities:  make(map[string]struct{}),
		OriginalCities: make(map[string]struct{}),
		RequestCount:   0,
	}

	copy(state.QueueNames, parsedData.Names)
	copy(state.QueueCities, parsedData.Cities)

	// Mark original cities
	for _, city := range parsedData.Cities {
		state.OriginalCities[city] = struct{}{}
	}

	log.Printf("Initial names queue: %v", state.QueueNames)
	log.Printf("Initial cities queue: %v", state.QueueCities)
	log.Printf("Original cities: %v", parsedData.Cities)

	rateLimiter := time.NewTicker(200 * time.Millisecond) // 5 req/s = 200ms between requests
	defer rateLimiter.Stop()

	for len(state.QueueNames) > 0 || len(state.QueueCities) > 0 {
		// Process names - search for people using /people endpoint
		if len(state.QueueNames) > 0 {
			name := state.QueueNames[0]
			state.QueueNames = state.QueueNames[1:]

			if _, visited := state.VisitedNames[name]; !visited {
				state.VisitedNames[name] = struct{}{}

				<-rateLimiter.C // Rate limiting
				if err := s.searchPeople(ctx, apiKey, name, state); err != nil {
					log.Printf("Error searching people for %s: %v", name, err)
				}
				state.RequestCount++
			}
		}

		// Process cities - search for places using /places endpoint
		if len(state.QueueCities) > 0 {
			city := state.QueueCities[0]
			state.QueueCities = state.QueueCities[1:]

			if _, visited := state.VisitedCities[city]; !visited {
				state.VisitedCities[city] = struct{}{}

				<-rateLimiter.C // Rate limiting
				found, err := s.searchPlaces(ctx, apiKey, city, state)
				if err != nil {
					log.Printf("Error searching places for %s: %v", city, err)
				}
				state.RequestCount++

				if found {
					log.Printf("Found Barbara in city: %s", city)
					state.BarbaraLocation = city
					return state, nil
				}
			}
		}

		// Safety check to prevent infinite loops
		if state.RequestCount > 1000 {
			return nil, fmt.Errorf("search exceeded maximum request limit")
		}
	}

	if state.BarbaraLocation == "" {
		return nil, fmt.Errorf("Barbara's location not found")
	}

	return state, nil
}

// searchPeople queries the /people endpoint with a person's name
func (s *Service) searchPeople(ctx context.Context, apiKey, name string, state *SearchState) error {
	log.Printf("=== SEARCHING PEOPLE for: %s ===", name)

	request := &BarbaraSearchRequest{
		APIKey: apiKey,
		Query:  name,
	}

	responseBody, err := s.httpClient.PostJSON(ctx, "https://c3ntrala.ag3nts.org/people", request)
	if err != nil {
		return fmt.Errorf("failed to search people: %w", err)
	}

	log.Printf("People API RAW response for %s: %s", name, responseBody)

	var response BarbaraSearchResponse
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		return fmt.Errorf("failed to parse people response: %w", err)
	}

	log.Printf("People API parsed - Code: %d, Message: %q", response.Code, response.Message)

	// Check for special case - BARBARA query might return location info
	if strings.ToUpper(name) == "BARBARA" {
		log.Printf("*** SPECIAL: Searching for BARBARA herself ***")
		log.Printf("*** BARBARA response code: %d ***", response.Code)
		log.Printf("*** BARBARA response message: %q ***", response.Message)
		log.Printf("*** BARBARA full response: %s ***", responseBody)
	}

	// Parse space-separated cities from people endpoint response
	if response.Message != "" && response.Message != "[**RESTRICTED DATA**]" && response.Code == 0 {
		items := strings.Fields(response.Message)
		log.Printf("People API extracted %d cities: %v", len(items), items)
		for _, item := range items {
			normalized := s.normalizeString(item)
			// /people endpoint returns cities - add to cities queue
			if _, visited := state.VisitedCities[normalized]; !visited {
				state.QueueCities = append(state.QueueCities, normalized)
				log.Printf("Added to cities queue: %s (from people search)", normalized)
			}
		}
	} else {
		log.Printf("People API: No valid data - Code: %d, Message: %q", response.Code, response.Message)
	}

	return nil
}

// searchPlaces queries the /places endpoint with a city name
func (s *Service) searchPlaces(ctx context.Context, apiKey, city string, state *SearchState) (bool, error) {
	log.Printf("=== SEARCHING PLACES for: %s ===", city)

	request := &BarbaraSearchRequest{
		APIKey: apiKey,
		Query:  city,
	}

	responseBody, err := s.httpClient.PostJSON(ctx, "https://c3ntrala.ag3nts.org/places", request)
	if err != nil {
		return false, fmt.Errorf("failed to search places: %w", err)
	}

	log.Printf("Places API RAW response for %s: %s", city, responseBody)

	var response BarbaraSearchResponse
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		return false, fmt.Errorf("failed to parse places response: %w", err)
	}

	log.Printf("Places API parsed - Code: %d, Message: %q", response.Code, response.Message)

	// Check if Barbara is mentioned in the response message
	barbaraFound := false
	if strings.Contains(strings.ToUpper(response.Message), "BARBARA") {
		// Only consider it a match if this city wasn't in the original note
		if _, wasOriginal := state.OriginalCities[city]; !wasOriginal {
			barbaraFound = true
			log.Printf("*** BARBARA FOUND in city %s (not original) ***", city)
		} else {
			log.Printf("*** BARBARA found in city %s but it was in original note - ignoring ***", city)
		}
	}

	// Parse space-separated people names from places endpoint response
	if response.Message != "" && response.Message != "[**RESTRICTED DATA**]" && response.Code == 0 {
		items := strings.Fields(response.Message)
		log.Printf("Places API extracted %d people: %v", len(items), items)
		for _, item := range items {
			normalized := s.normalizeString(item)
			// /places endpoint returns people names - add to names queue
			if _, visited := state.VisitedNames[normalized]; !visited {
				state.QueueNames = append(state.QueueNames, normalized)
				log.Printf("Added to names queue: %s (from places search)", normalized)
			}
		}
	} else {
		log.Printf("Places API: No valid data - Code: %d, Message: %q", response.Code, response.Message)
	}

	return barbaraFound, nil
}

// normalizeStrings normalizes a slice of strings
func (s *Service) normalizeStrings(strings []string) []string {
	result := make([]string, len(strings))
	for i, str := range strings {
		result[i] = s.normalizeString(str)
	}
	return result
}

// normalizeString removes diacritics and converts to uppercase
func (s *Service) normalizeString(input string) string {
	// Simple diacritics removal for Polish characters
	replacer := strings.NewReplacer(
		"ą", "a", "Ą", "A",
		"ć", "c", "Ć", "C",
		"ę", "e", "Ę", "E",
		"ł", "l", "Ł", "L",
		"ń", "n", "Ń", "N",
		"ó", "o", "Ó", "O",
		"ś", "s", "Ś", "S",
		"ź", "z", "Ź", "Z",
		"ż", "z", "Ż", "Z",
	)
	normalized := replacer.Replace(input)

	// Convert to uppercase
	return strings.ToUpper(normalized)
}

// getDiscoveredCities extracts newly discovered cities from search state
func (s *Service) getDiscoveredCities(state *SearchState) []string {
	var discovered []string
	for city := range state.VisitedCities {
		if _, wasOriginal := state.OriginalCities[city]; !wasOriginal {
			discovered = append(discovered, city)
		}
	}
	return discovered
}

// submitBarbaraLocation submits Barbara's location to the centrala API
func (s *Service) submitBarbaraLocation(ctx context.Context, apiKey, location string) (string, error) {
	log.Printf("Submitting Barbara's location: %s", location)

	response := s.httpClient.BuildAIDevsResponse("loop", apiKey, location)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit Barbara's location: %w", err)
	}

	return result, nil
}
