package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"firefly-task/config"
	"github.com/stretchr/testify/assert"
)

func TestApplication_Start(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	// Test successful start
	err := app.Start()
	assert.NoError(t, err)
	assert.False(t, app.IsShuttingDown())

	// Test that starting again returns error
	err = app.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "application is already running")

	// Clean up
	app.Shutdown()
}

func TestApplication_Shutdown(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	// Start the application
	err := app.Start()
	assert.NoError(t, err)
	assert.False(t, app.IsShuttingDown())

	// Test shutdown
	app.Shutdown()
	assert.True(t, app.IsShuttingDown())

	// Test that shutdown is idempotent
	app.Shutdown()
	assert.True(t, app.IsShuttingDown())
}

func TestApplication_Wait(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	err := app.Start()
	assert.NoError(t, err)

	// Start a goroutine that will shutdown the app after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		app.Shutdown()
	}()

	// Wait should block until shutdown
	start := time.Now()
	app.Wait()
	duration := time.Since(start)

	assert.True(t, app.IsShuttingDown())
	assert.True(t, duration >= 100*time.Millisecond)
	assert.True(t, duration < 200*time.Millisecond) // Should not take too long
}

func TestApplication_WaitWithTimeout(t *testing.T) {
	t.Run("Timeout before shutdown", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.SetDefaults()
		mockEC2 := &MockEC2Client{}
		mockTF := &MockTerraformParser{}
		mockDrift := &MockDriftDetector{}
		mockReport := &MockReportGenerator{}
		app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)
		err := app.Start()
		assert.NoError(t, err)

		// Test timeout case
		start := time.Now()
		done := app.WaitWithTimeout(50 * time.Millisecond)
		duration := time.Since(start)

		assert.False(t, done)
		assert.True(t, duration >= 50*time.Millisecond)
		assert.True(t, duration < 100*time.Millisecond)
		assert.False(t, app.IsShuttingDown())
		app.Shutdown() // Clean up
	})

	t.Run("Shutdown before timeout", func(t *testing.T) {
		cfg := &config.Config{}
		cfg.SetDefaults()
		mockEC2 := &MockEC2Client{}
		mockTF := &MockTerraformParser{}
		mockDrift := &MockDriftDetector{}
		mockReport := &MockReportGenerator{}
		app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)
		err := app.Start()
		assert.NoError(t, err)

		// Start a goroutine that will shutdown the app after a short delay
		go func() {
			time.Sleep(25 * time.Millisecond)
			app.Shutdown()
		}()

		// Wait with a longer timeout
		start := time.Now()
		done := app.WaitWithTimeout(100 * time.Millisecond)
		duration := time.Since(start)

		assert.True(t, done)
		assert.True(t, duration >= 25*time.Millisecond)
		assert.True(t, duration < 75*time.Millisecond)
		assert.True(t, app.IsShuttingDown())
	})
}

func TestApplication_SignalHandling(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	err := app.Start()
	assert.NoError(t, err)
	assert.False(t, app.IsShuttingDown())

	// Simulate signal handling by calling Shutdown directly
	go func() {
		time.Sleep(50 * time.Millisecond)
		app.Shutdown()
	}()

	// Wait for shutdown
	start := time.Now()
	app.Wait()
	duration := time.Since(start)

	assert.True(t, app.IsShuttingDown())
	assert.True(t, duration >= 50*time.Millisecond)
	assert.True(t, duration < 150*time.Millisecond)
}

func TestApplication_ShutdownTimeout(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	err := app.Start()
	assert.NoError(t, err)

	// Test that shutdown completes within reasonable time
	start := time.Now()
	app.Shutdown()
	duration := time.Since(start)

	assert.True(t, app.IsShuttingDown())
	assert.True(t, duration < 100*time.Millisecond) // Should be very fast
}

func TestApplication_ConcurrentOperations(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	err := app.Start()
	assert.NoError(t, err)

	// Test concurrent IsShuttingDown calls
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				app.IsShuttingDown()
			}
		}()
	}

	shutdownComplete := make(chan struct{})
	// Shutdown after a delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		app.Shutdown()
		close(shutdownComplete)
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// Wait for shutdown to be called
	<-shutdownComplete

	assert.True(t, app.IsShuttingDown())
}

func TestApplication_ContextCancellation(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	err := app.Start()
	assert.NoError(t, err)

	// Test that context is cancelled when app shuts down
	ctx := app.ctx
	assert.NotNil(t, ctx)

	select {
	case <-ctx.Done():
		t.Fatal("Context should not be cancelled yet")
	default:
		// Context is not cancelled, which is expected
	}

	// Shutdown the application
	app.Shutdown()

	// Context should be cancelled
	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Context should be cancelled after shutdown")
	}

	assert.Equal(t, context.Canceled, ctx.Err())
}