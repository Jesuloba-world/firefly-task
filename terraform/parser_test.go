package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test data for parser integration tests
const sampleIntegrationState = `{
  "format_version": "1.0",
  "version": 4,
  "terraform_version": "1.0.0",
  "serial": 1,
  "lineage": "test-lineage",
  "outputs": {},
  "resources": [],
  "values": {
    "root_module": {
      "resources": [
        {
          "address": "aws_instance.web",
          "mode": "managed",
          "type": "aws_instance",
          "name": "web",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "ami": "ami-12345678",
            "instance_type": "t3.micro",
            "key_name": "my-key",
            "subnet_id": "subnet-12345",
            "vpc_security_group_ids": ["sg-12345"],
            "tags": {
              "Name": "WebServer",
              "Environment": "test"
            }
          }
        },
        {
          "address": "aws_instance.db",
          "mode": "managed",
          "type": "aws_instance",
          "name": "db",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "ami": "ami-87654321",
            "instance_type": "t3.small",
            "subnet_id": "subnet-67890",
            "vpc_security_group_ids": ["sg-67890"],
            "tags": {
              "Name": "Database",
              "Environment": "test"
            }
          }
        }
      ]
    }
  }
}`

const sampleHCLConfig = `resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"
  key_name      = "my-key"
  subnet_id     = "subnet-12345"
  
  vpc_security_group_ids = ["sg-12345"]
  
  tags = {
    Name        = "WebServer"
    Environment = "test"
  }
}

resource "aws_instance" "db" {
  ami           = "ami-87654321"
  instance_type = "t3.small"
  subnet_id     = "subnet-67890"
  
  vpc_security_group_ids = ["sg-67890"]
  
  tags = {
    Name        = "Database"
    Environment = "test"
  }
}`

func TestNewParser(t *testing.T) {
	parser := NewParser()
	if parser == nil {
		t.Fatal("NewParser() returned nil")
	}

	options := parser.GetOptions()
	if options.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected default MaxFileSize to be 10MB, got %d", options.MaxFileSize)
	}

	if !options.IgnoreMissingFiles {
		t.Error("Expected default IgnoreMissingFiles to be true")
	}
}

func TestNewParserWithOptions(t *testing.T) {
	customOptions := ParserOptions{
		StrictMode:         true,
		IgnoreMissingFiles: false,
		MaxFileSize:        5 * 1024 * 1024,
		SupportedFormats:   []string{".tf"},
		IncludeMetadata:    false,
		RecursiveModules:   false,
	}

	parser := NewParserWithOptions(customOptions)
	if parser == nil {
		t.Fatal("NewParserWithOptions() returned nil")
	}

	options := parser.GetOptions()
	if !options.StrictMode {
		t.Error("Expected StrictMode to be true")
	}

	if options.IgnoreMissingFiles {
		t.Error("Expected IgnoreMissingFiles to be false")
	}

	if options.MaxFileSize != 5*1024*1024 {
		t.Errorf("Expected MaxFileSize to be 5MB, got %d", options.MaxFileSize)
	}
}

func TestParserSetOptions(t *testing.T) {
	parser := NewParser()

	newOptions := ParserOptions{
		StrictMode:         true,
		IgnoreMissingFiles: false,
	}

	parser.SetOptions(newOptions)
	options := parser.GetOptions()

	if !options.StrictMode {
		t.Error("Expected StrictMode to be true after SetOptions")
	}

	if options.IgnoreMissingFiles {
		t.Error("Expected IgnoreMissingFiles to be false after SetOptions")
	}
}

