package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	
	tests := []struct {
		name      string
		err       error
		operation string
		details   []interface{}
		expected  string
	}{
		{
			name:      "wrap error without details",
			err:       originalErr,
			operation: "get_instance",
			details:   nil,
			expected:  "get_instance failed: original error",
		},
		{
			name:      "wrap error with details",
			err:       originalErr,
			operation: "get_instance",
			details:   []interface{}{"instance %s not found", "i-123"},
			expected:  "get_instance failed: instance i-123 not found: original error",
		},
		{
			name:      "wrap nil error",
			err:       nil,
			operation: "get_instance",
			details:   nil,
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.err, tt.operation, tt.details...)
			
			if tt.err == nil {
				if result != nil {
					t.Errorf("Expected nil error, got %v", result)
				}
				return
			}
			
			if result == nil {
				t.Fatal("Expected non-nil error")
			}
			
			if result.Error() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Error())
			}
			
			// Test error unwrapping
			if !errors.Is(result, originalErr) {
				t.Error("Expected wrapped error to be unwrappable to original error")
			}
		})
	}
}

func TestWrapErrorf(t *testing.T) {
	originalErr := errors.New("AWS API error")
	
	result := WrapErrorf(originalErr, "failed to get EC2 instance %s", "i-123")
	
	expected := "failed to get EC2 instance i-123: AWS API error"
	if result.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, result.Error())
	}
	
	// Test error unwrapping
	if !errors.Is(result, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		message   string
		args      []interface{}
		expected  string
	}{
		{
			name:      "simple error",
			operation: "validate_config",
			message:   "invalid configuration",
			args:      nil,
			expected:  "validate_config failed: invalid configuration",
		},
		{
			name:      "formatted error",
			operation: "parse_file",
			message:   "invalid syntax at line %d",
			args:      []interface{}{42},
			expected:  "parse_file failed: invalid syntax at line 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewError(tt.operation, tt.message, tt.args...)
			
			if result.Error() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result.Error())
			}
		})
	}
}

func TestAWSErrorHelpers(t *testing.T) {
	originalErr := errors.New("access denied")
	
	// Test WrapAWSError
	awsErr := WrapAWSError(originalErr, "DescribeInstances", "i-123")
	expected := "AWS DescribeInstances failed for resource i-123: access denied"
	if awsErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, awsErr.Error())
	}
	if !errors.Is(awsErr, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
	
	// Test WrapEC2Error
	ec2Err := WrapEC2Error(originalErr, "DescribeInstances", "i-456")
	expected = "EC2 DescribeInstances failed for instance i-456: access denied"
	if ec2Err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, ec2Err.Error())
	}
	if !errors.Is(ec2Err, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
}

func TestTerraformErrorHelpers(t *testing.T) {
	originalErr := errors.New("syntax error")
	
	// Test WrapTerraformError
	terraformErr := WrapTerraformError(originalErr, "plan", "aws_instance.web")
	expected := "Terraform plan failed for resource aws_instance.web: syntax error"
	if terraformErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, terraformErr.Error())
	}
	if !errors.Is(terraformErr, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
	
	// Test WrapParseError with line number
	parseErr := WrapParseError(originalErr, "main.tf", 42)
	expected = "parse error in main.tf at line 42: syntax error"
	if parseErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, parseErr.Error())
	}
	
	// Test WrapParseError without line number
	parseErr2 := WrapParseError(originalErr, "main.tf", 0)
	expected = "parse error in main.tf: syntax error"
	if parseErr2.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, parseErr2.Error())
	}
}

func TestFileErrorHelpers(t *testing.T) {
	originalErr := errors.New("permission denied")
	
	fileErr := WrapFileError(originalErr, "read", "/path/to/file.txt")
	expected := "file read failed for /path/to/file.txt: permission denied"
	if fileErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, fileErr.Error())
	}
	if !errors.Is(fileErr, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
}

func TestValidationErrorHelpers(t *testing.T) {
	// Test NewValidationError
	validationErr := NewValidationError("instance_id", "invalid-id", "must start with i-")
	expected := "validation failed for field instance_id with value invalid-id: must start with i-"
	if validationErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, validationErr.Error())
	}
	
	// Test WrapValidationError
	originalErr := errors.New("field is required")
	wrappedErr := WrapValidationError(originalErr, "user input")
	expected = "validation error in user input: field is required"
	if wrappedErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, wrappedErr.Error())
	}
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
}

func TestNetworkErrorHelpers(t *testing.T) {
	originalErr := errors.New("connection timeout")
	
	networkErr := WrapNetworkError(originalErr, "GET", "https://api.aws.com")
	expected := "network GET failed for endpoint https://api.aws.com: connection timeout"
	if networkErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, networkErr.Error())
	}
	if !errors.Is(networkErr, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
}

func TestDatabaseErrorHelpers(t *testing.T) {
	originalErr := errors.New("table not found")
	
	dbErr := WrapDatabaseError(originalErr, "SELECT", "users")
	expected := "database SELECT failed for table users: table not found"
	if dbErr.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, dbErr.Error())
	}
	if !errors.Is(dbErr, originalErr) {
		t.Error("Expected wrapped error to be unwrappable to original error")
	}
}

func TestNilErrorHandling(t *testing.T) {
	// All error wrapping functions should return nil when given nil error
	funcs := []func() error{
		func() error { return WrapError(nil, "operation") },
		func() error { return WrapErrorf(nil, "format %s", "arg") },
		func() error { return WrapAWSError(nil, "operation", "resource") },
		func() error { return WrapEC2Error(nil, "operation", "instance") },
		func() error { return WrapTerraformError(nil, "operation", "resource") },
		func() error { return WrapParseError(nil, "file", 1) },
		func() error { return WrapFileError(nil, "operation", "file") },
		func() error { return WrapConfigError(nil, "key") },
		func() error { return WrapValidationError(nil, "context") },
		func() error { return WrapNetworkError(nil, "operation", "endpoint") },
		func() error { return WrapDatabaseError(nil, "operation", "table") },
	}
	
	for i, fn := range funcs {
		if err := fn(); err != nil {
			t.Errorf("Function %d should return nil for nil input, got %v", i, err)
		}
	}
}

func TestErrorChaining(t *testing.T) {
	// Test multiple levels of error wrapping
	originalErr := errors.New("root cause")
	level1 := WrapAWSError(originalErr, "DescribeInstances", "i-123")
	level2 := WrapError(level1, "drift_detection", "failed to get instance details")
	level3 := WrapErrorf(level2, "command execution failed")
	
	// Should be able to unwrap to any level
	if !errors.Is(level3, originalErr) {
		t.Error("Expected to unwrap to original error")
	}
	if !errors.Is(level3, level1) {
		t.Error("Expected to unwrap to level1 error")
	}
	if !errors.Is(level3, level2) {
		t.Error("Expected to unwrap to level2 error")
	}
	
	// Check final error message contains context from all levels
	finalMsg := level3.Error()
	if finalMsg != "" {
		t.Logf("Final error message: %s", finalMsg)
	}
	
	// Verify the error message contains expected components
	if !strings.Contains(finalMsg, "root cause") {
		t.Error("Expected final error message to contain original error")
	}
	if !strings.Contains(finalMsg, "AWS DescribeInstances") {
		t.Error("Expected final error message to contain AWS operation context")
	}
}