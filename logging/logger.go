package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"firefly-task/errors"
)

// Logger wraps logrus with additional functionality
type Logger struct {
	*logrus.Logger
	logFile *os.File
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"` // "text" or "json"
	Output     string `json:"output"` // "console", "file", or "both"
	LogFile    string `json:"log_file,omitempty"`
	MaxSize    int64  `json:"max_size,omitempty"` // Max log file size in MB
	MaxBackups int    `json:"max_backups,omitempty"`
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		Level:      "info",
		Format:     "text",
		Output:     "console",
		MaxSize:    10, // 10MB
		MaxBackups: 5,
	}
}

// GetDefaultLogPath returns the default log file path for Windows
func GetDefaultLogPath() string {
	// Use Windows-compatible paths
	appData := os.Getenv("APPDATA")
	if appData == "" {
		// Fallback to current directory if APPDATA is not available
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, "logs", "firefly.log")
		}
		return "firefly.log"
	}
	
	logDir := filepath.Join(appData, "Firefly")
	return filepath.Join(logDir, "firefly.log")
}

// NewLogger creates a new logger instance
func NewLogger(config *LogConfig) (*Logger, error) {
	logger := logrus.New()
	
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		return nil, errors.NewConfigurationError(fmt.Sprintf("invalid log level: %s", config.Level))
	}
	logger.SetLevel(level)
	
	// Set formatter
	switch strings.ToLower(config.Format) {
	case "json":
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	case "text":
		logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     false, // Disable colors for Windows compatibility
		})
	default:
		return nil, errors.NewConfigurationError(fmt.Sprintf("invalid log format: %s", config.Format))
	}
	
	wrapper := &Logger{Logger: logger}
	
	// Configure output
	if err := wrapper.configureOutput(config); err != nil {
		return nil, err
	}
	
	return wrapper, nil
}

// configureOutput sets up log output destinations
func (l *Logger) configureOutput(config *LogConfig) error {
	switch strings.ToLower(config.Output) {
	case "console":
		l.SetOutput(os.Stdout)
	case "file":
		logFile := config.LogFile
		if logFile == "" {
			logFile = GetDefaultLogPath()
		}
		
		if err := l.setupFileOutput(logFile); err != nil {
			return err
		}
	case "both":
		logFile := config.LogFile
		if logFile == "" {
			logFile = GetDefaultLogPath()
		}
		
		if err := l.setupFileOutput(logFile); err != nil {
			return err
		}
		
		// Create multi-writer for both console and file
		multiWriter := io.MultiWriter(os.Stdout, l.logFile)
		l.SetOutput(multiWriter)
	default:
		return errors.NewConfigurationError(fmt.Sprintf("invalid log output: %s", config.Output))
	}
	
	return nil
}

// setupFileOutput creates and opens the log file
func (l *Logger) setupFileOutput(logFile string) error {
	// Ensure directory exists
	logDir := filepath.Dir(logFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return errors.NewFileSystemError("create_log_directory", 
			fmt.Sprintf("failed to create log directory %s: %v", logDir, err)).WithCause(err)
	}
	
	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return errors.NewFileSystemError("open_log_file", 
			fmt.Sprintf("failed to open log file %s: %v", logFile, err)).WithCause(err)
	}
	
	l.logFile = file
	l.SetOutput(file)
	
	return nil
}

// Close closes the log file if it's open
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// LogError logs a FireflyError with appropriate context
func (l *Logger) LogError(err error) {
	if fireflyErr, ok := err.(*errors.FireflyError); ok {
		fields := logrus.Fields{
			"error_type": fireflyErr.Type,
			"operation":  fireflyErr.Operation,
		}
		
		if fireflyErr.ResourceID != "" {
			fields["resource_id"] = fireflyErr.ResourceID
		}
		
		if fireflyErr.Details != "" {
			fields["details"] = fireflyErr.Details
		}
		
		if len(fireflyErr.Suggestions) > 0 {
			fields["suggestions"] = fireflyErr.Suggestions
		}
		
		l.WithFields(fields).Error(fireflyErr.Message)
		
		// Log suggestions separately for better visibility in text format
		if len(fireflyErr.Suggestions) > 0 {
			for _, suggestion := range fireflyErr.Suggestions {
				l.Infof("Suggestion: %s", suggestion)
			}
		}
	} else {
		l.Error(err)
	}
}

// LogOperation logs the start and end of an operation
func (l *Logger) LogOperation(operation string, fn func() error) error {
	l.WithField("operation", operation).Info("Starting operation")
	start := time.Now()
	
	err := fn()
	duration := time.Since(start)
	
	fields := logrus.Fields{
		"operation": operation,
		"duration":  duration.String(),
	}
	
	if err != nil {
		l.WithFields(fields).Error("Operation failed")
		l.LogError(err)
	} else {
		l.WithFields(fields).Info("Operation completed successfully")
	}
	
	return err
}

// AddCallerInfo adds caller information to log entries
func (l *Logger) AddCallerInfo() *logrus.Entry {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return l.WithField("caller", "unknown")
	}
	
	// Get just the filename, not the full path
	filename := filepath.Base(file)
	return l.WithField("caller", fmt.Sprintf("%s:%d", filename, line))
}

// SetVerbose configures logging based on verbose flag
func (l *Logger) SetVerbose(verbose bool) {
	if verbose {
		l.SetLevel(logrus.DebugLevel)
		l.Debug("Verbose logging enabled")
	} else {
		l.SetLevel(logrus.InfoLevel)
	}
}

// GetLogStats returns basic logging statistics
func (l *Logger) GetLogStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["level"] = l.GetLevel().String()
	
	if l.logFile != nil {
		if stat, err := l.logFile.Stat(); err == nil {
			stats["log_file_size"] = stat.Size()
			stats["log_file_path"] = l.logFile.Name()
		}
	}
	
	return stats
}