package e05

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"ai-devs3/internal/http"
	"ai-devs3/internal/neo4j"
	"ai-devs3/pkg/errors"
)

// Service handles the connections processing task
type Service struct {
	httpClient  *http.Client
	neo4jClient *neo4j.Client
}

// NewService creates a new service instance
func NewService(httpClient *http.Client) *Service {
	return &Service{
		httpClient: httpClient,
	}
}

// SetNeo4jClient sets the Neo4j client (called after Neo4j connection is established)
func (s *Service) SetNeo4jClient(client *neo4j.Client) {
	s.neo4jClient = client
}

// ExecuteTask executes the complete S03E05 connections task workflow
func (s *Service) ExecuteTask(ctx context.Context, apiKey string) (*TaskResult, error) {
	startTime := time.Now()

	log.Println("Starting connections data retrieval and graph processing...")

	// Step 1: Retrieve users and connections from MySQL
	graphData, err := s.retrieveGraphData(ctx, apiKey)
	if err != nil {
		return nil, errors.NewTaskError("s03e05", "retrieve_graph_data", err)
	}

	log.Printf("Retrieved %d users and %d connections from MySQL", len(graphData.Users), len(graphData.Connections))

	// Step 2: Clear and populate Neo4j graph
	stats, err := s.populateNeo4jGraph(ctx, graphData)
	if err != nil {
		return nil, errors.NewTaskError("s03e05", "populate_neo4j", err)
	}

	// Step 3: Find shortest path between Rafał and Barbara
	shortestPath, err := s.findShortestPath(ctx, "Rafał", "Barbara")
	if err != nil {
		return nil, errors.NewTaskError("s03e05", "find_shortest_path", err)
	}

	log.Printf("Found shortest path with %d nodes: %v", len(shortestPath), shortestPath)

	// Step 4: Format path as comma-separated string
	pathString := strings.Join(shortestPath, ",")

	// Step 5: Submit the answer
	response, err := s.submitConnectionsResponse(ctx, apiKey, pathString)
	if err != nil {
		return nil, errors.NewTaskError("s03e05", "submit_response", err)
	}

	processingTime := time.Since(startTime).Seconds()

	// Update stats
	stats.UsersLoaded = len(graphData.Users)
	stats.ConnectionsLoaded = len(graphData.Connections)
	stats.PathFound = len(shortestPath) > 0
	stats.ProcessingTime = processingTime

	return &TaskResult{
		Response:       response,
		ShortestPath:   shortestPath,
		PathString:     pathString,
		ProcessingTime: processingTime,
		Stats:          *stats,
	}, nil
}

// retrieveGraphData retrieves users and connections from MySQL database
func (s *Service) retrieveGraphData(ctx context.Context, apiKey string) (*GraphData, error) {
	log.Println("Retrieving users from MySQL database...")

	// Get all users
	usersResult, err := s.executeQuery(ctx, apiKey, "SELECT id, username FROM users")
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Parse users
	var users []User
	for _, row := range usersResult.Reply {
		user := User{}

		if id, ok := row["id"]; ok {
			if userID, err := s.convertToInt(id); err == nil {
				user.ID = userID
			}
		}

		if username, ok := row["username"]; ok {
			if name, ok := username.(string); ok {
				user.Username = name
			}
		}

		if user.ID != 0 && user.Username != "" {
			users = append(users, user)
		}
	}

	log.Printf("Retrieved %d users from database", len(users))

	// Get all connections
	log.Println("Retrieving connections from MySQL database...")
	connectionsResult, err := s.executeQuery(ctx, apiKey, "SELECT user1_id, user2_id FROM connections")
	if err != nil {
		return nil, fmt.Errorf("failed to get connections: %w", err)
	}

	// Parse connections
	var connections []Connection
	for _, row := range connectionsResult.Reply {
		connection := Connection{}

		if user1ID, ok := row["user1_id"]; ok {
			if id, err := s.convertToInt(user1ID); err == nil {
				connection.User1ID = id
			}
		}

		if user2ID, ok := row["user2_id"]; ok {
			if id, err := s.convertToInt(user2ID); err == nil {
				connection.User2ID = id
			}
		}

		if connection.User1ID != 0 && connection.User2ID != 0 {
			connections = append(connections, connection)
		}
	}

	log.Printf("Retrieved %d connections from database", len(connections))

	return &GraphData{
		Users:       users,
		Connections: connections,
	}, nil
}

