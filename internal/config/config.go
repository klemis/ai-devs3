package config

import (
	"os"
	"time"

	pkgerrors "ai-devs3/pkg/errors"
)

// Config holds all configuration for the application
type Config struct {
	AIDevs AIDevsConfig
	OpenAI OpenAIConfig
	Ollama OllamaConfig
	HTTP   HTTPConfig
	Cache  CacheConfig
	Qdrant QdrantConfig
	Neo4j  Neo4jConfig
}

// AIDevsConfig holds AI-DEVS specific configuration
type AIDevsConfig struct {
	APIKey  string
	BaseURL string
}

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey      string
	Model       string
	Temperature float64
}

// HTTPConfig holds HTTP client configuration
type HTTPConfig struct {
	Timeout time.Duration
	Retries int
}

// OllamaConfig holds Ollama local LLM configuration
type OllamaConfig struct {
	BaseURL     string
	Model       string
	Temperature float64
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	BaseDir string
}

// QdrantConfig holds Qdrant vector database configuration
type QdrantConfig struct {
	Host   string
	Port   int
	APIKey string
	UseTLS bool
}

// Neo4jConfig holds Neo4j graph database configuration
type Neo4jConfig struct {
	URI      string
	Username string
	Password string
}

// Load creates a new Config instance from environment variables
func Load() (*Config, error) {
	config := &Config{
		AIDevs: AIDevsConfig{
			APIKey:  getEnv("AI_DEVS_API_KEY", ""),
			BaseURL: getEnv("AI_DEVS_BASE_URL", "https://c3ntrala.ag3nts.org"),
		},
		OpenAI: OpenAIConfig{
			APIKey:      getEnv("OPENAI_API_KEY", ""),
			Model:       getEnv("OPENAI_MODEL", "gpt-4o-mini"),
			Temperature: 0.3,
		},
		Ollama: OllamaConfig{
			BaseURL:     getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			Model:       getEnv("OLLAMA_MODEL", "llama3.2:3b"),
			Temperature: 0.5,
		},
		HTTP: HTTPConfig{
			Timeout: 30 * time.Second,
			Retries: 3,
		},
		Cache: CacheConfig{
			BaseDir: getEnv("CACHE_DIR", "data"),
		},
		Qdrant: QdrantConfig{
			Host:   getEnv("QDRANT_HOST", "localhost"),
			Port:   6334, // grpc port
			APIKey: getEnv("QDRANT_API_KEY", ""),
			UseTLS: true,
		},
		Neo4j: Neo4jConfig{
			URI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
			Username: getEnv("NEO4J_USER", "neo4j"),
			Password: getEnv("NEO4J_PASSWORD", ""),
		},
	}

	// Validate required fields
	if config.AIDevs.APIKey == "" {
		return nil, pkgerrors.NewConfigError("AI_DEVS_API_KEY", "environment variable is required", nil)
	}

	if config.OpenAI.APIKey == "" {
		return nil, pkgerrors.NewConfigError("OPENAI_API_KEY", "environment variable is required", nil)
	}

	if config.Qdrant.APIKey == "" {
		return nil, pkgerrors.NewConfigError("QDRANT_API_KEY", "environment variable is required", nil)
	}

	// Note: Neo4j password is validated when creating the Neo4j client, not here
	// This allows for optional Neo4j usage depending on the task

	return config, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
