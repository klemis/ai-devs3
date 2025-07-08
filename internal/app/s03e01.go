package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// DocumentsAnswer represents the answer structure for the documents task
type DocumentsAnswer map[string]string

// Person represents a person with their attributes
type Person struct {
	Name       string   `json:"name"`
	Profession string   `json:"profession"`
	Skills     []string `json:"skills"`
	Location   string   `json:"location"`
	Status     string   `json:"status"`
	Relations  []string `json:"relations"`
}

// FactsKeywords represents processed facts with key information
type FactsKeywords struct {
	People   []Person `json:"people"`
	Sectors  []string `json:"sectors"`
	Keywords []string `json:"keywords"`
}

// ProcessedFacts represents the structure of processed facts cache
type ProcessedFacts map[string]FactsKeywords

// RunS03E01 processes factory security reports and generates Polish keywords
func (app *App) RunS03E01(apiKey string) (string, error) {
	log.Println("Starting S03E01 documents processing task")

	// Validate API key
	if apiKey == "" {
		return "", fmt.Errorf("API key is required")
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}

	var reportsDir, factsDir string

	// Step 1: Read all TXT report files
	// Adjust path based on current working directory
	if strings.Contains(wd, "cmd/s03e01") {
		// Running from debug mode
		reportsDir = "../../../lessons-md/pliki_z_fabryki"
		factsDir = "../../../lessons-md/pliki_z_fabryki/facts"
	} else {
		// Running from project root
		reportsDir = "../lessons-md/pliki_z_fabryki"
		factsDir = "../lessons-md/pliki_z_fabryki/facts"
	}

	// Get list of TXT files from reports directory
	txtFiles, err := app.getTXTFiles(reportsDir)
	if err != nil {
		return "", fmt.Errorf("failed to get TXT files: %w", err)
	}

	if len(txtFiles) == 0 {
		return "", fmt.Errorf("no TXT files found in %s", reportsDir)
	}

	log.Printf("Found %d TXT report files", len(txtFiles))

	// Step 2: Process facts folder for cross-referencing
	factsKeywords, err := app.processFactsFolder(factsDir)
	if err != nil {
		return "", fmt.Errorf("failed to process facts folder: %w", err)
	}

	log.Printf("Loaded processed facts from %d files", len(factsKeywords))

	// Step 3: Process each report file
	answer := make(DocumentsAnswer)

	for _, txtFile := range txtFiles {
		log.Printf("Processing report: %s", txtFile)

		// Read report content
		reportPath := filepath.Join(reportsDir, txtFile)
		reportContent, err := os.ReadFile(reportPath)
		if err != nil {
			log.Printf("Failed to read report %s: %v", txtFile, err)
			continue
		}

		// Generate keywords for this report
		keywords, err := app.generateKeywordsForReport(txtFile, string(reportContent), factsKeywords)
		if err != nil {
			log.Printf("Failed to generate keywords for %s: %v", txtFile, err)
			continue
		}

		answer[txtFile] = keywords
		log.Printf("Generated keywords for %s: %s", txtFile, keywords)
	}

	// Validate we have exactly 10 entries
	if len(answer) != 10 {
		return "", fmt.Errorf("expected exactly 10 reports, but processed %d", len(answer))
	}

	// Step 4: Submit response
	response := app.httpClient.BuildResponse("dokumenty", answer)
	result, err := app.httpClient.PostReport(response)
	if err != nil {
		return "", fmt.Errorf("failed to submit report: %w", err)
	}

	return result, nil
}

// getTXTFiles returns all .txt files from the specified directory
func (app *App) getTXTFiles(dir string) ([]string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	var txtFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(file.Name()), ".txt") {
			txtFiles = append(txtFiles, file.Name())
		}
	}

	return txtFiles, nil
}

