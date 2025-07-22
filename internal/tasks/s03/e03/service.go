package e03

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"ai-devs3/internal/http"
	"ai-devs3/internal/llm/openai"
	"ai-devs3/pkg/errors"
)

// Service handles the database processing task
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

// ExecuteTask executes the complete S03E03 database task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	startTime := time.Now()

	log.Println("Starting database discovery process...")

	// Step 1: Discover database structure
	dbInfo, err := s.discoverDatabaseStructure(ctx, apiKey)
	if err != nil {
		return nil, errors.NewTaskError("s03e03", "discover_structure", err)
	}

	// Step 2: Generate SQL query using LLM
	sqlQuery, err := s.generateSQLQuery(ctx, dbInfo)
	if err != nil {
		return nil, errors.NewTaskError("s03e03", "generate_query", err)
	}

	log.Printf("Generated SQL query: %s", sqlQuery)

	// Step 3: Execute the query
	datacenterIDs, err := s.executeQueryAndExtractIDs(ctx, apiKey, sqlQuery)
	if err != nil {
		return nil, errors.NewTaskError("s03e03", "execute_query", err)
	}

	log.Printf("Found %d active datacenters with inactive managers", len(datacenterIDs))

	// Step 4: Submit the answer
	response, err := s.submitDatabaseResponse(ctx, apiKey, datacenterIDs)
	if err != nil {
		return nil, errors.NewTaskError("s03e03", "submit_response", err)
	}

	processingTime := time.Since(startTime).Seconds()

	return &TaskResult{
		Response:       response,
		DatacenterIDs:  datacenterIDs,
		GeneratedQuery: sqlQuery,
		ProcessingTime: processingTime,
	}, nil
}

// discoverDatabaseStructure discovers tables and their schemas
func (s *Service) discoverDatabaseStructure(ctx context.Context, apiKey string) (*DatabaseInfo, error) {
	log.Println("Discovering database tables...")

	// Get table list
	tables, err := s.executeQuery(ctx, apiKey, "SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to get table list: %w", err)
	}

	dbInfo := &DatabaseInfo{
		Tables:  make([]string, 0),
		Schemas: make(map[string]*TableSchema),
	}

	// Extract table names
	for _, row := range tables.Reply {
		for _, value := range row {
			if tableName, ok := value.(string); ok {
				dbInfo.Tables = append(dbInfo.Tables, tableName)
				log.Printf("Found table: %s", tableName)
			}
		}
	}

	// Get schema for each table
	for _, tableName := range dbInfo.Tables {
		log.Printf("Getting schema for table: %s", tableName)

		schemaQuery := fmt.Sprintf("SHOW CREATE TABLE %s", tableName)
		schemaResult, err := s.executeQuery(ctx, apiKey, schemaQuery)
		if err != nil {
			log.Printf("Warning: Failed to get schema for table %s: %v", tableName, err)
			continue
		}

		// Find and print side flag in correct_order table
		if tableName == "correct_order" {
			log.Printf("Getting content for table: %s", tableName)

			contentQuery := fmt.Sprintf("SELECT * FROM %s order by weight", tableName)
			contentResult, err := s.executeQuery(ctx, apiKey, contentQuery)
			if err != nil {
				log.Printf("Warning: Failed to get content for table %s: %v", tableName, err)
				continue
			}

			var sideFlag string
			for _, row := range contentResult.Reply {
				sideFlag += fmt.Sprintf("%v", row["letter"])
			}

			log.Printf("Found side flag: %s", sideFlag)
		}

		schema := &TableSchema{
			Name:    tableName,
			Columns: make([]string, 0),
		}

		// Extract CREATE statement
		if len(schemaResult.Reply) > 0 {
			for key, value := range schemaResult.Reply[0] {
				if strings.Contains(strings.ToLower(key), "create") {
					if createSQL, ok := value.(string); ok {
						schema.CreateSQL = createSQL
						// log.Printf("Schema for %s: %s", tableName, createSQL)
					}
				}
			}
		}

		dbInfo.Schemas[tableName] = schema
	}

	return dbInfo, nil
}

