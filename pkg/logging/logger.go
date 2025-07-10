package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

// InitLogger initializes the global logger with appropriate configuration
func InitLogger(level string, isProduction bool) {
	var cfg zap.Config
	if isProduction {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	// Set log level based on input
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	log = logger.Sugar()
}

// GetLogger returns the global logger instance
func GetLogger() *zap.SugaredLogger {
	return log
}

// Debug logs a debug message with optional fields
func Debug(msg string, fields ...interface{}) {
	if log != nil {
		log.Debugw(msg, fields...)
	}
}

// Info logs an info message with optional fields
func Info(msg string, fields ...interface{}) {
	if log != nil {
		log.Infow(msg, fields...)
	}
}

// Warn logs a warning message with optional fields
func Warn(msg string, fields ...interface{}) {
	if log != nil {
		log.Warnw(msg, fields...)
	}
}

// Error logs an error message with optional fields
func Error(msg string, fields ...interface{}) {
	if log != nil {
		log.Errorw(msg, fields...)
	}
}

// With creates a new logger with the given fields
func With(fields ...interface{}) *zap.SugaredLogger {
	if log != nil {
		return log.With(fields...)
	}
	return nil
}

// InitTestLogger initializes a logger for testing with output to the provided writer
func InitTestLogger(level string) {
	cfg := zap.NewDevelopmentConfig()
	
	// Set log level based on input
	switch level {
	case "debug":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		cfg.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		cfg.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic("failed to initialize test logger: " + err.Error())
	}
	log = logger.Sugar()
}

// Sync flushes any buffered log entries
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}