package main

import (
	"fmt"
	"os"
	"strings"

	"firefly-task/config"
	"firefly-task/pkg/app"
	"firefly-task/pkg/logging"
)

// initApplication initializes the application with all dependencies
func initApplication() (*app.Application, error) {
	// Build configuration with defaults and environment variables
	cfg := &config.Config{}
	cfg.SetDefaults()

	// Create application instance using dependency injection
	return app.NewFromContainer(cfg)
}

// Command execution is now handled by the app package

// Command initialization is now handled by the app package

func main() {
	// Initialize application
	appInstance, err := initApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Create command handler and execute with error handling middleware
	cmdHandler := app.NewCommandHandler(appInstance)
	if err := executeWithErrorHandling(cmdHandler); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// executeWithErrorHandling wraps command execution with proper error handling and logging
func executeWithErrorHandling(cmdHandler *app.CommandHandler) error {
	// Get logging configuration from environment or use defaults
	logLevel := getEnvOrDefault("LOG_LEVEL", "info")
	logJSON := getEnvOrDefault("LOG_JSON", "false") == "true"
	isProduction := getEnvOrDefault("ENVIRONMENT", "development") == "production"
	
	// Initialize logger
	logging.InitLogger(logLevel, isProduction)
	logger := logging.GetLogger()
	
	// Log application startup
	logger.Infow("Starting Firefly Task application", 
		"log_level", logLevel,
		"log_json", logJSON,
		"is_production", isProduction)
	
	// Execute command with error logging
	err := cmdHandler.ExecuteRootCommand()
	if err != nil {
		logger.Errorw("Command execution failed", "error", err.Error())
		return err
	}
	
	// Sync logger before exit
	logging.Sync()
	return nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value)
	}
	return defaultValue
}
