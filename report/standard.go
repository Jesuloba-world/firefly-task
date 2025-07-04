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

	"firefly-task/drift"
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
func (srg *StandardReportGenerator) GenerateReport(results map[string]*drift.DriftResult, config ReportConfig) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	// Apply filters
	filteredResults, err := srg.filterResults(results, config.FilterSeverity)
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
func (srg *StandardReportGenerator) GenerateJSONReport(results map[string]*drift.DriftResult) ([]byte, error) {
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
func (srg *StandardReportGenerator) GenerateYAMLReport(results map[string]*drift.DriftResult) ([]byte, error) {
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
func (srg *StandardReportGenerator) GenerateTableReport(results map[string]*drift.DriftResult) (string, error) {
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
		severity := strings.ToUpper(result.OverallSeverity.String())
		
		// Format differences
		var differences string
		if result.HasDrift {
			var diffNames []string
			for _, diff := range result.Differences {
				diffNames = append(diffNames, diff.AttributeName)
			}
			differences = strings.Join(diffNames, ", ")
			if len(differences) > 47 {
				differences = differences[:47] + "..."
			}
		} else {
			differences = "None"
		}

		builder.WriteString(fmt.Sprintf(rowFormat, result.ResourceID, driftStatus, severity, differences))
	}

	builder.WriteString(strings.Repeat("=", 120) + "\n")

	return builder.String(), nil
}

// GenerateConsoleReport generates a console report with color formatting
func (srg *StandardReportGenerator) GenerateConsoleReport(results map[string]*drift.DriftResult) (string, error) {
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
func (srg *StandardReportGenerator) buildReportData(results map[string]*drift.DriftResult) *ReportData {
	summary := srg.generateSummary(results)
	recommendations := srg.generateRecommendations(results)

	reportData := &ReportData{
		Summary:         summary,
		Results:         results,
		Recommendations: recommendations,
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
func (srg *StandardReportGenerator) generateSummary(results map[string]*drift.DriftResult) ReportSummary {
	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0
	severityCounts := make(map[string]int)
	highestSeverity := drift.SeverityNone

	for _, result := range results {
		if result.HasDrift {
			resourcesWithDrift++
			totalDifferences += len(result.Differences)
			
			// Count individual difference severities and track highest
			for _, diff := range result.Differences {
				severityStr := diff.Severity.String()
				severityCounts[severityStr]++
				
				// Track highest severity
				if diff.Severity > highestSeverity {
					highestSeverity = diff.Severity
				}
			}
		} else {
			// Count resources without drift as "low" severity
			severityStr := result.OverallSeverity.String()
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
		HighestSeverity:    highestSeverity.String(),
	}
}

// generateRecommendations creates actionable recommendations based on drift results
func (srg *StandardReportGenerator) generateRecommendations(results map[string]*drift.DriftResult) []Recommendation {
	var recommendations []Recommendation
	priority := 1

	for _, result := range results {
		if !result.HasDrift {
			continue
		}

		// Generate recommendations based on severity and type of drift
		for _, diff := range result.Differences {
			recommendation := Recommendation{
				ResourceID:  result.ResourceID,
				Severity:    diff.Severity,
				Title:       fmt.Sprintf("Address %s drift in %s", diff.AttributeName, result.ResourceID),
				Description: fmt.Sprintf("The %s attribute has drifted from expected value '%v' to '%v'", diff.AttributeName, diff.ExpectedValue, diff.ActualValue),
				Action:      srg.generateActionRecommendation(diff),
				Priority:    priority,
			}
			recommendations = append(recommendations, recommendation)
			priority++
		}
	}

	// Sort recommendations by severity (highest first) then by priority
	sort.Slice(recommendations, func(i, j int) bool {
		if recommendations[i].Severity != recommendations[j].Severity {
			return recommendations[i].Severity > recommendations[j].Severity
		}
		return recommendations[i].Priority < recommendations[j].Priority
	})

	return recommendations
}

// generateActionRecommendation generates specific action recommendations based on the difference
func (srg *StandardReportGenerator) generateActionRecommendation(diff drift.AttributeDifference) string {
	switch diff.Severity {
	case drift.SeverityCritical:
		return "IMMEDIATE ACTION REQUIRED: Update Terraform configuration and apply changes"
	case drift.SeverityHigh:
		return "Update Terraform configuration to match actual state or revert AWS changes"
	case drift.SeverityMedium:
		return "Review and align Terraform configuration with actual AWS state"
	case drift.SeverityLow:
		return "Consider updating Terraform configuration during next maintenance window"
	default:
		return "Review configuration alignment"
	}
}

// filterResults filters results based on minimum severity level
func (srg *StandardReportGenerator) filterResults(results map[string]*drift.DriftResult, minSeverity drift.DriftSeverity) (map[string]*drift.DriftResult, error) {
	if minSeverity == drift.SeverityNone {
		return results, nil
	}

	filteredResults := make(map[string]*drift.DriftResult)
	for resourceID, result := range results {
		if result.OverallSeverity >= minSeverity {
			filteredResults[resourceID] = result
		}
	}

	return filteredResults, nil
}

// getDriftStatus returns a human-readable drift status
func (srg *StandardReportGenerator) getDriftStatus(result *drift.DriftResult) string {
	if !result.HasDrift {
		return "NO DRIFT"
	}
	return fmt.Sprintf("DRIFT DETECTED (%d differences)", len(result.Differences))
}
