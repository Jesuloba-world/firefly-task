package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"firefly-task/pkg/interfaces"
)

// StandardReportGenerator implements the ReportGenerator interface
type StandardReportGenerator struct {
	config *ReportConfig
}

// NewStandardReportGenerator creates a new StandardReportGenerator
func NewStandardReportGenerator() *StandardReportGenerator {
	return &StandardReportGenerator{
		config: NewReportConfig(),
	}
}

// WithConfig applies configuration to the generator
func (srg *StandardReportGenerator) WithConfig(config *ReportConfig) ReportGenerator {
	srg.config = config
	return srg
}

// GenerateReport generates a report based on the configured format
func (srg *StandardReportGenerator) GenerateReport(results map[string]*interfaces.DriftResult, config ReportConfig) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	// Apply filters
	filteredResults, err := srg.filterResults(results, interfaces.SeverityLevel(config.FilterSeverity))
	if err != nil {
		return nil, WrapError(ErrorTypeFilterError, "failed to filter results", err)
	}

	switch config.Format {
	case FormatJSON:
		return srg.GenerateJSONReport(filteredResults)
	case FormatYAML:
		return srg.GenerateYAMLReport(filteredResults)
	case FormatTable:
		tableReport, err := srg.GenerateTableReport(filteredResults)
		if err != nil {
			return nil, err
		}
		return []byte(tableReport), nil
	case FormatConsole:
		consoleReport, err := srg.GenerateConsoleReport(filteredResults)
		if err != nil {
			return nil, err
		}
		return []byte(consoleReport), nil
	default:
		return nil, NewReportError(ErrorTypeInvalidFormat, fmt.Sprintf("unsupported format: %s", config.Format.String()))
	}
}

// GenerateJSONReport generates a JSON format report
func (srg *StandardReportGenerator) GenerateJSONReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	reportData := srg.buildReportData(results)

	jsonData, err := json.MarshalIndent(reportData, "", "  ")
	if err != nil {
		return nil, WrapError(ErrorTypeMarshaling, "failed to marshal JSON", err)
	}

	return jsonData, nil
}

// GenerateYAMLReport generates a YAML format report
func (srg *StandardReportGenerator) GenerateYAMLReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	reportData := srg.buildReportData(results)

	yamlData, err := yaml.Marshal(reportData)
	if err != nil {
		return nil, WrapError(ErrorTypeMarshaling, "failed to marshal YAML", err)
	}

	return yamlData, nil
}

