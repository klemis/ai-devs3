package cache

import (
	"context"
)

// TaskCache provides task-specific cache operations
type TaskCache struct {
	cache   Cache
	taskID  string
	baseDir string
}

// NewTaskCache creates a new task-specific cache
func NewTaskCache(cache Cache, taskID string) *TaskCache {
	return &TaskCache{
		cache:   cache,
		taskID:  taskID,
		baseDir: taskID,
	}
}

// GetAudioTranscript retrieves cached audio transcript
func (t *TaskCache) GetAudioTranscript(ctx context.Context, audioKey string) (string, error) {
	key := CacheKey(t.taskID, "audio", audioKey, "transcript")
	data, err := t.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetAudioTranscript stores audio transcript in cache
func (t *TaskCache) SetAudioTranscript(ctx context.Context, audioKey, transcript string) error {
	key := CacheKey(t.taskID, "audio", audioKey, "transcript")
	return t.cache.Set(ctx, key, []byte(transcript))
}

// GetOCRText retrieves cached OCR text
func (t *TaskCache) GetOCRText(ctx context.Context, imageKey string) (string, error) {
	key := CacheKey(t.taskID, "image", imageKey, "ocr")
	data, err := t.cache.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetOCRText stores OCR text in cache
func (t *TaskCache) SetOCRText(ctx context.Context, imageKey, text string) error {
	key := CacheKey(t.taskID, "image", imageKey, "ocr")
	return t.cache.Set(ctx, key, []byte(text))
}