// populateNeo4jGraph clears and populates the Neo4j graph with users and connections
func (s *Service) populateNeo4jGraph(ctx context.Context, graphData *GraphData) (*ProcessingStats, error) {
	if s.neo4jClient == nil {
		return nil, fmt.Errorf("Neo4j client not initialized")
	}

	stats := &ProcessingStats{}

	log.Println("Clearing Neo4j database...")
	if err := s.neo4jClient.ClearDatabase(ctx); err != nil {
		return nil, fmt.Errorf("failed to clear Neo4j database: %w", err)
	}

	log.Println("Creating user nodes in Neo4j...")

	// Create user nodes
	for _, user := range graphData.Users {
		if err := s.neo4jClient.CreateUser(ctx, user.ID, user.Username); err != nil {
			return nil, fmt.Errorf("failed to create user %s: %w", user.Username, err)
		}
		stats.NodesCreated++
	}

	log.Printf("Created %d user nodes", stats.NodesCreated)

	log.Println("Creating connection relationships in Neo4j...")

	// Create connections (relationships)
	for _, connection := range graphData.Connections {
		if err := s.neo4jClient.CreateConnection(ctx, connection.User1ID, connection.User2ID); err != nil {
			return nil, fmt.Errorf("failed to create connection %d -> %d: %w", connection.User1ID, connection.User2ID, err)
		}
		stats.RelationshipsCreated++
	}

	log.Printf("Created %d relationship edges", stats.RelationshipsCreated)

	// Verify the data was loaded correctly
	nodeCount, err := s.neo4jClient.GetNodeCount(ctx)
	if err != nil {
		log.Printf("Warning: could not verify node count: %v", err)
	} else {
		log.Printf("Neo4j verification: %d nodes in database", nodeCount)
	}

	relationshipCount, err := s.neo4jClient.GetRelationshipCount(ctx)
	if err != nil {
		log.Printf("Warning: could not verify relationship count: %v", err)
	} else {
		log.Printf("Neo4j verification: %d relationships in database", relationshipCount)
	}

	return stats, nil
}

// findShortestPath finds the shortest path between two users using Neo4j
func (s *Service) findShortestPath(ctx context.Context, startUsername, endUsername string) ([]string, error) {
	if s.neo4jClient == nil {
		return nil, fmt.Errorf("Neo4j client not initialized")
	}

	log.Printf("Finding shortest path from %s to %s...", startUsername, endUsername)

	path, err := s.neo4jClient.FindShortestPath(ctx, startUsername, endUsername)
	if err != nil {
		return nil, fmt.Errorf("failed to find shortest path: %w", err)
	}

	if len(path) == 0 {
		return nil, fmt.Errorf("no path found between %s and %s", startUsername, endUsername)
	}

	log.Printf("Shortest path found: %v (length: %d steps)", path, len(path)-1)
	return path, nil
}

// executeQuery executes a SQL query against the database API (reused from S03E03)
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

// convertToInt converts various types to int (reused from S03E03)
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

// submitConnectionsResponse submits the connections result to the centrala API
func (s *Service) submitConnectionsResponse(ctx context.Context, apiKey, pathString string) (string, error) {
	response := s.httpClient.BuildAIDevsResponse("connections", apiKey, pathString)

	result, err := s.httpClient.PostReport(ctx, "https://c3ntrala.ag3nts.org", response)
	if err != nil {
		return "", fmt.Errorf("failed to submit connections response: %w", err)
	}

	return result, nil
}