// GenerateTableReport generates a table format report
func (srg *StandardReportGenerator) GenerateTableReport(results map[string]*interfaces.DriftResult) (string, error) {
	if results == nil {
		return "", NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	var builder strings.Builder

	// Header
	builder.WriteString("\n=== DRIFT DETECTION REPORT ===\n")
	builder.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Summary
	summary := srg.generateSummary(results)
	builder.WriteString("SUMMARY:\n")
	builder.WriteString(fmt.Sprintf("  Total Resources: %d\n", summary.TotalResources))
	builder.WriteString(fmt.Sprintf("  Resources with Drift: %d\n", summary.ResourcesWithDrift))
	builder.WriteString(fmt.Sprintf("  Total Differences: %d\n", summary.TotalDifferences))
	builder.WriteString(fmt.Sprintf("  Overall Status: %s\n\n", summary.OverallStatus))

	// Severity breakdown
	if len(summary.SeverityCounts) > 0 {
		builder.WriteString("SEVERITY BREAKDOWN:\n")
		for severity, count := range summary.SeverityCounts {
			if count > 0 {
				builder.WriteString(fmt.Sprintf("  %s: %d\n", strings.ToUpper(severity), count))
			}
		}
		builder.WriteString("\n")
	}

	// Table format
	builder.WriteString("RESOURCE TABLE:\n")
	builder.WriteString(strings.Repeat("=", 120) + "\n")

	// Table headers
	headerFormat := "%-30s | %-15s | %-10s | %-50s\n"
	builder.WriteString(fmt.Sprintf(headerFormat, "Resource ID", "Drift Status", "Severity", "Differences"))
	builder.WriteString(strings.Repeat("-", 120) + "\n")

	// Sort results by resource ID for consistent output
	var resourceIDs []string
	for resourceID := range results {
		resourceIDs = append(resourceIDs, resourceID)
	}
	sort.Strings(resourceIDs)

	// Table rows
	rowFormat := "%-30s | %-15s | %-10s | %-50s\n"
	for _, resourceID := range resourceIDs {
		result := results[resourceID]
		driftStatus := srg.getDriftStatus(result)
		severity := strings.ToUpper(string(result.Severity))
		
		// Format differences
		var differences string
		if result.IsDrifted {
			var diffNames []string
			for _, diff := range result.DriftDetails {
				diffNames = append(diffNames, diff.Attribute)
			}
			differences = strings.Join(diffNames, ", ")
			if len(differences) > 47 {
				differences = differences[:47] + "..."
			}
		} else {
			differences = "None"
		}

		builder.WriteString(fmt.Sprintf(rowFormat, resourceID, driftStatus, severity, differences))
	}

	builder.WriteString(strings.Repeat("=", 120) + "\n")

	return builder.String(), nil
}

// GenerateConsoleReport generates a console report with color formatting
func (srg *StandardReportGenerator) GenerateConsoleReport(results map[string]*interfaces.DriftResult) (string, error) {
	if results == nil {
		return "", NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	// Use console generator for color support
	consoleGen := NewConsoleReportGenerator()
	if srg.config != nil {
		consoleGen = consoleGen.WithConfig(srg.config).(*ConsoleReportGenerator)
	}
	return consoleGen.GenerateConsoleReport(results)
}

// WriteToFile writes the report content to a file
func (srg *StandardReportGenerator) WriteToFile(content []byte, filePath string) error {
	if filePath == "" {
		return NewReportError(ErrorTypeInvalidInput, "file path cannot be empty")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return WrapError(ErrorTypeFileWrite, "failed to create directory", err)
	}

	// Write file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return WrapError(ErrorTypeFileWrite, "failed to write file", err)
	}

	return nil
}

// Helper methods

// buildReportData creates the complete report data structure
func (srg *StandardReportGenerator) buildReportData(results map[string]*interfaces.DriftResult) *ReportData {
	summary := srg.generateSummary(results)


	reportData := &ReportData{
		Summary:         summary,
		Results:         results,

		Metadata: map[string]interface{}{
			"generator_version": "1.0.0",
			"report_format":     "standard",
		},
	}

	// Include timestamp if configured
	if srg.config != nil && srg.config.IncludeTimestamp {
		reportData.Timestamp = time.Now().Format(time.RFC3339)
	}

	return reportData
}

// generateSummary creates summary statistics from the results
func (srg *StandardReportGenerator) generateSummary(results map[string]*interfaces.DriftResult) ReportSummary {
	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0
	severityCounts := make(map[string]int)
	highestSeverity := interfaces.SeverityLow

	for _, result := range results {
		if result.IsDrifted {
			resourcesWithDrift++
			totalDifferences += len(result.DriftDetails)
			
			// Count individual difference severities and track highest
			for _, detail := range result.DriftDetails {
				severityStr := string(detail.Severity)
				severityCounts[severityStr]++
				
				// Track highest severity
				if detail.Severity == interfaces.SeverityHigh {
					highestSeverity = interfaces.SeverityHigh
				} else if detail.Severity == interfaces.SeverityMedium && highestSeverity != interfaces.SeverityHigh {
					highestSeverity = interfaces.SeverityMedium
				}
			}
		} else {
			// Count resources without drift as "low" severity
			severityStr := string(result.Severity)
			severityCounts[severityStr]++
		}
	}

	overallStatus := "CLEAN"
	if resourcesWithDrift > 0 {
		overallStatus = "DRIFT_DETECTED"
	}

	return ReportSummary{
		TotalResources:     totalResources,
		ResourcesWithDrift: resourcesWithDrift,
		TotalDifferences:   totalDifferences,
		SeverityCounts:     severityCounts,
		GenerationTime:     time.Now().Format(time.RFC3339),
		OverallStatus:      overallStatus,
		HighestSeverity:    string(highestSeverity),
	}
}





// getSeverityOrder returns the numeric order of severity levels for comparison
func getSeverityOrder(severity interfaces.SeverityLevel) int {
	switch severity {
	case interfaces.SeverityLow:
		return 1
	case interfaces.SeverityMedium:
		return 2
	case interfaces.SeverityHigh:
		return 3
	case interfaces.SeverityCritical:
		return 4
	default:
		return 0
	}
}

// filterResults filters results based on minimum severity level
func (srg *StandardReportGenerator) filterResults(results map[string]*interfaces.DriftResult, minSeverity interfaces.SeverityLevel) (map[string]*interfaces.DriftResult, error) {
	if minSeverity == interfaces.SeverityLow {
		return results, nil
	}

	filteredResults := make(map[string]*interfaces.DriftResult)
	minSeverityOrder := getSeverityOrder(minSeverity)
	
	for resourceID, result := range results {
		// Use numeric comparison for severity filtering
		if getSeverityOrder(result.Severity) >= minSeverityOrder {
			filteredResults[resourceID] = result
		}
	}

	return filteredResults, nil
}

// getDriftStatus returns a human-readable drift status
func (srg *StandardReportGenerator) getDriftStatus(result *interfaces.DriftResult) string {
	if !result.IsDrifted {
		return "NO DRIFT"
	}
	return fmt.Sprintf("DRIFT DETECTED (%d differences)", len(result.DriftDetails))
}