func TestParseStateIntegration(t *testing.T) {
	// Create temporary state file
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "terraform.tfstate")

	err := os.WriteFile(stateFile, []byte(sampleIntegrationState), 0644)
	if err != nil {
		t.Fatalf("Failed to create test state file: %v", err)
	}

	parser := NewParser()
	configs, err := parser.ParseState(stateFile)
	if err != nil {
		t.Fatalf("ParseState failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configurations, got %d", len(configs))
	}

	// Check web instance
	webConfig, exists := configs["aws_instance.web"]
	if !exists {
		t.Error("Expected aws_instance.web configuration not found")
	} else {
		if webConfig.Attributes["instance_type"] != "t3.micro" {
			t.Errorf("Expected instance type t3.micro, got %s", webConfig.Attributes["instance_type"])
		}
		if webConfig.Attributes["ami"] != "ami-12345678" {
			t.Errorf("Expected AMI ami-12345678, got %s", webConfig.Attributes["ami"])
		}
		if webConfig.TerraformVersion != "1.0.0" {
			t.Errorf("Expected Terraform version 1.0.0, got %s", webConfig.TerraformVersion)
		}
	}

	// Check db instance
	dbConfig, exists := configs["aws_instance.db"]
	if !exists {
		t.Error("Expected aws_instance.db configuration not found")
	} else {
		if dbConfig.Attributes["instance_type"] != "t3.small" {
			t.Errorf("Expected instance type t3.small, got %s", dbConfig.Attributes["instance_type"])
		}
		if dbConfig.Attributes["ami"] != "ami-87654321" {
			t.Errorf("Expected AMI ami-87654321, got %s", dbConfig.Attributes["ami"])
		}
	}
}

