package errors

import (
	"fmt"
)

// TaskError represents an error that occurred during task execution
type TaskError struct {
	TaskID string
	Step   string
	Err    error
}

func (e TaskError) Error() string {
	return fmt.Sprintf("task %s failed at step %s: %v", e.TaskID, e.Step, e.Err)
}

func (e TaskError) Unwrap() error {
	return e.Err
}

// NewTaskError creates a new TaskError
func NewTaskError(taskID, step string, err error) TaskError {
	return TaskError{
		TaskID: taskID,
		Step:   step,
		Err:    err,
	}
}

// APIError represents an error from external API calls
type APIError struct {
	Service    string
	StatusCode int
	Message    string
	Err        error
}

func (e APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("%s API error (status %d): %s", e.Service, e.StatusCode, e.Message)
	}
	return fmt.Sprintf("%s API error: %s", e.Service, e.Message)
}

func (e APIError) Unwrap() error {
	return e.Err
}

// NewAPIError creates a new APIError
func NewAPIError(service string, statusCode int, message string, err error) APIError {
	return APIError{
		Service:    service,
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// ConfigError represents a configuration error
type ConfigError struct {
	Field   string
	Message string
	Err     error
}

func (e ConfigError) Error() string {
	return fmt.Sprintf("configuration error for field %s: %s", e.Field, e.Message)
}

func (e ConfigError) Unwrap() error {
	return e.Err
}

// NewConfigError creates a new ConfigError
func NewConfigError(field, message string, err error) ConfigError {
	return ConfigError{
		Field:   field,
		Message: message,
		Err:     err,
	}
}

// ProcessingError represents an error during data processing
type ProcessingError struct {
	Type    string // "image", "audio", "text"
	Source  string // file path or URL
	Message string
	Err     error
}

func (e ProcessingError) Error() string {
	return fmt.Sprintf("%s processing error for %s: %s", e.Type, e.Source, e.Message)
}

func (e ProcessingError) Unwrap() error {
	return e.Err
}

// NewProcessingError creates a new ProcessingError
func NewProcessingError(processType, source, message string, err error) ProcessingError {
	return ProcessingError{
		Type:    processType,
		Source:  source,
		Message: message,
		Err:     err,
	}
}
