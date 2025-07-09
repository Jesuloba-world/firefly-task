package report

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"firefly-task/pkg/interfaces"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// ConcreteReportGenerator implements the interfaces.ReportGenerator interface
type ConcreteReportGenerator struct {
	logger *logrus.Logger
}

// ConcreteReportWriter implements the interfaces.ReportWriter interface
type ConcreteReportWriter struct {
	logger *logrus.Logger
}

// ConcreteReportFormatter implements the interfaces.ReportFormatter interface
type ConcreteReportFormatter struct {
	logger *logrus.Logger
}

// ConcreteReportFilter implements the interfaces.ReportFilter interface
type ConcreteReportFilter struct {
	logger *logrus.Logger
}

// ConcreteReportFactory implements the interfaces.ReportFactory interface
type ConcreteReportFactory struct {
	logger *logrus.Logger
}

// NewConcreteReportGenerator creates a new concrete report generator
func NewConcreteReportGenerator(logger *logrus.Logger) *ConcreteReportGenerator {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteReportGenerator{
		logger: logger,
	}
}

// NewConcreteReportWriter creates a new concrete report writer
func NewConcreteReportWriter(logger *logrus.Logger) *ConcreteReportWriter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteReportWriter{
		logger: logger,
	}
}

// NewConcreteReportFormatter creates a new concrete report formatter
func NewConcreteReportFormatter(logger *logrus.Logger) *ConcreteReportFormatter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteReportFormatter{
		logger: logger,
	}
}

// NewConcreteReportFilter creates a new concrete report filter
func NewConcreteReportFilter(logger *logrus.Logger) *ConcreteReportFilter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteReportFilter{
		logger: logger,
	}
}

// NewConcreteReportFactory creates a new concrete report factory
func NewConcreteReportFactory(logger *logrus.Logger) *ConcreteReportFactory {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteReportFactory{
		logger: logger,
	}
}

// ReportGenerator implementation methods

// GenerateJSONReport generates a JSON format report
func (g *ConcreteReportGenerator) GenerateJSONReport(ctx context.Context, driftResults map[string]*interfaces.DriftResult, options map[string]interface{}) ([]byte, error) {
	g.logger.Debugf("ConcreteReportGenerator: Generating JSON report for %d drift results", len(driftResults))
	
	if driftResults == nil {
		driftResults = make(map[string]*interfaces.DriftResult)
	}
	
	jsonData, err := json.MarshalIndent(driftResults, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal drift results to JSON: %w", err)
	}
	
	return jsonData, nil
}

// GenerateYAMLReport generates a YAML format report
func (g *ConcreteReportGenerator) GenerateYAMLReport(ctx context.Context, driftResults map[string]*interfaces.DriftResult, options map[string]interface{}) ([]byte, error) {
	g.logger.Debugf("ConcreteReportGenerator: Generating YAML report for %d drift results", len(driftResults))
	
	if driftResults == nil {
		driftResults = make(map[string]*interfaces.DriftResult)
	}
	
	yamlData, err := yaml.Marshal(driftResults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal drift results to YAML: %w", err)
	}
	
	return yamlData, nil
}

// GenerateTableReport generates a table format report
func (g *ConcreteReportGenerator) GenerateTableReport(ctx context.Context, driftResults map[string]*interfaces.DriftResult, options map[string]interface{}) ([]byte, error) {
	g.logger.Debugf("ConcreteReportGenerator: Generating table report for %d drift results", len(driftResults))
	// Implement table generation logic here
	return nil, fmt.Errorf("not implemented")
}

// GenerateHTMLReport generates an HTML format report
func (g *ConcreteReportGenerator) GenerateHTMLReport(ctx context.Context, driftResults map[string]*interfaces.DriftResult, options map[string]interface{}) ([]byte, error) {
	g.logger.Debugf("ConcreteReportGenerator: Generating HTML report for %d drift results", len(driftResults))
	// Implement HTML generation logic here
	return nil, fmt.Errorf("not implemented")
}

