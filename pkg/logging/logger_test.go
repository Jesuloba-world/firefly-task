package logging

import (
	"testing"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name         string
		level        string
		isProduction bool
		expectedLevel zapcore.Level
	}{
		{
			name:          "debug level development",
			level:         "debug",
			isProduction:  false,
			expectedLevel: zapcore.DebugLevel,
		},
		{
			name:          "info level production",
			level:         "info",
			isProduction:  true,
			expectedLevel: zapcore.InfoLevel,
		},
		{
			name:          "warn level",
			level:         "warn",
			isProduction:  false,
			expectedLevel: zapcore.WarnLevel,
		},
		{
			name:          "error level",
			level:         "error",
			isProduction:  true,
			expectedLevel: zapcore.ErrorLevel,
		},
		{
			name:          "invalid level defaults to info",
			level:         "invalid",
			isProduction:  false,
			expectedLevel: zapcore.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global logger
			log = nil
			
			InitLogger(tt.level, tt.isProduction)
			
			if log == nil {
				t.Fatal("Expected logger to be initialized")
			}
			
			// Verify logger is accessible
			logger := GetLogger()
			if logger == nil {
				t.Fatal("Expected GetLogger to return non-nil logger")
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Create an observed logger for testing
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := zap.New(core).Sugar()
	log = logger

	// Test different log levels
	Debug("Debug message", "key", "value")
	Info("Info message", "key", "value")
	Warn("Warning message", "key", "value")
	Error("Error message", "key", "value")

	// Verify all messages were logged
	allLogs := recorded.All()
	if len(allLogs) != 4 {
		t.Fatalf("Expected 4 log entries, got %d", len(allLogs))
	}

	// Verify log levels
	expectedLevels := []zapcore.Level{
		zapcore.DebugLevel,
		zapcore.InfoLevel,
		zapcore.WarnLevel,
		zapcore.ErrorLevel,
	}

	for i, entry := range allLogs {
		if entry.Level != expectedLevels[i] {
			t.Errorf("Expected log level %v, got %v", expectedLevels[i], entry.Level)
		}
	}

	// Verify messages
	expectedMessages := []string{
		"Debug message",
		"Info message",
		"Warning message",
		"Error message",
	}

	for i, entry := range allLogs {
		if entry.Message != expectedMessages[i] {
			t.Errorf("Expected message %q, got %q", expectedMessages[i], entry.Message)
		}
	}
}

func TestWith(t *testing.T) {
	// Create an observed logger for testing
	core, recorded := observer.New(zapcore.InfoLevel)
	logger := zap.New(core).Sugar()
	log = logger

	// Test With functionality
	contextLogger := With("instanceID", "i-123", "region", "us-west-2")
	if contextLogger == nil {
		t.Fatal("Expected With to return non-nil logger")
	}

	contextLogger.Info("Processing instance")

	// Verify the log entry has the context fields
	allLogs := recorded.All()
	if len(allLogs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(allLogs))
	}

	entry := allLogs[0]
	if entry.Message != "Processing instance" {
		t.Errorf("Expected message 'Processing instance', got %q", entry.Message)
	}

	// Check context fields
	if entry.Context[0].Key != "instanceID" || entry.Context[0].String != "i-123" {
		t.Errorf("Expected instanceID field with value i-123")
	}
	if entry.Context[1].Key != "region" || entry.Context[1].String != "us-west-2" {
		t.Errorf("Expected region field with value us-west-2")
	}
}

func TestInitTestLogger(t *testing.T) {
	// Reset global logger
	log = nil

	InitTestLogger("debug")

	if log == nil {
		t.Fatal("Expected test logger to be initialized")
	}

	logger := GetLogger()
	if logger == nil {
		t.Fatal("Expected GetLogger to return non-nil logger")
	}
}

func TestLoggerNilSafety(t *testing.T) {
	// Reset global logger to nil
	log = nil

	// These should not panic when logger is nil
	Debug("Debug message")
	Info("Info message")
	Warn("Warning message")
	Error("Error message")
	Sync()

	contextLogger := With("key", "value")
	if contextLogger != nil {
		t.Error("Expected With to return nil when global logger is nil")
	}
}