package app

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Start initializes and starts the application
func (a *Application) Start() error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return errors.New("application is already running")
	}
	a.running = true
	a.mu.Unlock()

	log.Println("Starting application...")

	// Validate configuration
	if err := a.config.ValidateConfig(); err != nil {
		return err
	}

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Shutdown signal received")
		a.Shutdown()
	}()

	// Initialize components that need startup
	// This could include health checks, connection validation, etc.
	log.Printf("Application configuration: %s", a.config.String())

	log.Println("Application started successfully")
	return nil
}

// Shutdown gracefully stops the application
func (a *Application) Shutdown() {
	a.mu.Lock()
	if !a.running || a.shuttingDown {
		a.mu.Unlock()
		return
	}
	a.shuttingDown = true
	a.mu.Unlock()

	log.Println("Shutting down application...")

	// Signal shutdown via context
	a.cancelFunc()

	// Set a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for all goroutines to finish
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All tasks completed successfully")
	case <-ctx.Done():
		log.Println("Shutdown timed out, forcing exit")
	}

	log.Println("Application shutdown complete")
}

// Wait blocks until the application context is cancelled
func (a *Application) Wait() {
	<-a.ctx.Done()
}

// WaitWithTimeout blocks until the application context is cancelled or timeout is reached
func (a *Application) WaitWithTimeout(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	select {
	case <-a.ctx.Done():
		return true
	case <-ctx.Done():
		return false
	}
}
