package app

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"firefly-task/config"
)

func TestNewCommandHandler(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)

	// Create command handler
	handler := NewCommandHandler(app)

	if handler == nil {
		t.Fatal("Expected command handler to be created, got nil")
	}

	if handler.app != app {
		t.Error("Expected command handler to have the correct app instance")
	}
}

func TestCreateRootCommand(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)
	handler := NewCommandHandler(app)

	// Create root command
	rootCmd := handler.CreateRootCommand()

	if rootCmd == nil {
		t.Fatal("Expected root command to be created, got nil")
	}

	if rootCmd.Use != "firefly-task" {
		t.Errorf("Expected command use to be 'firefly-task', got '%s'", rootCmd.Use)
	}

	if !strings.Contains(rootCmd.Short, "drift") {
		t.Error("Expected short description to mention drift")
	}

	// Check that subcommands are added
	subcommands := rootCmd.Commands()
	expectedCommands := []string{"check", "batch", "attribute"}

	if len(subcommands) != len(expectedCommands) {
		t.Errorf("Expected %d subcommands, got %d", len(expectedCommands), len(subcommands))
	}

	for _, expectedCmd := range expectedCommands {
		found := false
		for _, cmd := range subcommands {
			if cmd.Use == expectedCmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected subcommand '%s' not found", expectedCmd)
		}
	}
}

func TestCreateCheckCommand(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)
	handler := NewCommandHandler(app)

	// Create check command
	checkCmd := handler.CreateCheckCommand()

	if checkCmd == nil {
		t.Fatal("Expected check command to be created, got nil")
	}

	if checkCmd.Use != "check" {
		t.Errorf("Expected command use to be 'check', got '%s'", checkCmd.Use)
	}

	// Check required flags
	requiredFlags := []string{"instance-id", "tf-path"}
	for _, flagName := range requiredFlags {
		flag := checkCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagName)
			continue
		}
	}

	// Check optional flags
	optionalFlags := []string{"output", "attributes"}
	for _, flagName := range optionalFlags {
		flag := checkCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagName)
		}
	}
}

func TestCreateBatchCommand(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)
	handler := NewCommandHandler(app)

	// Create batch command
	batchCmd := handler.CreateBatchCommand()

	if batchCmd == nil {
		t.Fatal("Expected batch command to be created, got nil")
	}

	if batchCmd.Use != "batch" {
		t.Errorf("Expected command use to be 'batch', got '%s'", batchCmd.Use)
	}

	// Check required flags
	requiredFlags := []string{"input-file", "tf-path"}
	for _, flagName := range requiredFlags {
		flag := batchCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagName)
			continue
		}
	}

	// Check optional flags
	optionalFlags := []string{"output", "attributes"}
	for _, flagName := range optionalFlags {
		flag := batchCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagName)
		}
	}
}

func TestCreateAttributeCommand(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)
	handler := NewCommandHandler(app)

	// Create attribute command
	attributeCmd := handler.CreateAttributeCommand()

	if attributeCmd == nil {
		t.Fatal("Expected attribute command to be created, got nil")
	}

	if attributeCmd.Use != "attribute" {
		t.Errorf("Expected command use to be 'attribute', got '%s'", attributeCmd.Use)
	}

	// Check required flags
	requiredFlags := []string{"instance-id", "tf-path", "attribute"}
	for _, flagName := range requiredFlags {
		flag := attributeCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagName)
			continue
		}
	}

	// Check optional flags
	optionalFlags := []string{"output"}
	for _, flagName := range optionalFlags {
		flag := attributeCmd.Flags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to exist", flagName)
		}
	}
}

func TestOutputResult(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)
	handler := NewCommandHandler(app)

	testData := []byte("test report data")

	t.Run("Output to file", func(t *testing.T) {
		// Create temporary file
		tempFile, err := os.CreateTemp("", "test_output_*.txt")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		tempFile.Close()

		// Test output to file
		err = handler.outputResult(testData, tempFile.Name())
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// Verify file content
		content, err := os.ReadFile(tempFile.Name())
		if err != nil {
			t.Fatalf("Failed to read temp file: %v", err)
		}

		if !bytes.Equal(content, testData) {
			t.Errorf("Expected file content to be '%s', got '%s'", string(testData), string(content))
		}
	})

	t.Run("Output to stdout", func(t *testing.T) {
		// Test output to stdout (no file specified)
		err := handler.outputResult(testData, "")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		// Note: We can't easily test stdout output in unit tests
		// This test mainly ensures no error occurs
	})

	t.Run("Invalid file path", func(t *testing.T) {
		// Test with invalid file path
		err := handler.outputResult(testData, "/invalid/path/file.txt")
		if err == nil {
			t.Error("Expected error for invalid file path, got nil")
		}
	})
}

func TestExecuteCommand(t *testing.T) {
	// Create a mock application
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockAWSClient := &MockEC2Client{}
	mockTerraformParser := &MockTerraformParser{}
	mockDriftDetector := &MockDriftDetector{}
	mockReportGenerator := &MockReportGenerator{}

	app := New(cfg, mockAWSClient, mockTerraformParser, mockDriftDetector, mockReportGenerator)
	handler := NewCommandHandler(app)

	t.Run("Help command", func(t *testing.T) {
		// Test help command
		err := handler.ExecuteCommand([]string{"--help"})
		// Help command should not return an error
		if err != nil {
			t.Errorf("Expected no error for help command, got: %v", err)
		}
	})

	t.Run("Invalid command", func(t *testing.T) {
		// Test invalid command
		err := handler.ExecuteCommand([]string{"invalid-command"})
		if err == nil {
			t.Error("Expected error for invalid command, got nil")
		}
	})
}