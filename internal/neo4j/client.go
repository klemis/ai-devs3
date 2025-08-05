package neo4j

import (
	"context"
	"fmt"
	"log"

	"ai-devs3/internal/config"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Client wraps the Neo4j driver with application-specific methods
type Client struct {
	driver neo4j.DriverWithContext
	config *config.Neo4jConfig
}

// NewClient creates a new Neo4j client instance
func NewClient(cfg *config.Neo4jConfig) (*Client, error) {
	if cfg.Password == "" {
		return nil, fmt.Errorf("NEO4J_PASSWORD is required")
	}

	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		neo4j.BasicAuth(cfg.Username, cfg.Password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	return &Client{
		driver: driver,
		config: cfg,
	}, nil
}

// Close closes the Neo4j driver connection
func (c *Client) Close(ctx context.Context) error {
	return c.driver.Close(ctx)
}

// VerifyConnectivity verifies the connection to Neo4j
func (c *Client) VerifyConnectivity(ctx context.Context) error {
	return c.driver.VerifyConnectivity(ctx)
}

// ClearDatabase clears all nodes and relationships from the database
func (c *Client) ClearDatabase(ctx context.Context) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
	if err != nil {
		return fmt.Errorf("failed to clear database: %w", err)
	}

	log.Println("Neo4j database cleared")
	return nil
}

// CreateUser creates a Person node in Neo4j
func (c *Client) CreateUser(ctx context.Context, userID int, username string) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		CREATE (u:Person {userId: $userId, username: $username})
	`

	params := map[string]any{
		"userId":   userID,
		"username": username,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", username, err)
	}

	return nil
}

// CreateConnection creates a KNOWS relationship between two users
func (c *Client) CreateConnection(ctx context.Context, user1ID, user2ID int) error {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH (u1:Person {userId: $user1Id})
		MATCH (u2:Person {userId: $user2Id})
		CREATE (u1)-[:KNOWS]->(u2)
	`

	params := map[string]any{
		"user1Id": user1ID,
		"user2Id": user2ID,
	}

	_, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("failed to create connection between %d and %d: %w", user1ID, user2ID, err)
	}

	return nil
}

// FindShortestPath finds the shortest path between two users
func (c *Client) FindShortestPath(ctx context.Context, startUsername, endUsername string) ([]string, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `
		MATCH path = shortestPath((start:Person {username: $startUsername})-[:KNOWS*]-(end:Person {username: $endUsername}))
		RETURN [node in nodes(path) | node.username] as path
	`

	params := map[string]any{
		"startUsername": startUsername,
		"endUsername":   endUsername,
	}

	result, err := session.Run(ctx, query, params)
	if err != nil {
		return nil, fmt.Errorf("failed to find shortest path: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		pathValue, found := record.Get("path")
		if !found {
			return nil, fmt.Errorf("no path found in result")
		}

		pathInterface, ok := pathValue.([]any)
		if !ok {
			return nil, fmt.Errorf("path is not a slice")
		}

		path := make([]string, len(pathInterface))
		for i, v := range pathInterface {
			username, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("path element is not a string")
			}
			path[i] = username
		}

		return path, nil
	}

	if err = result.Err(); err != nil {
		return nil, fmt.Errorf("error reading result: %w", err)
	}

	return nil, fmt.Errorf("no path found between %s and %s", startUsername, endUsername)
}

// GetNodeCount returns the total number of nodes in the database
func (c *Client) GetNodeCount(ctx context.Context) (int, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx, "MATCH (n) RETURN count(n) as count", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get node count: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		countValue, found := record.Get("count")
		if !found {
			return 0, fmt.Errorf("count not found in result")
		}

		count, ok := countValue.(int64)
		if !ok {
			return 0, fmt.Errorf("count is not an integer")
		}

		return int(count), nil
	}

	return 0, fmt.Errorf("no result returned")
}

// GetRelationshipCount returns the total number of relationships in the database
func (c *Client) GetRelationshipCount(ctx context.Context) (int, error) {
	session := c.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	result, err := session.Run(ctx, "MATCH ()-[r]->() RETURN count(r) as count", nil)
	if err != nil {
		return 0, fmt.Errorf("failed to get relationship count: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		countValue, found := record.Get("count")
		if !found {
			return 0, fmt.Errorf("count not found in result")
		}

		count, ok := countValue.(int64)
		if !ok {
			return 0, fmt.Errorf("count is not an integer")
		}

		return int(count), nil
	}

	return 0, fmt.Errorf("no result returned")
}