// processFactsFolder processes facts folder and extracts key information using LLM
func (app *App) processFactsFolder(factsDir string) (ProcessedFacts, error) {
	processedFacts := make(ProcessedFacts)

	// Check if facts directory exists
	if _, err := os.Stat(factsDir); os.IsNotExist(err) {
		log.Printf("Facts directory %s does not exist, continuing without facts", factsDir)
		return processedFacts, nil
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Check if processed facts cache exists
	// Adjust path based on current working directory
	var cacheDir string
	if strings.Contains(wd, "cmd/s03e01") {
		// Running from debug mode
		cacheDir = "../../data/s03e01"
	} else {
		// Running from project root
		cacheDir = "data/s03e01"
	}

	cacheFile := filepath.Join(cacheDir, "processed_facts.json")
	if _, err := os.Stat(cacheFile); err == nil {
		log.Printf("Loading processed facts from cache: %s", cacheFile)
		return app.loadProcessedFactsCache(cacheFile)
	}

	log.Printf("Processing facts folder: %s", factsDir)

	// Read and process all facts files
	files, err := os.ReadDir(factsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read facts directory %s: %w", factsDir, err)
	}

	for _, file := range files {
		if file.IsDir() || strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		filePath := filepath.Join(factsDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read facts file %s: %v", file.Name(), err)
			continue
		}

		log.Printf("Processing facts file: %s", file.Name())
		keywords, err := app.extractFactsKeywords(file.Name(), string(content))
		if err != nil {
			log.Printf("Failed to extract keywords from %s: %v", file.Name(), err)
			continue
		}

		processedFacts[file.Name()] = keywords
	}

	// Save processed facts to cache
	if err := app.saveProcessedFactsCache(cacheFile, processedFacts); err != nil {
		log.Printf("Failed to save processed facts cache: %v", err)
	}

	return processedFacts, nil
}

// loadProcessedFactsCache loads processed facts from cache file
func (app *App) loadProcessedFactsCache(cacheFile string) (ProcessedFacts, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var processedFacts ProcessedFacts
	if err := json.Unmarshal(data, &processedFacts); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache file: %w", err)
	}

	return processedFacts, nil
}