// GenerateMarkdownReport generates a Markdown format report
func (g *ConcreteReportGenerator) GenerateMarkdownReport(ctx context.Context, driftResults map[string]*interfaces.DriftResult, options map[string]interface{}) ([]byte, error) {
	g.logger.Debugf("ConcreteReportGenerator: Generating Markdown report for %d drift results", len(driftResults))
	// Implement Markdown generation logic here
	return nil, fmt.Errorf("not implemented")
}

// GenerateCustomReport generates a custom format report
func (g *ConcreteReportGenerator) GenerateCustomReport(ctx context.Context, driftResults map[string]*interfaces.DriftResult, format string, options map[string]interface{}) ([]byte, error) {
	g.logger.Debugf("ConcreteReportGenerator: Generating custom %s report for %d drift results", format, len(driftResults))
	
	switch strings.ToLower(format) {
	case "json":
		return g.GenerateJSONReport(ctx, driftResults, options)
	case "yaml", "yml":
		return g.GenerateYAMLReport(ctx, driftResults, options)
	case "table":
		return g.GenerateTableReport(ctx, driftResults, options)
	case "html":
		return g.GenerateHTMLReport(ctx, driftResults, options)
	case "markdown", "md":
		return g.GenerateMarkdownReport(ctx, driftResults, options)
	default:
		return nil, fmt.Errorf("unsupported custom format: %s", format)
	}
}

// ReportWriter implementation methods

// WriteToFile writes report content to a file
func (w *ConcreteReportWriter) WriteToFile(ctx context.Context, content []byte, filePath string, options map[string]interface{}) error {
	w.logger.Debugf("ConcreteReportWriter: Writing %d bytes to file %s", len(content), filePath)
	
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()
	
	_, err = file.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content to file %s: %w", filePath, err)
	}
	
	return nil
}

// WriteToConsole writes report content to console
func (w *ConcreteReportWriter) WriteToConsole(ctx context.Context, content []byte, options map[string]interface{}) error {
	w.logger.Debug("ConcreteReportWriter: Writing content to console")
	
	_, err := os.Stdout.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content to console: %w", err)
	}
	
	return nil
}

// WriteToStream writes report content to a stream
func (w *ConcreteReportWriter) WriteToStream(ctx context.Context, content []byte, stream io.Writer, options map[string]interface{}) error {
	w.logger.Debugf("ConcreteReportWriter: Writing %d bytes to stream", len(content))
	
	_, err := stream.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write content to stream: %w", err)
	}
	
	return nil
}

// ReportFormatter implementation methods

// FormatDriftResults formats drift results for display
func (f *ConcreteReportFormatter) FormatDriftResults(ctx context.Context, driftResults map[string]*interfaces.DriftResult, format string) ([]byte, error) {
	f.logger.Debugf("ConcreteReportFormatter: Formatting %d drift results as %s", len(driftResults), format)
	// Implement formatting logic here
	return nil, fmt.Errorf("not implemented")
}

// FormatSummary formats a summary of drift results
func (f *ConcreteReportFormatter) FormatSummary(ctx context.Context, summary interface{}, format string) ([]byte, error) {
	f.logger.Debugf("ConcreteReportFormatter: Formatting summary as %s", format)
	// Implement summary formatting logic here
	return nil, fmt.Errorf("not implemented")
}

// FormatAttributeDrift formats attribute drift information
func (f *ConcreteReportFormatter) FormatAttributeDrift(ctx context.Context, attributeDrift []*interfaces.DriftDetail, format string) ([]byte, error) {
	f.logger.Debugf("ConcreteReportFormatter: Formatting %d attribute drifts as %s", len(attributeDrift), format)
	// Implement attribute drift formatting logic here
	return nil, fmt.Errorf("not implemented")
}

// ReportFilter implementation methods

