package report

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		expected  string
	}{
		{
			name:      "Invalid input error",
			errorType: ErrorTypeInvalidInput,
			expected:  "invalid_input",
		},
		{
			name:      "Unsupported format error",
			errorType: ErrorTypeUnsupportedFormat,
			expected:  "unsupported_format",
		},
		{
			name:      "File operation error",
			errorType: ErrorTypeFileOperation,
			expected:  "file_operation",
		},
		{
			name:      "Generation error",
			errorType: ErrorTypeGeneration,
			expected:  "generation",
		},
		{
			name:      "Configuration error",
			errorType: ErrorTypeConfiguration,
			expected:  "configuration",
		},
		{
			name:      "Unknown error type",
			errorType: ErrorType(999),
			expected:  "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errorType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReportError_Error(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		message   string
		expected  string
	}{
		{
			name:      "Simple error message",
			errorType: ErrorTypeInvalidInput,
			message:   "test message",
			expected:  "report error [invalid_input]: test message",
		},
		{
			name:      "Empty message",
			errorType: ErrorTypeGeneration,
			message:   "",
			expected:  "report error [generation]: ",
		},
		{
			name:      "Long message",
			errorType: ErrorTypeFileOperation,
			message:   "this is a very long error message that describes the problem in detail",
			expected:  "report error [file_operation]: this is a very long error message that describes the problem in detail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &ReportError{
				Type:    tt.errorType,
				Message: tt.message,
			}
			result := err.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReportError_Unwrap(t *testing.T) {
	// Test with wrapped error
	originalErr := errors.New("original error")
	reportErr := &ReportError{
		Type:    ErrorTypeGeneration,
		Message: "wrapped error",
		Cause:   originalErr,
	}

	unwrapped := reportErr.Unwrap()
	assert.Equal(t, originalErr, unwrapped)

	// Test without wrapped error
	reportErr2 := &ReportError{
		Type:    ErrorTypeInvalidInput,
		Message: "simple error",
	}

	unwrapped2 := reportErr2.Unwrap()
	assert.Nil(t, unwrapped2)
}

func TestNewReportError(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		message   string
	}{
		{
			name:      "Invalid input error",
			errorType: ErrorTypeInvalidInput,
			message:   "invalid input provided",
		},
		{
			name:      "File operation error",
			errorType: ErrorTypeFileOperation,
			message:   "failed to write file",
		},
		{
			name:      "Empty message",
			errorType: ErrorTypeConfiguration,
			message:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewReportError(tt.errorType, tt.message)

			require.NotNil(t, err)
			assert.Equal(t, tt.errorType, err.Type)
			assert.Equal(t, tt.message, err.Message)
			assert.Nil(t, err.Cause)

			// Test error message format
			expected := fmt.Sprintf("report error [%s]: %s", tt.errorType.String(), tt.message)
			assert.Equal(t, expected, err.Error())
		})
	}
}

func TestNewReportErrorf(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		format    string
		args      []interface{}
		expected  string
	}{
		{
			name:      "Simple format",
			errorType: ErrorTypeInvalidInput,
			format:    "invalid value: %s",
			args:      []interface{}{"test"},
			expected:  "invalid value: test",
		},
		{
			name:      "Multiple arguments",
			errorType: ErrorTypeGeneration,
			format:    "failed to generate %s report for %d resources",
			args:      []interface{}{"JSON", 5},
			expected:  "failed to generate JSON report for 5 resources",
		},
		{
			name:      "No arguments",
			errorType: ErrorTypeFileOperation,
			format:    "file operation failed",
			args:      []interface{}{},
			expected:  "file operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewReportErrorf(tt.errorType, tt.format, tt.args...)

			require.NotNil(t, err)
			assert.Equal(t, tt.errorType, err.Type)
			assert.Equal(t, tt.expected, err.Message)
			assert.Nil(t, err.Cause)
		})
	}
}

func TestWrapReportError(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		message   string
		cause     error
	}{
		{
			name:      "Wrap standard error",
			errorType: ErrorTypeFileOperation,
			message:   "failed to write report",
			cause:     errors.New("permission denied"),
		},
		{
			name:      "Wrap another ReportError",
			errorType: ErrorTypeGeneration,
			message:   "generation failed",
			cause:     NewReportError(ErrorTypeInvalidInput, "invalid data"),
		},
		{
			name:      "Wrap nil error",
			errorType: ErrorTypeConfiguration,
			message:   "config error",
			cause:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapReportError(tt.errorType, tt.message, tt.cause)

			require.NotNil(t, err)
			assert.Equal(t, tt.errorType, err.Type)
			assert.Equal(t, tt.message, err.Message)
			assert.Equal(t, tt.cause, err.Cause)

			// Test unwrapping
			unwrapped := err.Unwrap()
			assert.Equal(t, tt.cause, unwrapped)

			// Test error chain with errors.Is
			if tt.cause != nil {
				assert.True(t, errors.Is(err, tt.cause))
			}
		})
	}
}

