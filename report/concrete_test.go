package report

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"firefly-task/pkg/interfaces"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestNewConcreteReportGenerator(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewConcreteReportGenerator(tt.logger)
			assert.NotNil(t, generator)
			assert.IsType(t, &ConcreteReportGenerator{}, generator)
		})
	}
}

func TestNewConcreteReportWriter(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewConcreteReportWriter(tt.logger)
			assert.NotNil(t, writer)
			assert.IsType(t, &ConcreteReportWriter{}, writer)
		})
	}
}

func TestNewConcreteReportFormatter(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewConcreteReportFormatter(tt.logger)
			assert.NotNil(t, formatter)
			assert.IsType(t, &ConcreteReportFormatter{}, formatter)
		})
	}
}

func TestNewConcreteReportFilter(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewConcreteReportFilter(tt.logger)
			assert.NotNil(t, filter)
			assert.IsType(t, &ConcreteReportFilter{}, filter)
		})
	}
}

func TestNewConcreteReportFactory(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewConcreteReportFactory(tt.logger)
			assert.NotNil(t, factory)
			assert.IsType(t, &ConcreteReportFactory{}, factory)
		})
	}
}



func TestConcreteReportGenerator_GenerateJSONReport(t *testing.T) {
	logger := logrus.New()
	generator := NewConcreteReportGenerator(logger)
	driftResults := createTestDriftResults()

	tests := []struct {
		name          string
		driftResults  map[string]*interfaces.DriftResult
		options       map[string]interface{}
		expectedError string
	}{
		{
			name:         "successful JSON generation",
			driftResults: driftResults,
			options:      map[string]interface{}{},
		},
		{
			name:         "empty drift results",
			driftResults: map[string]*interfaces.DriftResult{},
			options:      map[string]interface{}{},
		},
		{
			name:         "nil drift results",
			driftResults: nil,
			options:      map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateJSONReport(context.Background(), tt.driftResults, tt.options)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify it's valid JSON
				var jsonData map[string]*interfaces.DriftResult
				err = json.Unmarshal(result, &jsonData)
				assert.NoError(t, err)
			}
		})
	}
}

func TestConcreteReportGenerator_GenerateYAMLReport(t *testing.T) {
	logger := logrus.New()
	generator := NewConcreteReportGenerator(logger)
	driftResults := createTestDriftResults()

	tests := []struct {
		name          string
		driftResults  map[string]*interfaces.DriftResult
		options       map[string]interface{}
		expectedError string
	}{
		{
			name:         "successful YAML generation",
			driftResults: driftResults,
			options:      map[string]interface{}{},
		},
		{
			name:         "empty drift results",
			driftResults: map[string]*interfaces.DriftResult{},
			options:      map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateYAMLReport(context.Background(), tt.driftResults, tt.options)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Verify it's valid YAML
				var yamlData map[string]*interfaces.DriftResult
				err = yaml.Unmarshal(result, &yamlData)
				assert.NoError(t, err)
			}
		})
	}
}

