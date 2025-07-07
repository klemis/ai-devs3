package app

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// DocumentsAnswer represents the answer structure for the documents task
type DocumentsAnswer map[string]string

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

	// Step 2: Read facts folder for cross-referencing
	factsContent, err := app.readFactsFolder(factsDir)
	if err != nil {
		return "", fmt.Errorf("failed to read facts folder: %w", err)
	}

	log.Printf("Loaded facts from %d files", len(factsContent))

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
		keywords, err := app.generateKeywordsForReport(txtFile, string(reportContent), factsContent)
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

// readFactsFolder reads all files from the facts folder for cross-referencing
func (app *App) readFactsFolder(factsDir string) (map[string]string, error) {
	factsContent := make(map[string]string)

	// Check if facts directory exists
	if _, err := os.Stat(factsDir); os.IsNotExist(err) {
		log.Printf("Facts directory %s does not exist, continuing without facts", factsDir)
		return factsContent, nil
	}

	files, err := os.ReadDir(factsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read facts directory %s: %w", factsDir, err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(factsDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read facts file %s: %v", file.Name(), err)
			continue
		}

		factsContent[file.Name()] = string(content)
	}

	return factsContent, nil
}

// generateKeywordsForReport uses LLM to generate Polish keywords for a specific report
func (app *App) generateKeywordsForReport(filename, reportContent string, factsContent map[string]string) (string, error) {
	// Build facts context
	factsContext := ""
	if len(factsContent) > 0 {
		factsContext = "\n\n=== DOSTĘPNE FAKTY DO CROSS-REFERENCINGU ===\n"
		for factFile, factContent := range factsContent {
			factsContext += fmt.Sprintf("\n--- %s ---\n%s\n", factFile, factContent)
		}
	}

	systemPrompt := fmt.Sprintf(`
	<prompt_objective>
	Jesteś ekspertem w analizie raportów bezpieczeństwa fabryki. Twoim zadaniem jest wygenerowanie polskich słów kluczowych dla każdego raportu, które pomogą centralnemu systemowi w wyszukiwaniu tych raportów.
	</prompt_objective>

	<prompt_rules>
	1. Przeanalizuj treść raportu i zidentyfikuj kluczowe informacje:
	   - Co się wydarzyło (rodzaj zdarzenia, incydent)
	   - Gdzie się wydarzyło (lokalizacja, sektor)
	   - Kto był zaangażowany (osoby, funkcje, zawody)
	   - Jakie obiekty/technologie pojawiły się
	   - Jeśli w raporcie pojawia się osoba, a w "faktach" znajdują się informacje o tej osobie lub inne istotne szczegóły, muszą one trafić do słów kluczowych dla tego raportu

	2. Wykorzystaj informacje z nazwy pliku (data, numer raportu, sektor)

	3. Skrzyżuj informacje z dostępnymi faktami, szczególnie osoby wymienione w obu źródłach,
	jeśli w raporcie jest nazwisko, sprawdź fakty aby poznać zawód/rolę tej osoby.

	4. Wygeneruj słowa kluczowe w języku polskim:
	   - Używaj przypadku mianownika (np. "nauczyciel", "programista", "laborant")
	   - Słowa oddzielaj przecinkami BEZ spacji po przecinkach
	   - ZAWSZE uwzględniaj zawód/rolę osoby jeśli jest wymieniona
	   - Używaj konkretnych terminów związanych z raportem
	   - Uwzględniaj nazwiska jeśli są istotne
	   - Używaj ogólnych terminów jak "zwierzęta" dla treści o przyrodzie
	   - Uwzględniaj różnice w pisowni nazwisk z folderu faktów
		- Słowa kluczowe mają być jak najbardziej specyficzne dla danego raportu i powiązanych faktów.

	5. Zwróć TYLKO słowa kluczowe oddzielone przecinkami, bez dodatkowych komentarzy

	PRZYKŁAD ODPOWIEDZI: "patrol,sektor-A,alarm,techniczny,naprawa,Joseph,awaria,nauczyciel"
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
