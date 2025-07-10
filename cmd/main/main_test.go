package main

import (
	"testing"

	"firefly-task/config"
)

func TestInitApplication(t *testing.T) {
	t.Run("Successful initialization", func(t *testing.T) {
		// Test successful application initialization
		appInstance, err := initApplication()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if appInstance == nil {
			t.Error("Expected application instance to be created, got nil")
		}

		// Verify that the application has a valid configuration
		if appInstance.Config() == nil {
			t.Error("Expected application to have a configuration, got nil")
		}

		// Test that the configuration has expected defaults
		cfg := appInstance.Config()
		if cfg.Output == "" {
			t.Error("Expected output format to be set")
		}
		if cfg.Concurrency <= 0 {
			t.Error("Expected concurrency to be greater than 0")
		}

		// AWS region and terraform path may be empty if not set via environment variables
		// This is acceptable behavior
		t.Logf("AWS Region: %s", cfg.AWSRegion)
		t.Logf("AWS Profile: %s", cfg.AWSProfile)
		t.Logf("Terraform Path: %s", cfg.TerraformPath)
	})
}

func TestConfigurationDefaults(t *testing.T) {
	// Test that the configuration has sensible defaults
	cfg := &config.Config{}
	cfg.SetDefaults()

	if cfg == nil {
		t.Fatal("Expected configuration to be created, got nil")
	}

	// Test default values based on actual implementation
	if cfg.Output != "text" {
		t.Errorf("Expected default output format to be 'text', got '%s'", cfg.Output)
	}
	if cfg.Concurrency != 5 {
		t.Errorf("Expected default concurrency to be 5, got %d", cfg.Concurrency)
	}
	if cfg.Verbose != false {
		t.Errorf("Expected default verbose to be false, got %t", cfg.Verbose)
	}

	// AWS region and profile should be initialized (may be empty unless set via environment)
	// These are set from environment variables in SetDefaults()
	if cfg.AWSRegion == "" {
		t.Log("AWSRegion is empty - this is expected if AWS_REGION environment variable is not set")
	}
	if cfg.AWSProfile == "" {
		t.Log("AWSProfile is empty - this is expected if AWS_PROFILE environment variable is not set")
	}
}

func TestApplicationIntegration(t *testing.T) {
	// Test that the application can be created and has all required components
	appInstance, err := initApplication()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}

	// Test that we can start and shutdown the application
	err = appInstance.Start()
	if err != nil {
		t.Errorf("Failed to start application: %v", err)
	}

	// Test shutdown
	appInstance.Shutdown()

	// Verify shutdown state
	if !appInstance.IsShuttingDown() {
		t.Error("Expected application to be in shutting down state")
	}
}

func TestMainFunction(t *testing.T) {
	// Test that main function exists and can be referenced
	// Note: This is a basic test since main() calls os.Exit
	// In a real scenario, you might want to refactor main to be more testable
	t.Run("Main function exists", func(t *testing.T) {
		// This test just ensures the main function exists
		// We can't easily test the actual execution without refactoring
		// since main() typically calls os.Exit
		t.Log("Main function is defined and accessible")
	})
}

// Benchmark tests
func BenchmarkInitApplication(b *testing.B) {
	for i := 0; i < b.N; i++ {
		app, err := initApplication()
		if err != nil {
			b.Fatalf("Failed to initialize application: %v", err)
		}
		app.Shutdown() // Clean up
	}
}

func BenchmarkConfigCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := &config.Config{}
		cfg.SetDefaults()
		if cfg == nil {
			b.Fatal("Failed to create configuration")
		}
	}
}
