package report

import (
	"firefly-task/drift"
)

// ReportFormat represents the different output formats supported
type ReportFormat int

const (
	// FormatJSON outputs the report in JSON format
	FormatJSON ReportFormat = iota
	// FormatYAML outputs the report in YAML format
	FormatYAML
	// FormatTable outputs the report in table format
	FormatTable
	// FormatConsole outputs the report with color coding for console
	FormatConsole
	// FormatCI outputs the report in CI/CD friendly format
	FormatCI
)

// String returns the string representation of ReportFormat
func (rf ReportFormat) String() string {
	switch rf {
	case FormatJSON:
		return "json"
	case FormatYAML:
		return "yaml"
	case FormatTable:
		return "table"
	case FormatConsole:
		return "console"
	case FormatCI:
		return "ci"
	default:
		return "unknown"
	}
}

// ReportConfig contains configuration options for report generation
type ReportConfig struct {
	// Format specifies the output format
	Format ReportFormat
	// OutputFile specifies the file to write the report to (optional)
	OutputFile string
	// IncludeTimestamp includes timestamp in the report
	IncludeTimestamp bool
	// Color enables color output for console
	Color bool
	// IncludeSummary includes summary statistics
	IncludeSummary bool
	// ColorOutput enables color coding for console output
	ColorOutput bool
	// FilterSeverity filters results by minimum severity level
	FilterSeverity drift.DriftSeverity
	// IncludeRecommendations includes actionable recommendations
	IncludeRecommendations bool
	// ShowProgressIndicator shows progress for long operations
	ShowProgressIndicator bool
}

// ReportGenerator defines the interface for generating drift reports
type ReportGenerator interface {
	// GenerateReport generates a report from drift results
	GenerateReport(results map[string]*drift.DriftResult, config ReportConfig) ([]byte, error)

	// GenerateJSONReport generates a JSON format report
	GenerateJSONReport(results map[string]*drift.DriftResult) ([]byte, error)

	// GenerateYAMLReport generates a YAML format report
	GenerateYAMLReport(results map[string]*drift.DriftResult) ([]byte, error)

	// GenerateTableReport generates a table format report
	GenerateTableReport(results map[string]*drift.DriftResult) (string, error)

	// GenerateConsoleReport generates a console report with color coding
	GenerateConsoleReport(results map[string]*drift.DriftResult) (string, error)

	// WriteToFile writes the report to a file
	WriteToFile(content []byte, filePath string) error
}

// ReportSummary contains summary statistics for the drift report
type ReportSummary struct {
	// TotalResources is the total number of resources checked
	TotalResources int `json:"total_resources"`
	// ResourcesWithDrift is the number of resources that have drift
	ResourcesWithDrift int `json:"resources_with_drift"`
	// TotalDifferences is the total number of differences found
	TotalDifferences int `json:"total_differences"`
	// SeverityCounts contains counts by severity level
	SeverityCounts map[string]int `json:"severity_counts"`
	// GenerationTime is when the report was generated
	GenerationTime string `json:"generation_time"`
	// OverallStatus indicates the overall drift status
	OverallStatus string `json:"overall_status"`
	// HighestSeverity indicates the highest severity level found
	HighestSeverity string `json:"highest_severity"`
}

// Recommendation represents an actionable recommendation
type Recommendation struct {
	// ResourceID identifies the resource this recommendation applies to
	ResourceID string `json:"resource_id"`
	// Severity indicates the importance of this recommendation
	Severity drift.DriftSeverity `json:"severity"`
	// Title is a short description of the recommendation
	Title string `json:"title"`
	// Description provides detailed information about the recommendation
	Description string `json:"description"`
	// Action suggests what action should be taken
	Action string `json:"action"`
	// Priority indicates the order in which recommendations should be addressed
	Priority int `json:"priority"`
}

// ReportData represents the complete report data structure
type ReportData struct {
	// Summary contains summary statistics
	Summary ReportSummary `json:"summary"`
	// Results contains the detailed drift results
	Results map[string]*drift.DriftResult `json:"results"`
	// Recommendations contains actionable recommendations
	Recommendations []Recommendation `json:"recommendations,omitempty"`
	// Metadata contains additional report metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	// Timestamp indicates when the report was generated
	Timestamp string `json:"timestamp,omitempty"`
}

// NewReportConfig creates a new ReportConfig with default values
func NewReportConfig() *ReportConfig {
	return &ReportConfig{
		Format:                 FormatJSON,
		IncludeTimestamp:       true,
		IncludeSummary:         true,
		ColorOutput:            true,
		FilterSeverity:         drift.SeverityNone,
		IncludeRecommendations: true,
		ShowProgressIndicator:  false,
	}
}

// WithFormat sets the output format
func (rc *ReportConfig) WithFormat(format ReportFormat) *ReportConfig {
	rc.Format = format
	return rc
}

// WithColor sets the color option for the report
func (rc *ReportConfig) WithColor(color bool) *ReportConfig {
	rc.Color = color
	return rc
}

// WithOutputFile sets the output file path
func (rc *ReportConfig) WithOutputFile(filePath string) *ReportConfig {
	rc.OutputFile = filePath
	return rc
}

// WithFilterSeverity sets the minimum severity level to include
func (rc *ReportConfig) WithFilterSeverity(severity drift.DriftSeverity) *ReportConfig {
	rc.FilterSeverity = severity
	return rc
}

// WithColorOutput enables or disables color output
func (rc *ReportConfig) WithColorOutput(enabled bool) *ReportConfig {
	rc.ColorOutput = enabled
	return rc
}