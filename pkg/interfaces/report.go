package interfaces

import (
	"io"
)



// ReportGenerator defines the interface for generating drift detection reports
type ReportGenerator interface {
	// GenerateJSONReport generates a JSON format report from drift results
	GenerateJSONReport(results map[string]*DriftResult) ([]byte, error)

	// GenerateYAMLReport generates a YAML format report from drift results
	GenerateYAMLReport(results map[string]*DriftResult) ([]byte, error)

	// GenerateTableReport generates a human-readable table format report
	GenerateTableReport(results map[string]*DriftResult) (string, error)

	// GenerateHTMLReport generates an HTML format report with styling
	GenerateHTMLReport(results map[string]*DriftResult) ([]byte, error)

	// GenerateMarkdownReport generates a Markdown format report
	GenerateMarkdownReport(results map[string]*DriftResult) ([]byte, error)

	// WriteReport writes a report to an io.Writer
	WriteReport(results map[string]*DriftResult, format string, writer io.Writer) error
}

// ReportWriter defines the interface for writing reports to different destinations
type ReportWriter interface {
	// WriteToFile writes the report content to a file
	WriteToFile(content []byte, filePath string) error

	// WriteToConsole writes the report content to console/stdout
	WriteToConsole(content []byte) error

	// WriteToStream writes the report content to an io.Writer
	WriteToStream(content []byte, writer io.Writer) error
}

// ReportFormatter defines the interface for formatting report data
type ReportFormatter interface {
	// FormatDriftResult formats a single drift result for display
	FormatDriftResult(result *DriftResult) (string, error)

	// FormatDriftSummary formats a summary of all drift results
	FormatDriftSummary(results map[string]*DriftResult) (string, error)

	// FormatAttributeDrift formats attribute-level drift information
	FormatAttributeDrift(attributeDrift *DriftDetail) (string, error)

	// GetSupportedFormats returns the list of formats this formatter supports
	GetSupportedFormats() []string
}

// ReportFilter defines the interface for filtering report data
type ReportFilter interface {
	// FilterByResourceType filters results by resource type
	FilterByResourceType(results map[string]*DriftResult, resourceType string) map[string]*DriftResult

	// FilterByDriftStatus filters results by drift status (has drift, no drift)
	FilterByDriftStatus(results map[string]*DriftResult, hasDrift bool) map[string]*DriftResult

	// FilterBySeverity filters results by severity level
	FilterBySeverity(results map[string]*DriftResult, severity SeverityLevel) map[string]*DriftResult

	// FilterByAttributes filters results that have drift in specific attributes
	FilterByAttributes(results map[string]*DriftResult, attributes []string) map[string]*DriftResult
}

// ReportFactory defines the interface for creating different types of reports
type ReportFactory interface {
	// CreateGenerator creates a report generator for the specified format
	CreateGenerator(format string) (ReportGenerator, error)

	// CreateWriter creates a report writer for the specified destination
	CreateWriter(destination string) (ReportWriter, error)

	// CreateFormatter creates a report formatter for the specified format
	CreateFormatter(format string) (ReportFormatter, error)

	// CreateFilter creates a report filter with the specified criteria
	CreateFilter(criteria map[string]interface{}) (ReportFilter, error)

	// GetSupportedFormats returns all supported report formats
	GetSupportedFormats() []string
}
