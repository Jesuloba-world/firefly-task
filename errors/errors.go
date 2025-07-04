package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeAWS           ErrorType = "AWS_API"
	ErrorTypeTerraform     ErrorType = "TERRAFORM"
	ErrorTypeConfiguration ErrorType = "CONFIGURATION"
	ErrorTypeRuntime       ErrorType = "RUNTIME"
	ErrorTypeValidation    ErrorType = "VALIDATION"
	ErrorTypeFileSystem    ErrorType = "FILESYSTEM"
)

// FireflyError represents a custom error with additional context
type FireflyError struct {
	Type        ErrorType `json:"type"`
	Message     string    `json:"message"`
	Details     string    `json:"details,omitempty"`
	Cause       error     `json:"cause,omitempty"`
	Operation   string    `json:"operation,omitempty"`
	ResourceID  string    `json:"resource_id,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// Error implements the error interface
func (e *FireflyError) Error() string {
	var parts []string

	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Operation))
	}

	parts = append(parts, fmt.Sprintf("%s: %s", e.Type, e.Message))

	if e.Details != "" {
		parts = append(parts, fmt.Sprintf("Details: %s", e.Details))
	}

	if e.ResourceID != "" {
		parts = append(parts, fmt.Sprintf("Resource: %s", e.ResourceID))
	}

	if e.Cause != nil {
		parts = append(parts, fmt.Sprintf("Cause: %v", e.Cause))
	}

	return strings.Join(parts, " | ")
}

// Unwrap returns the underlying cause error
func (e *FireflyError) Unwrap() error {
	return e.Cause
}

// WithSuggestions adds suggestions to the error
func (e *FireflyError) WithSuggestions(suggestions ...string) *FireflyError {
	e.Suggestions = append(e.Suggestions, suggestions...)
	return e
}

// WithDetails adds additional details to the error
func (e *FireflyError) WithDetails(details string) *FireflyError {
	e.Details = details
	return e
}

// WithCause adds a cause error
func (e *FireflyError) WithCause(cause error) *FireflyError {
	e.Cause = cause
	return e
}

// WithResource adds a resource ID to the error
func (e *FireflyError) WithResource(resourceID string) *FireflyError {
	e.ResourceID = resourceID
	return e
}

// NewAWSError creates a new AWS API error
func NewAWSError(operation, message string) *FireflyError {
	return &FireflyError{
		Type:      ErrorTypeAWS,
		Message:   message,
		Operation: operation,
	}
}

// NewTerraformError creates a new Terraform parsing error
func NewTerraformError(operation, message string) *FireflyError {
	return &FireflyError{
		Type:      ErrorTypeTerraform,
		Message:   message,
		Operation: operation,
	}
}

// NewConfigurationError creates a new configuration error
func NewConfigurationError(message string) *FireflyError {
	return &FireflyError{
		Type:    ErrorTypeConfiguration,
		Message: message,
	}
}

// NewRuntimeError creates a new runtime error
func NewRuntimeError(operation, message string) *FireflyError {
	return &FireflyError{
		Type:      ErrorTypeRuntime,
		Message:   message,
		Operation: operation,
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *FireflyError {
	return &FireflyError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// NewFileSystemError creates a new filesystem error
func NewFileSystemError(operation, message string) *FireflyError {
	return &FireflyError{
		Type:      ErrorTypeFileSystem,
		Message:   message,
		Operation: operation,
	}
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	if fireflyErr, ok := err.(*FireflyError); ok {
		return fireflyErr.Type == errorType
	}
	return false
}

// GetSuggestions returns suggestions from a FireflyError, if any
func GetSuggestions(err error) []string {
	if fireflyErr, ok := err.(*FireflyError); ok {
		return fireflyErr.Suggestions
	}
	return nil
}

// WrapError wraps a standard error into a FireflyError
func WrapError(err error, errorType ErrorType, operation, message string) *FireflyError {
	return &FireflyError{
		Type:      errorType,
		Message:   message,
		Operation: operation,
		Cause:     err,
	}
}
