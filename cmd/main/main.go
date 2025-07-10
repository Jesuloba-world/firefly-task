package main

import (
	"fmt"
	"os"

	"firefly-task/config"
	"firefly-task/pkg/app"
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

	// Create command handler and execute
	cmdHandler := app.NewCommandHandler(appInstance)
	if err := cmdHandler.ExecuteRootCommand(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