func TestIsReportError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		errorType ErrorType
		expected  bool
	}{
		{
			name:      "Matching ReportError",
			err:       NewReportError(ErrorTypeInvalidInput, "test"),
			errorType: ErrorTypeInvalidInput,
			expected:  true,
		},
		{
			name:      "Non-matching ReportError",
			err:       NewReportError(ErrorTypeGeneration, "test"),
			errorType: ErrorTypeInvalidInput,
			expected:  false,
		},
		{
			name:      "Wrapped ReportError",
			err:       WrapReportError(ErrorTypeFileOperation, "wrapper", NewReportError(ErrorTypeInvalidInput, "inner")),
			errorType: ErrorTypeFileOperation,
			expected:  true,
		},
		{
			name:      "Standard error",
			err:       errors.New("standard error"),
			errorType: ErrorTypeInvalidInput,
			expected:  false,
		},
		{
			name:      "Nil error",
			err:       nil,
			errorType: ErrorTypeInvalidInput,
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReportError(tt.err, tt.errorType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetReportErrorType(t *testing.T) {
	// Test direct ReportError
	err1 := NewReportError(ErrorTypeInvalidInput, "test")
	result1 := GetReportErrorType(err1)
	assert.Equal(t, ErrorTypeInvalidInput, result1)

	// Test wrapped ReportError
	err2 := fmt.Errorf("wrapper: %w", NewReportError(ErrorTypeGeneration, "inner"))
	result2 := GetReportErrorType(err2)
	assert.Equal(t, ErrorType(-1), result2) // wrapped errors don't work with type assertion

	// Test standard error
	err3 := errors.New("standard error")
	result3 := GetReportErrorType(err3)
	assert.Equal(t, ErrorType(-1), result3)

	// Test nil error
	result4 := GetReportErrorType(nil)
	assert.Equal(t, ErrorType(-1), result4)
}

// Test error chaining and unwrapping
func TestErrorChaining(t *testing.T) {
	// Create a chain of errors
	originalErr := errors.New("original error")
	reportErr1 := WrapReportError(ErrorTypeInvalidInput, "first wrap", originalErr)
	reportErr2 := WrapReportError(ErrorTypeGeneration, "second wrap", reportErr1)
	standardWrap := fmt.Errorf("standard wrap: %w", reportErr2)

	// Test errors.Is
	assert.True(t, errors.Is(standardWrap, originalErr))
	assert.True(t, errors.Is(standardWrap, reportErr1))
	assert.True(t, errors.Is(standardWrap, reportErr2))

	// Test errors.As
	var reportErr *ReportError
	assert.True(t, errors.As(standardWrap, &reportErr))
	assert.Equal(t, ErrorTypeGeneration, reportErr.Type)
	assert.Equal(t, "second wrap", reportErr.Message)

	// Test GetReportErrorType
	foundErrType := GetReportErrorType(standardWrap)
	assert.Equal(t, ErrorType(-1), foundErrType) // wrapped errors don't work with simple type assertion
}

// Test concurrent error creation
func TestConcurrentErrorCreation(t *testing.T) {
	const numGoroutines = 10
	const numIterations = 100

	results := make(chan *ReportError, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numIterations; j++ {
				errorType := ErrorType(j % 5) // Cycle through error types
				message := fmt.Sprintf("error from goroutine %d, iteration %d", id, j)
				err := NewReportError(errorType, message)
				results <- err
			}
		}(i)
	}

	// Collect and verify results
	for i := 0; i < numGoroutines*numIterations; i++ {
		err := <-results
		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Message)
	}
}

// Test error formatting edge cases
func TestErrorFormattingEdgeCases(t *testing.T) {
	// Test with special characters
	err1 := NewReportError(ErrorTypeInvalidInput, "error with \n newline and \t tab")
	assert.Contains(t, err1.Error(), "\n")
	assert.Contains(t, err1.Error(), "\t")

	// Test with Unicode characters
	err2 := NewReportError(ErrorTypeGeneration, "error with unicode: 你好世界")
	assert.Contains(t, err2.Error(), "你好世界")

	// Test with very long message
	longMessage := string(make([]byte, 1000))
	for i := range longMessage {
		longMessage = longMessage[:i] + "a" + longMessage[i+1:]
	}
	err3 := NewReportError(ErrorTypeFileOperation, longMessage)
	assert.Contains(t, err3.Error(), longMessage)
}

// Test error type validation
func TestErrorTypeValidation(t *testing.T) {
	// Test all valid error types
	validTypes := []ErrorType{
		ErrorTypeInvalidInput,
		ErrorTypeUnsupportedFormat,
		ErrorTypeFileOperation,
		ErrorTypeGeneration,
		ErrorTypeConfiguration,
	}

	for _, errorType := range validTypes {
		t.Run(errorType.String(), func(t *testing.T) {
			err := NewReportError(errorType, "test message")
			assert.Equal(t, errorType, err.Type)
			assert.NotEqual(t, "unknown", errorType.String())
		})
	}

	// Test invalid error type
	invalidType := ErrorType(999)
	assert.Equal(t, "unknown", invalidType.String())
}

// Benchmark tests
func BenchmarkNewReportError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewReportError(ErrorTypeInvalidInput, "benchmark error message")
	}
}

func BenchmarkNewReportErrorf(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewReportErrorf(ErrorTypeGeneration, "benchmark error %d with %s", i, "formatting")
	}
}

func BenchmarkWrapReportError(b *testing.B) {
	originalErr := errors.New("original error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapReportError(ErrorTypeFileOperation, "wrapped error", originalErr)
	}
}

func BenchmarkIsReportError(b *testing.B) {
	err := NewReportError(ErrorTypeInvalidInput, "test error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsReportError(err, ErrorTypeInvalidInput)
	}
}

func BenchmarkGetReportErrorType(b *testing.B) {
	err := fmt.Errorf("wrapped: %w", NewReportError(ErrorTypeGeneration, "inner error"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetReportErrorType(err)
	}
}

func BenchmarkErrorString(b *testing.B) {
	err := NewReportError(ErrorTypeInvalidInput, "benchmark error message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err.Error()
	}
}