func TestConcreteReportGenerator_GenerateTableReport(t *testing.T) {
	logger := logrus.New()
	generator := NewConcreteReportGenerator(logger)
	driftResults := createTestDriftResults()

	result, err := generator.GenerateTableReport(context.Background(), driftResults, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	assert.Nil(t, result)
}

func TestConcreteReportGenerator_GenerateHTMLReport(t *testing.T) {
	logger := logrus.New()
	generator := NewConcreteReportGenerator(logger)
	driftResults := createTestDriftResults()

	result, err := generator.GenerateHTMLReport(context.Background(), driftResults, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	assert.Nil(t, result)
}

func TestConcreteReportGenerator_GenerateMarkdownReport(t *testing.T) {
	logger := logrus.New()
	generator := NewConcreteReportGenerator(logger)
	driftResults := createTestDriftResults()

	result, err := generator.GenerateMarkdownReport(context.Background(), driftResults, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	assert.Nil(t, result)
}

func TestConcreteReportGenerator_GenerateCustomReport(t *testing.T) {
	logger := logrus.New()
	generator := NewConcreteReportGenerator(logger)
	driftResults := createTestDriftResults()

	tests := []struct {
		name          string
		format        string
		expectedError string
	}{
		{
			name:   "JSON format",
			format: "json",
		},
		{
			name:   "YAML format",
			format: "yaml",
		},
		{
			name:   "YML format",
			format: "yml",
		},
		{
			name:          "table format (not implemented)",
			format:        "table",
			expectedError: "not implemented",
		},
		{
			name:          "HTML format (not implemented)",
			format:        "html",
			expectedError: "not implemented",
		},
		{
			name:          "markdown format (not implemented)",
			format:        "markdown",
			expectedError: "not implemented",
		},
		{
			name:          "unsupported format",
			format:        "unsupported",
			expectedError: "unsupported custom format: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateCustomReport(context.Background(), driftResults, tt.format, map[string]interface{}{})

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestConcreteReportWriter_WriteToFile(t *testing.T) {
	logger := logrus.New()
	writer := NewConcreteReportWriter(logger)
	content := []byte("test content")

	// Create a temporary file
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "test-report.txt")

	err := writer.WriteToFile(context.Background(), content, filePath, map[string]interface{}{})
	assert.NoError(t, err)

	// Verify file was created and content is correct
	writtenContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, writtenContent)
}

func TestConcreteReportWriter_WriteToFile_Error(t *testing.T) {
	logger := logrus.New()
	writer := NewConcreteReportWriter(logger)
	content := []byte("test content")

	// Try to write to an invalid path
	invalidPath := "/invalid/path/that/does/not/exist/file.txt"
	err := writer.WriteToFile(context.Background(), content, invalidPath, map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create file")
}

func TestConcreteReportWriter_WriteToConsole(t *testing.T) {
	logger := logrus.New()
	writer := NewConcreteReportWriter(logger)
	content := []byte("test console content")

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := writer.WriteToConsole(context.Background(), content, map[string]interface{}{})
	assert.NoError(t, err)

	// Restore stdout and read captured content
	w.Close()
	os.Stdout = oldStdout
	captured, _ := io.ReadAll(r)

	assert.Equal(t, content, captured)
}

func TestConcreteReportWriter_WriteToStream(t *testing.T) {
	logger := logrus.New()
	writer := NewConcreteReportWriter(logger)
	content := []byte("test stream content")

	// Create a buffer to act as a stream
	var buffer bytes.Buffer

	err := writer.WriteToStream(context.Background(), content, &buffer, map[string]interface{}{})
	assert.NoError(t, err)

	assert.Equal(t, content, buffer.Bytes())
}

func TestConcreteReportFormatter_FormatDriftResults(t *testing.T) {
	logger := logrus.New()
	formatter := NewConcreteReportFormatter(logger)
	driftResults := createTestDriftResults()

	result, err := formatter.FormatDriftResults(context.Background(), driftResults, "json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	assert.Nil(t, result)
}

func TestConcreteReportFormatter_FormatSummary(t *testing.T) {
	logger := logrus.New()
	formatter := NewConcreteReportFormatter(logger)
	summary := map[string]interface{}{"total": 2, "drifted": 1}

	result, err := formatter.FormatSummary(context.Background(), summary, "json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	assert.Nil(t, result)
}

func TestConcreteReportFormatter_FormatAttributeDrift(t *testing.T) {
	logger := logrus.New()
	formatter := NewConcreteReportFormatter(logger)
	attributeDrift := []*interfaces.DriftDetail{
		{
			Attribute:     "InstanceType",
			ExpectedValue: "t2.micro",
			ActualValue:   "t2.small",
			Severity:      interfaces.SeverityHigh,
		},
	}

	result, err := formatter.FormatAttributeDrift(context.Background(), attributeDrift, "json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
	assert.Nil(t, result)
}

func TestConcreteReportFilter_FilterByResourceType(t *testing.T) {
	logger := logrus.New()
	filter := NewConcreteReportFilter(logger)
	driftResults := createTestDriftResults()

	// Add a different resource type for testing
	driftResults["sg-12345"] = &interfaces.DriftResult{
		ResourceID:   "sg-12345",
		ResourceType: "SecurityGroup",
		IsDrifted:    true,
		Severity:     interfaces.SeverityMedium,
		DetectionTime: time.Now(),
		DriftDetails: []*interfaces.DriftDetail{},
	}

	tests := []struct {
		name         string
		resourceType string
		expectedCount int
	}{
		{
			name:         "filter EC2 instances",
			resourceType: "aws_instance",
			expectedCount: 2,
		},
		{
			name:         "filter load balancers",
			resourceType: "aws_lb",
			expectedCount: 1,
		},
		{
			name:         "filter non-existent type",
			resourceType: "NonExistent",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filter.FilterByResourceType(context.Background(), driftResults, tt.resourceType)
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all results have the correct resource type
			for _, driftResult := range result {
				assert.Equal(t, tt.resourceType, driftResult.ResourceType)
			}
		})
	}
}

func TestConcreteReportFilter_FilterByDriftStatus(t *testing.T) {
	logger := logrus.New()
	filter := NewConcreteReportFilter(logger)
	driftResults := createTestDriftResults()

	tests := []struct {
		name          string
		hasDrift      bool
		expectedCount int
	}{
		{
			name:          "filter drifted resources",
			hasDrift:      true,
			expectedCount: 3,
		},
		{
			name:          "filter non-drifted resources",
			hasDrift:      false,
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filter.FilterByDriftStatus(context.Background(), driftResults, tt.hasDrift)
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all results have the correct drift status
			for _, driftResult := range result {
				assert.Equal(t, tt.hasDrift, driftResult.IsDrifted)
			}
		})
	}
}

func TestConcreteReportFilter_FilterBySeverity(t *testing.T) {
	logger := logrus.New()
	filter := NewConcreteReportFilter(logger)
	driftResults := createTestDriftResults()

	tests := []struct {
		name          string
		severity      string
		expectedCount int
	}{
		{
			name:          "filter high severity",
			severity:      "high",
			expectedCount: 1,
		},
		{
			name:          "filter low severity",
			severity:      "low",
			expectedCount: 1,
		},
		{
			name:          "filter medium severity",
			severity:      "medium",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filter.FilterBySeverity(context.Background(), driftResults, tt.severity)
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify all results have the correct severity
			for _, driftResult := range result {
				assert.Equal(t, interfaces.SeverityLevel(tt.severity), driftResult.Severity)
			}
		})
	}
}

func TestConcreteReportFilter_FilterByAttributes(t *testing.T) {
	logger := logrus.New()
	filter := NewConcreteReportFilter(logger)
	driftResults := createTestDriftResults()

	tests := []struct {
		name          string
		attributes    []string
		expectedCount int
	}{
		{
			name:          "filter by InstanceType attribute",
			attributes:    []string{"instance_type"},
			expectedCount: 1,
		},
		{
			name:          "filter by non-existent attribute",
			attributes:    []string{"NonExistent"},
			expectedCount: 0,
		},
		{
			name:          "filter by multiple attributes",
			attributes:    []string{"instance_type", "security_groups"},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := filter.FilterByAttributes(context.Background(), driftResults, tt.attributes)
			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestConcreteReportFactory_CreateReportGenerator(t *testing.T) {
	logger := logrus.New()
	factory := NewConcreteReportFactory(logger)

	generator, err := factory.CreateReportGenerator(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, generator)
	assert.IsType(t, &ConcreteReportGenerator{}, generator)
}

func TestConcreteReportFactory_CreateReportWriter(t *testing.T) {
	logger := logrus.New()
	factory := NewConcreteReportFactory(logger)

	writer, err := factory.CreateReportWriter(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, writer)
	assert.IsType(t, &ConcreteReportWriter{}, writer)
}

func TestConcreteReportFactory_CreateReportFormatter(t *testing.T) {
	logger := logrus.New()
	factory := NewConcreteReportFactory(logger)

	formatter, err := factory.CreateReportFormatter(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, formatter)
	assert.IsType(t, &ConcreteReportFormatter{}, formatter)
}

func TestConcreteReportFactory_CreateReportFilter(t *testing.T) {
	logger := logrus.New()
	factory := NewConcreteReportFactory(logger)

	filter, err := factory.CreateReportFilter(context.Background(), map[string]interface{}{})
	assert.NoError(t, err)
	assert.NotNil(t, filter)
	assert.IsType(t, &ConcreteReportFilter{}, filter)
}

func TestConcreteReportFactory_ListSupportedFormats(t *testing.T) {
	logger := logrus.New()
	factory := NewConcreteReportFactory(logger)

	formats, err := factory.ListSupportedFormats(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, formats)
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "yaml")
	assert.Contains(t, formats, "table")
	assert.Contains(t, formats, "html")
	assert.Contains(t, formats, "markdown")
	assert.Len(t, formats, 5)
}