// generateSQLQuery uses LLM to generate the appropriate SQL query
func (s *Service) generateSQLQuery(ctx context.Context, dbInfo *DatabaseInfo) (string, error) {
	log.Println("Generating SQL query using LLM...")

	// Build prompt with database schema information
	prompt := s.buildQueryGenerationPrompt(dbInfo)

	// Call LLM to generate query
	systemPrompt := "You are a SQL expert. Generate only the SQL query text without any markdown formatting, explanations, or additional text. Return ONLY the raw SQL query."

	response, err := s.llmClient.GetAnswerWithContext(ctx, systemPrompt, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate SQL query: %w", err)
	}

	// Extract clean SQL query
	sqlQuery := strings.TrimSpace(response)
	sqlQuery = strings.Trim(sqlQuery, "`")
	sqlQuery = strings.TrimPrefix(sqlQuery, "sql")
	sqlQuery = strings.TrimSpace(sqlQuery)

	return sqlQuery, nil
}

// buildQueryGenerationPrompt creates a prompt for SQL generation
func (s *Service) buildQueryGenerationPrompt(dbInfo *DatabaseInfo) string {
	var prompt strings.Builder

	prompt.WriteString("Database Schema Information:\n\n")

	for tableName, schema := range dbInfo.Schemas {
		prompt.WriteString(fmt.Sprintf("Table: %s\n", tableName))
		if schema.CreateSQL != "" {
			prompt.WriteString(fmt.Sprintf("Schema: %s\n\n", schema.CreateSQL))
		}
	}

	prompt.WriteString("Task: Generate a SQL query to find active datacenters that are managed by managers who are currently inactive (on vacation).\n\n")
	prompt.WriteString("Requirements:\n")
	prompt.WriteString("- Find datacenters that are active\n")
	prompt.WriteString("- Their managers must be inactive/on vacation\n")
	prompt.WriteString("- Return the datacenter IDs\n")
	prompt.WriteString("- Look for status columns, active flags, or similar indicators\n")
	prompt.WriteString("- Pay attention to relationships between tables (likely manager_id references)\n\n")
	prompt.WriteString("Return ONLY the SQL query, no explanations or formatting:")

	return prompt.String()
}

// executeQueryAndExtractIDs executes the query and extracts datacenter IDs
func (s *Service) executeQueryAndExtractIDs(ctx context.Context, apiKey, sqlQuery string) ([]int, error) {
	log.Printf("Executing query: %s", sqlQuery)

	result, err := s.executeQuery(ctx, apiKey, sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	var datacenterIDs []int

	// Extract IDs from result
	for _, row := range result.Reply {
		for key, value := range row {
			// Look for ID fields (id, datacenter_id, dc_id, etc.)
			if strings.Contains(strings.ToLower(key), "id") {
				if id, err := s.convertToInt(value); err == nil {
					datacenterIDs = append(datacenterIDs, id)
					log.Printf("Found datacenter ID: %d", id)
				}
			}
		}
	}

	log.Printf("Total datacenter IDs found: %d", len(datacenterIDs))
	return datacenterIDs, nil
}

// executeQuery executes a SQL query against the database API
func (s *Service) executeQuery(ctx context.Context, apiKey, query string) (*DatabaseResponse, error) {
	request := &DatabaseRequest{
		Task:   "database",
		APIKey: apiKey,
		Query:  query,
	}

	responseBody, err := s.httpClient.PostJSON(ctx, "https://c3ntrala.ag3nts.org/apidb", request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute database query: %w", err)
	}

	var response DatabaseResponse
	if err := json.Unmarshal([]byte(responseBody), &response); err != nil {
		return nil, fmt.Errorf("failed to parse database response: %w", err)
	}

	if response.Error != "" && response.Error != "OK" {
		return nil, fmt.Errorf("database error: %s", response.Error)
	}

	return &response, nil
}

// convertToInt converts various types to int
func (s *Service) convertToInt(value any) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, fmt.Errorf("cannot convert %T to int", v)
	}
}

// submitDatabaseResponse submits the database results to the centrala API
func (s *Service) submitDatabaseResponse(ctx context.Context, apiKey string, datacenterIDs []int) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("database", apiKey, datacenterIDs)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit database response: %w", err)
	}

	return result, nil
}