// saveProcessedFactsCache saves processed facts to cache file
func (app *App) saveProcessedFactsCache(cacheFile string, processedFacts ProcessedFacts) error {
	// Create cache directory if it doesn't exist
	cacheDir := filepath.Dir(cacheFile)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(processedFacts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal processed facts: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	log.Printf("Saved processed facts cache to: %s", cacheFile)
	return nil
}

// extractFactsKeywords uses LLM to extract key information from facts file
func (app *App) extractFactsKeywords(filename, content string) (FactsKeywords, error) {
	systemPrompt := `
	<prompt_objective>
	Jesteś ekspertem w analizie dokumentów. Twoim zadaniem jest wydobycie kluczowych informacji z pliku faktów w formacie JSON, grupując informacje o osobach razem.
	</prompt_objective>

	<prompt_rules>
	1. Przeanalizuj treść pliku i wydobądź:
	   - People: tablica obiektów z informacjami o osobach
	   - Sectors: lista nazw sektorów/lokalizacji
	   - Keywords: wszystkie inne ważne słowa kluczowe

	2. Dla każdej osoby grupuj informacje:
	   - name: imię i nazwisko (w mianowniku)
	   - profession: zawód/rola
	   - skills: umiejętności, języki programowania, specjalizacje
	   - location: sektor/miejsce pracy/pobytu
	   - status: aktualny stan (np. "ukrywa się", "w ośrodku psychiatrycznym")
	   - relations: powiązania z innymi osobami

	3. Zwróć odpowiedź w formacie JSON z polskimi słowami kluczowymi
	4. Używaj form w mianowniku
	5. Jeśli kategoria jest pusta, zwróć pustą tablicę
	6. Unikaj duplikatów

	PRZYKŁAD ODPOWIEDZI:
	{
		"people": [
			{
				"name": "Jan Kowalski",
				"profession": "programista",
				"skills": ["Java", "Python"],
				"location": "Sektor A",
				"status": "ukrywa się",
				"relations": ["Anna Nowak"]
			}
		],
		"sectors": ["Sektor A", "Sektor B"],
		"keywords": ["roboty", "sztuczna inteligencja", "ruch oporu"]
	}
	</prompt_rules>`

	userPrompt := fmt.Sprintf("Wydobądź kluczowe informacje z pliku: %s\n\nTreść:\n%s", filename, content)

	response, err := app.llmClient.AnswerWithContext(systemPrompt, userPrompt)
	if err != nil {
		return FactsKeywords{}, fmt.Errorf("failed to extract keywords: %w", err)
	}

	var keywords FactsKeywords
	if err := json.Unmarshal([]byte(response), &keywords); err != nil {
		return FactsKeywords{}, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return keywords, nil
}

// generateKeywordsForReport uses LLM to generate Polish keywords for a specific report
func (app *App) generateKeywordsForReport(filename, reportContent string, factsKeywords ProcessedFacts) (string, error) {
	// Build facts context from processed keywords
	factsContext := ""
	if len(factsKeywords) > 0 {
		factsContext = "\n\n=== DOSTĘPNE FAKTY DO CROSS-REFERENCINGU ===\n"
		for factFile, keywords := range factsKeywords {
			factsContext += fmt.Sprintf("\n--- %s ---\n", factFile)
			if len(keywords.People) > 0 {
				factsContext += "OSOBY:\n"
				for _, person := range keywords.People {
					factsContext += fmt.Sprintf("  - %s", person.Name)
					if person.Profession != "" {
						factsContext += fmt.Sprintf(" (%s)", person.Profession)
					}
					if len(person.Skills) > 0 {
						factsContext += fmt.Sprintf(" - umiejętności: %s", strings.Join(person.Skills, ", "))
					}
					if person.Location != "" {
						factsContext += fmt.Sprintf(" - lokalizacja: %s", person.Location)
					}
					if person.Status != "" {
						factsContext += fmt.Sprintf(" - status: %s", person.Status)
					}
					if len(person.Relations) > 0 {
						factsContext += fmt.Sprintf(" - relacje: %s", strings.Join(person.Relations, ", "))
					}
					factsContext += "\n"
				}
			}
			if len(keywords.Sectors) > 0 {
				factsContext += fmt.Sprintf("SEKTORY: %s\n", strings.Join(keywords.Sectors, ", "))
			}
			if len(keywords.Keywords) > 0 {
				factsContext += fmt.Sprintf("SŁOWA KLUCZOWE: %s\n", strings.Join(keywords.Keywords, ", "))
			}
		}
	}

	systemPrompt := fmt.Sprintf(`
	<prompt_objective>
	Jesteś ekspertem w analizie raportów bezpieczeństwa fabryki. Twoim zadaniem jest wygenerowanie obszernego zestawu polskich słów kluczowych dla każdego raportu, które pomogą centralnemu systemowi w kompleksowym wyszukiwaniu tych raportów.
	</prompt_objective>

	<prompt_rules>
	1. Przeanalizuj treść raportu i zidentyfikuj kluczowe informacje:
	   - CO się wydarzyło (rodzaj zdarzenia, incydent, działania)
	   - GDZIE się wydarzyło (lokalizacja, sektor, obszar)
	   - KTO był zaangażowany (osoby, funkcje, zawody, role)
	   - JAKIE obiekty/technologie/narzędzia/języki programowania pojawiły się
	   - KIEDY się wydarzyło (czas, okres, okoliczności)

	2. Wykorzystaj informacje z nazwy pliku (data, numer raportu, sektor)

	3. Połącz informacje z TREŚĆ RAPORTU oraz FAKTY znajdującymi się w <context>, szczególnie osoby wymienione w obu źródłach.
	   Jeśli w raporcie jest nazwisko, sprawdź fakty aby poznać zawód/rolę/technologie tej osoby.

	4. WAŻNE: Generuj OBSZERNY zestaw słów kluczowych w języku polskim:
	   - Używaj przypadku mianownika (np. "nauczyciel", "programista", "laborant")
	   - ZAWSZE tłumacz angielskie terminy na polski
	   - Słowa oddzielaj przecinkami BEZ spacji po przecinkach
	   - ZAWSZE uwzględniaj zawód/rolę osoby jeśli jest wymieniona
	   - ZAWSZE uwzględniaj technologie/umiejętności osoby z faktów (np. "JavaScript", "Python", "Java")
	   - Generuj synonimy i pokrewne terminy (np. "programista,deweloper,inżynier")
	   - Uwzględniaj nazwiska jeśli są istotne
	   - Używaj ogólnych terminów jak "zwierzęta" dla treści o przyrodzie
	   - Używaj ogólnych i szczegółowych terminów
	   - Uwzględniaj różnice w pisowni nazwisk z folderu faktów
       - Generuj minimum 10-15 słów kluczowych dla każdego raportu

	5. Zwróć TYLKO słowa kluczowe oddzielone przecinkami, bez dodatkowych komentarzy

	PRZYKŁAD ODPOWIEDZI: "patrol,sektor-A,alarm,techniczny,naprawa,Joseph,awaria,nauczyciel,edukacja,system,monitoring,kontrola,bezpieczeństwo,incydent,JavaScript,Python,Java,programista,frontend,developer"
	</prompt_rules>

	<context>
	NAZWA PLIKU: %s

	TREŚĆ RAPORTU:
	%s

	FAKTY:
	%s
	</context>`, filename, reportContent, factsContext)

	userPrompt := "Wygeneruj polskie słowa kluczowe dla tego raportu zgodnie z zasadami."

	keywords, err := app.llmClient.AnswerWithContext(systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate keywords: %w", err)
	}

	// Clean up the keywords - remove any extra whitespace and ensure proper formatting
	keywords = strings.TrimSpace(keywords)

	// Remove any quotes that might be added by the LLM
	keywords = strings.Trim(keywords, "\"'")

	return keywords, nil
}