func TestParseHCLIntegration(t *testing.T) {
	// Create temporary HCL file
	tempDir := t.TempDir()
	hclFile := filepath.Join(tempDir, "main.tf")

	err := os.WriteFile(hclFile, []byte(sampleHCLConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test HCL file: %v", err)
	}

	parser := NewParser()
	configs, err := parser.ParseHCL(tempDir)
	if err != nil {
		t.Fatalf("ParseHCL failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configurations, got %d", len(configs))
	}

	// Check web instance exists (terraform-config-inspect doesn't expose detailed config)
	webConfig, exists := configs["aws_instance.web"]
	if !exists {
		t.Error("Expected aws_instance.web configuration not found")
	} else {
		if webConfig.ResourceName != "web" {
			t.Errorf("Expected resource name 'web', got %s", webConfig.ResourceName)
		}
	}

	// Check db instance exists
	dbConfig, exists := configs["aws_instance.db"]
	if !exists {
		t.Error("Expected aws_instance.db configuration not found")
	} else {
		if dbConfig.ResourceName != "db" {
			t.Errorf("Expected resource name 'db', got %s", dbConfig.ResourceName)
		}
	}
}

func TestParseBothIntegration(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	stateFile := filepath.Join(tempDir, "terraform.tfstate")
	hclFile := filepath.Join(tempDir, "main.tf")

	err := os.WriteFile(stateFile, []byte(sampleIntegrationState), 0644)
	if err != nil {
		t.Fatalf("Failed to create test state file: %v", err)
	}

	err = os.WriteFile(hclFile, []byte(sampleHCLConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create test HCL file: %v", err)
	}

	parser := NewParser()
	configs, err := parser.ParseBoth(stateFile, tempDir)
	if err != nil {
		t.Fatalf("ParseBoth failed: %v", err)
	}

	if len(configs) != 2 {
		t.Errorf("Expected 2 configurations, got %d", len(configs))
	}

	// Check that state data is merged (should have Terraform version from state)
	webConfig, exists := configs["aws_instance.web"]
	if !exists {
		t.Error("Expected aws_instance.web configuration not found")
	} else {
		if webConfig.TerraformVersion != "1.0.0" {
			t.Errorf("Expected Terraform version from state (1.0.0), got %s", webConfig.TerraformVersion)
		}
		if webConfig.ResourceName != "web" {
			t.Errorf("Expected resource name 'web', got %s", webConfig.ResourceName)
		}
	}
}

func TestParseStateWithMissingFile(t *testing.T) {
	parser := NewParser()

	// Test with IgnoreMissingFiles = true (default)
	configs, err := parser.ParseState("/nonexistent/file.tfstate")
	if err != nil {
		t.Errorf("Expected no error with IgnoreMissingFiles=true, got: %v", err)
	}
	if len(configs) != 0 {
		t.Errorf("Expected empty configs with missing file, got %d configs", len(configs))
	}

	// Test with IgnoreMissingFiles = false
	options := parser.GetOptions()
	options.IgnoreMissingFiles = false
	parser.SetOptions(options)

	_, err = parser.ParseState("/nonexistent/file.tfstate")
	if err == nil {
		t.Error("Expected error with IgnoreMissingFiles=false and missing file")
	}
}

func TestParseStateWithStrictMode(t *testing.T) {
	// Create temporary file with wrong extension
	tempDir := t.TempDir()
	wrongExtFile := filepath.Join(tempDir, "terraform.txt")

	err := os.WriteFile(wrongExtFile, []byte(sampleIntegrationState), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewParser()

	// Test with StrictMode = false (default)
	_, err = parser.ParseState(wrongExtFile)
	if err != nil {
		t.Errorf("Expected no error with StrictMode=false, got: %v", err)
	}

	// Test with StrictMode = true
	options := parser.GetOptions()
	options.StrictMode = true
	parser.SetOptions(options)

	_, err = parser.ParseState(wrongExtFile)
	if err == nil {
		t.Error("Expected error with StrictMode=true and unsupported file extension")
	}
}

func TestParserErrorHandling(t *testing.T) {
	parser := NewParser()

	// Test with invalid JSON and IgnoreMissingFiles=false
	options := parser.GetOptions()
	options.IgnoreMissingFiles = false
	parser.SetOptions(options)

	tempDir := t.TempDir()
	invalidFile := filepath.Join(tempDir, "invalid.tfstate")

	err := os.WriteFile(invalidFile, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err = parser.ParseState(invalidFile)
	if err == nil {
		t.Error("Expected error when parsing invalid JSON with IgnoreMissingFiles=false")
	}
}

func TestDefaultParserOptions(t *testing.T) {
	options := DefaultParserOptions()

	if options.StrictMode {
		t.Error("Expected default StrictMode to be false")
	}

	if !options.IgnoreMissingFiles {
		t.Error("Expected default IgnoreMissingFiles to be true")
	}

	if options.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected default MaxFileSize to be 10MB, got %d", options.MaxFileSize)
	}

	expectedFormats := []string{".tf", ".tf.json", ".tfstate"}
	if len(options.SupportedFormats) != len(expectedFormats) {
		t.Errorf("Expected %d supported formats, got %d", len(expectedFormats), len(options.SupportedFormats))
	}

	if !options.IncludeMetadata {
		t.Error("Expected default IncludeMetadata to be true")
	}

	if !options.RecursiveModules {
		t.Error("Expected default RecursiveModules to be true")
	}
}

func TestParseError(t *testing.T) {
	err := NewParseError("test", "test message", "/test/path", nil)

	if err.Type != "test" {
		t.Errorf("Expected error type 'test', got '%s'", err.Type)
	}

	if err.Message != "test message" {
		t.Errorf("Expected error message 'test message', got '%s'", err.Message)
	}

	if err.Path != "/test/path" {
		t.Errorf("Expected error path '/test/path', got '%s'", err.Path)
	}

	errorStr := err.Error()
	if !strings.Contains(errorStr, "test error in /test/path: test message") {
		t.Errorf("Unexpected error string: %s", errorStr)
	}
}

func TestParseErrorWithCause(t *testing.T) {
	cause := fmt.Errorf("underlying error")
	err := NewParseError("test", "test message", "/test/path", cause)

	if err.Unwrap() != cause {
		t.Error("Expected Unwrap() to return the cause error")
	}

	errorStr := err.Error()
	if !strings.Contains(errorStr, "caused by: underlying error") {
		t.Errorf("Expected error string to contain cause, got: %s", errorStr)
	}
}

func TestParserErrorAccumulation(t *testing.T) {
	parser := NewParser()

	// Initially no errors
	errors := parser.GetErrors()
	if len(errors) != 0 {
		t.Errorf("Expected no initial errors, got %d", len(errors))
	}

	// Parse non-existent file with IgnoreMissingFiles=true
	// This should accumulate an error but not fail
	parser.ParseState("/nonexistent/file.tfstate")

	errors = parser.GetErrors()
	if len(errors) != 1 {
		t.Errorf("Expected 1 accumulated error, got %d", len(errors))
	}

	// Clear errors
	parser.ClearErrors()
	errors = parser.GetErrors()
	if len(errors) != 0 {
		t.Errorf("Expected no errors after ClearErrors(), got %d", len(errors))
	}
}