// FilterByResourceType filters drift results by resource type
func (rf *ConcreteReportFilter) FilterByResourceType(ctx context.Context, driftResults map[string]*interfaces.DriftResult, resourceType string) (map[string]*interfaces.DriftResult, error) {
	rf.logger.Debugf("ConcreteReportFilter: Filtering %d drift results by resource type %s", len(driftResults), resourceType)
	
	filteredResults := make(map[string]*interfaces.DriftResult)
	
	for id, result := range driftResults {
		if result.ResourceType == resourceType {
			filteredResults[id] = result
		}
	}
	
	return filteredResults, nil
}

// FilterByDriftStatus filters drift results by drift status
func (rf *ConcreteReportFilter) FilterByDriftStatus(ctx context.Context, driftResults map[string]*interfaces.DriftResult, hasDrift bool) (map[string]*interfaces.DriftResult, error) {
	rf.logger.Debugf("ConcreteReportFilter: Filtering %d drift results by drift status %t", len(driftResults), hasDrift)
	
	filteredResults := make(map[string]*interfaces.DriftResult)
	
	for id, result := range driftResults {
		if result.IsDrifted == hasDrift {
			filteredResults[id] = result
		}
	}
	
	return filteredResults, nil
}

// FilterBySeverity filters drift results by severity level
func (rf *ConcreteReportFilter) FilterBySeverity(ctx context.Context, driftResults map[string]*interfaces.DriftResult, severity string) (map[string]*interfaces.DriftResult, error) {
	rf.logger.Debugf("ConcreteReportFilter: Filtering %d drift results by severity %s", len(driftResults), severity)
	
	filteredResults := make(map[string]*interfaces.DriftResult)
	
	sev := interfaces.SeverityLevel(severity)

	for id, result := range driftResults {
		if result.Severity == sev {
			filteredResults[id] = result
		}
	}
	
	return filteredResults, nil
}

// FilterByAttributes filters drift results by specific attributes
func (rf *ConcreteReportFilter) FilterByAttributes(ctx context.Context, driftResults map[string]*interfaces.DriftResult, attributes []string) (map[string]*interfaces.DriftResult, error) {
	rf.logger.Debugf("ConcreteReportFilter: Filtering %d drift results by %d attributes", len(driftResults), len(attributes))
	
	filteredResults := make(map[string]*interfaces.DriftResult)
	
	for id, result := range driftResults {
		// Check if any of the specified attributes have drift
		hasMatchingAttribute := false
		for _, attrDrift := range result.DriftDetails {
			for _, targetAttr := range attributes {
				if attrDrift.Attribute == targetAttr {
					hasMatchingAttribute = true
					break
				}
			}
			if hasMatchingAttribute {
				break
			}
		}
		
		if hasMatchingAttribute {
				filteredResults[id] = result
			}
	}
	
	return filteredResults, nil
}

// ReportFactory implementation methods

// CreateReportGenerator creates a new report generator instance
func (factory *ConcreteReportFactory) CreateReportGenerator(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	factory.logger.Debug("ConcreteReportFactory: Creating report generator")
	return NewConcreteReportGenerator(factory.logger), nil
}

// CreateReportWriter creates a new report writer instance
func (factory *ConcreteReportFactory) CreateReportWriter(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	factory.logger.Debug("ConcreteReportFactory: Creating report writer")
	return NewConcreteReportWriter(factory.logger), nil
}

// CreateReportFormatter creates a new report formatter instance
func (factory *ConcreteReportFactory) CreateReportFormatter(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	factory.logger.Debug("ConcreteReportFactory: Creating report formatter")
	return NewConcreteReportFormatter(factory.logger), nil
}

// CreateReportFilter creates a new report filter instance
func (factory *ConcreteReportFactory) CreateReportFilter(ctx context.Context, config map[string]interface{}) (interface{}, error) {
	factory.logger.Debug("ConcreteReportFactory: Creating report filter")
	return NewConcreteReportFilter(factory.logger), nil
}

// ListSupportedFormats returns a list of supported report formats
func (factory *ConcreteReportFactory) ListSupportedFormats(ctx context.Context) ([]string, error) {
	factory.logger.Debug("ConcreteReportFactory: Listing supported formats")
	return []string{"json", "yaml", "table", "html", "markdown"}, nil
}