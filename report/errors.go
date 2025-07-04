package report

import (
	"fmt"
)

// ErrorType represents different types of report generation errors
type ErrorType int

const (
	// ErrorTypeInvalidGenerator indicates an invalid generator type was specified
	ErrorTypeInvalidGenerator ErrorType = iota
	// ErrorTypeInvalidFormat indicates an unsupported output format
	ErrorTypeInvalidFormat
	// ErrorTypeFileWrite indicates an error writing to file
	ErrorTypeFileWrite
	// ErrorTypeMarshaling indicates an error during JSON/YAML marshaling
	ErrorTypeMarshaling
	// ErrorTypeInvalidInput indicates invalid input data
	ErrorTypeInvalidInput
	// ErrorTypeTemplateError indicates an error in report template processing
	ErrorTypeTemplateError
	// ErrorTypeFilterError indicates an error in result filtering
	ErrorTypeFilterError
	// ErrorTypeRecommendationError indicates an error generating recommendations
	ErrorTypeRecommendationError
	// ErrorTypeUnsupportedFormat indicates an unsupported format was requested
	ErrorTypeUnsupportedFormat
	// ErrorTypeFileOperation indicates a file operation error
	ErrorTypeFileOperation
	// ErrorTypeGeneration indicates a report generation error
	ErrorTypeGeneration
	// ErrorTypeGenerationFailed indicates report generation failed
	ErrorTypeGenerationFailed
	// ErrorTypeConfiguration indicates a configuration error
	ErrorTypeConfiguration
	// ErrorTypeNotImplemented indicates a feature is not implemented
	ErrorTypeNotImplemented
)

// String returns the string representation of ErrorType
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeInvalidGenerator:
		return "invalid_generator"
	case ErrorTypeInvalidFormat:
		return "invalid_format"
	case ErrorTypeFileWrite:
		return "file_write"
	case ErrorTypeMarshaling:
		return "marshaling"
	case ErrorTypeInvalidInput:
		return "invalid_input"
	case ErrorTypeTemplateError:
		return "template_error"
	case ErrorTypeFilterError:
		return "filter_error"
	case ErrorTypeRecommendationError:
		return "recommendation_error"
	case ErrorTypeUnsupportedFormat:
		return "unsupported_format"
	case ErrorTypeFileOperation:
		return "file_operation"
	case ErrorTypeGeneration:
		return "generation"
	case ErrorTypeGenerationFailed:
		return "generation_failed"
	case ErrorTypeConfiguration:
		return "configuration"
	case ErrorTypeNotImplemented:
		return "not_implemented"
	default:
		return "unknown"
	}
}

// ReportError represents an error that occurred during report generation
type ReportError struct {
	// Type indicates the type of error
	Type ErrorType
	// Message provides a human-readable error message
	Message string
	// Cause contains the underlying error that caused this error
	Cause error
	// Context provides additional context about where the error occurred
	Context map[string]interface{}
}

// Error implements the error interface
func (re *ReportError) Error() string {
	if re.Cause != nil {
		return fmt.Sprintf("report error [%s]: %s (caused by: %v)", re.Type.String(), re.Message, re.Cause)
	}
	return fmt.Sprintf("report error [%s]: %s", re.Type.String(), re.Message)
}

// Unwrap returns the underlying error
func (re *ReportError) Unwrap() error {
	return re.Cause
}

// WithContext adds context information to the error
func (re *ReportError) WithContext(key string, value interface{}) *ReportError {
	if re.Context == nil {
		re.Context = make(map[string]interface{})
	}
	re.Context[key] = value
	return re
}

// NewReportError creates a new ReportError
func NewReportError(errorType ErrorType, message string) *ReportError {
	return &ReportError{
		Type:    errorType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewReportErrorWithCause creates a new ReportError with an underlying cause
func NewReportErrorWithCause(errorType ErrorType, message string, cause error) *ReportError {
	return &ReportError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WrapError wraps an existing error as a ReportError
func WrapError(errorType ErrorType, message string, err error) *ReportError {
	return NewReportErrorWithCause(errorType, message, err)
}

// WrapReportError wraps an existing error as a ReportError
func WrapReportError(errorType ErrorType, message string, err error) *ReportError {
	return NewReportErrorWithCause(errorType, message, err)
}

// NewReportErrorf creates a new ReportError with formatted message
func NewReportErrorf(errorType ErrorType, format string, args ...interface{}) *ReportError {
	return &ReportError{
		Type:    errorType,
		Message: fmt.Sprintf(format, args...),
		Context: make(map[string]interface{}),
	}
}

// IsReportError checks if an error is a ReportError of a specific type
func IsReportError(err error, errorType ErrorType) bool {
	if reportErr, ok := err.(*ReportError); ok {
		return reportErr.Type == errorType
	}
	return false
}

// GetReportErrorType returns the ErrorType of a ReportError, or -1 if not a ReportError
func GetReportErrorType(err error) ErrorType {
	if reportErr, ok := err.(*ReportError); ok {
		return reportErr.Type
	}
	return -1
}