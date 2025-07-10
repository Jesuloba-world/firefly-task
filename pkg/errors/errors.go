package errors

import (
	"fmt"
)

// WrapError wraps an error with additional context information
func WrapError(err error, operation string, details ...interface{}) error {
	if err == nil {
		return nil
	}
	
	if len(details) > 0 {
		message := fmt.Sprintf(details[0].(string), details[1:]...)
		return fmt.Errorf("%s failed: %s: %w", operation, message, err)
	}
	
	return fmt.Errorf("%s failed: %w", operation, err)
}

// WrapErrorf wraps an error with a formatted message
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}

// NewError creates a new error with context
func NewError(operation string, message string, args ...interface{}) error {
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	return fmt.Errorf("%s failed: %s", operation, message)
}

// AWS-specific error helpers

// WrapAWSError wraps AWS-related errors with additional context
func WrapAWSError(err error, operation string, resourceID string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("AWS %s failed for resource %s: %w", operation, resourceID, err)
}

// WrapEC2Error wraps EC2-specific errors
func WrapEC2Error(err error, operation string, instanceID string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("EC2 %s failed for instance %s: %w", operation, instanceID, err)
}

// Terraform-specific error helpers

// WrapTerraformError wraps Terraform-related errors
func WrapTerraformError(err error, operation string, resource string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("Terraform %s failed for resource %s: %w", operation, resource, err)
}

// WrapParseError wraps parsing errors with file context
func WrapParseError(err error, filePath string, lineNumber int) error {
	if err == nil {
		return nil
	}
	if lineNumber > 0 {
		return fmt.Errorf("parse error in %s at line %d: %w", filePath, lineNumber, err)
	}
	return fmt.Errorf("parse error in %s: %w", filePath, err)
}

// File system error helpers

// WrapFileError wraps file system errors
func WrapFileError(err error, operation string, filePath string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("file %s failed for %s: %w", operation, filePath, err)
}

// Configuration error helpers

// WrapConfigError wraps configuration-related errors
func WrapConfigError(err error, configKey string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("configuration error for %s: %w", configKey, err)
}

// Validation error helpers

// NewValidationError creates a validation error
func NewValidationError(field string, value interface{}, reason string) error {
	return fmt.Errorf("validation failed for field %s with value %v: %s", field, value, reason)
}

// WrapValidationError wraps validation errors with context
func WrapValidationError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("validation error in %s: %w", context, err)
}

// Network error helpers

// WrapNetworkError wraps network-related errors
func WrapNetworkError(err error, operation string, endpoint string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("network %s failed for endpoint %s: %w", operation, endpoint, err)
}

// Database error helpers

// WrapDatabaseError wraps database-related errors
func WrapDatabaseError(err error, operation string, table string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("database %s failed for table %s: %w", operation, table, err)
}