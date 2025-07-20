package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ai-devs3/internal/config"
	"ai-devs3/pkg/errors"
)

// Cache defines the interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, data []byte) error
}

// FileCache implements Cache interface using file system
type FileCache struct {
	baseDir string
}

// NewFileCache creates a new file-based cache
func NewFileCache(cfg config.CacheConfig) (*FileCache, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(cfg.BaseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &FileCache{
		baseDir: cfg.BaseDir,
	}, nil
}

// Get retrieves data from cache
func (f *FileCache) Get(ctx context.Context, key string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	filename := f.keyToFilename(key)
	filePath := filepath.Join(f.baseDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.NewProcessingError("cache", key, "cache miss", err)
		}
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	return data, nil
}

// Set stores data in cache
func (f *FileCache) Set(ctx context.Context, key string, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	filename := f.keyToFilename(key)
	filePath := filepath.Join(f.baseDir, filename)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache subdirectory: %w", err)
	}

	// Write to temporary file first, then rename for atomicity
	tempFile := filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // Clean up on error
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// keyToFilename converts cache key to safe filename
func (f *FileCache) keyToFilename(key string) string {
	// Replace path separators and other unsafe characters
	safe := strings.ReplaceAll(key, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	safe = strings.ReplaceAll(safe, "?", "_")
	safe = strings.ReplaceAll(safe, "*", "_")
	safe = strings.ReplaceAll(safe, "<", "_")
	safe = strings.ReplaceAll(safe, ">", "_")
	safe = strings.ReplaceAll(safe, "|", "_")
	safe = strings.ReplaceAll(safe, "\"", "_")

	// If the key is very long, hash it
	if len(safe) > 200 {
		hash := sha256.Sum256([]byte(key))
		return fmt.Sprintf("%x", hash)
	}

	return safe
}

// CacheKey generates a consistent cache key from components
func CacheKey(components ...string) string {
	return strings.Join(components, "_")
}
