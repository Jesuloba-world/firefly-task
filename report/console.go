package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"firefly-task/pkg/interfaces"
)

// ANSI color codes for console output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// ConsoleReportGenerator implements the ReportGenerator interface for console output
type ConsoleReportGenerator struct {
	config       *ReportConfig
	colorEnabled bool
}

// NewConsoleReportGenerator creates a new ConsoleReportGenerator
func NewConsoleReportGenerator() *ConsoleReportGenerator {
	return &ConsoleReportGenerator{
		config:       NewReportConfig(),
		colorEnabled: true,
	}
}

// WithConfig applies configuration to the generator
func (crg *ConsoleReportGenerator) WithConfig(config *ReportConfig) ReportGenerator {
	crg.config = config
	crg.colorEnabled = config.ColorOutput
	return crg
}

// GenerateReport generates a console-optimized report
func (crg *ConsoleReportGenerator) GenerateReport(results map[string]*interfaces.DriftResult, config ReportConfig) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	// Validate config - check if it's properly initialized (empty ReportConfig{} has Format = 0 = FormatJSON)
	// But we want to reject completely uninitialized configs
	if config.Format == 0 && !config.IncludeTimestamp && !config.IncludeSummary && !config.ColorOutput {
		return nil, NewReportError(ErrorTypeInvalidInput, "config appears to be uninitialized")
	}

	// Update color setting from config
	crg.colorEnabled = config.ColorOutput

	switch config.Format {
	case FormatConsole:
		consoleReport, err := crg.GenerateConsoleReport(results)
		if err != nil {
			return nil, err
		}
		return []byte(consoleReport), nil
	case FormatTable:
		tableReport, err := crg.GenerateTableReport(results)
		if err != nil {
			return nil, err
		}
		return []byte(tableReport), nil
	default:
		// Default to console format
		consoleReport, err := crg.GenerateConsoleReport(results)
		if err != nil {
			return nil, err
		}
		return []byte(consoleReport), nil
	}
}

// GenerateJSONReport delegates to standard generator
func (crg *ConsoleReportGenerator) GenerateJSONReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	standardGen := NewStandardReportGenerator()
	return standardGen.GenerateJSONReport(results)
}

// GenerateYAMLReport delegates to standard generator
func (crg *ConsoleReportGenerator) GenerateYAMLReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	standardGen := NewStandardReportGenerator()
	return standardGen.GenerateYAMLReport(results)
}

// GenerateTableReport generates a table with color coding
func (crg *ConsoleReportGenerator) GenerateTableReport(results map[string]*interfaces.DriftResult) (string, error) {
	if results == nil {
		return "", NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	var builder strings.Builder

	// Header with color
	header := "\n=== DRIFT DETECTION REPORT ==="
	if crg.colorEnabled {
		header = crg.colorize(header, ColorBold+ColorCyan)
	}
	builder.WriteString(header + "\n")
	builder.WriteString(crg.colorize(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)), ColorDim))

	// Table header
	tableHeader := fmt.Sprintf("%-30s %-15s %-10s %-15s\n", "Resource ID", "Type", "Status", "Severity")
	builder.WriteString(crg.colorize(tableHeader, ColorBold+ColorWhite))
	builder.WriteString(crg.colorize(strings.Repeat("-", 70), ColorDim) + "\n")

	// Sort results by resource ID for consistent output
	var resourceIDs []string
	for resourceID := range results {
		resourceIDs = append(resourceIDs, resourceID)
	}
	sort.Strings(resourceIDs)

	// Table rows
	for _, resourceID := range resourceIDs {
		result := results[resourceID]
		status := "No Drift"
		statusColor := ColorGreen
		if result.IsDrifted {
			status = "Drift"
			statusColor = crg.getSeverityColor(result.Severity)
		}

		row := fmt.Sprintf("%-30s %-15s %-10s %-15s\n",
			result.ResourceID,
			result.ResourceType,
			crg.colorize(status, statusColor),
			crg.colorize(string(result.Severity), crg.getSeverityColor(result.Severity)))
		builder.WriteString(row)
	}

	return builder.String(), nil
}

// GenerateConsoleReport generates a console report with enhanced formatting
func (crg *ConsoleReportGenerator) GenerateConsoleReport(results map[string]*interfaces.DriftResult) (string, error) {
	if results == nil {
		return "", NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	var builder strings.Builder

	// Enhanced header
	builder.WriteString(crg.generateHeader())

	// Summary section
	builder.WriteString(crg.generateColoredSummary(results))

	// Progress indicator simulation (if enabled)
	if crg.config != nil && crg.config.ShowProgressIndicator {
		builder.WriteString(crg.generateProgressIndicator(len(results)))
	}

	// Detailed results section
	builder.WriteString(crg.colorize("\n📋 DETAILED RESULTS:\n", ColorBold+ColorWhite))
	builder.WriteString(crg.colorize(strings.Repeat("═", 80), ColorDim) + "\n")

	// Sort results by resource ID for consistent output
	var resourceIDs []string
	for resourceID := range results {
		resourceIDs = append(resourceIDs, resourceID)
	}
	sort.Strings(resourceIDs)

	for _, resourceID := range resourceIDs {
		result := results[resourceID]
		builder.WriteString(crg.formatResourceResult(resourceID, result))
	}

	// Results by severity
	builder.WriteString(crg.generateResultsBySeverity(results))



	return builder.String(), nil
}

// WriteToFile delegates to standard generator
func (crg *ConsoleReportGenerator) WriteToFile(content []byte, filePath string) error {
	standardGen := NewStandardReportGenerator()
	return standardGen.WriteToFile(content, filePath)
}

// Helper methods for console formatting

// colorize applies color to text if colors are enabled
func (crg *ConsoleReportGenerator) colorize(text, color string) string {
	if !crg.colorEnabled {
		return text
	}
	return color + text + ColorReset
}

// getSeverityColor returns the appropriate color for a severity level
func (crg *ConsoleReportGenerator) getSeverityColor(severity interfaces.SeverityLevel) string {
	if !crg.colorEnabled {
		return ""
	}

	switch severity {
	case interfaces.SeverityCritical:
		return ColorRed + ColorBold
	case interfaces.SeverityHigh:
		return ColorRed
	case interfaces.SeverityMedium:
		return ColorYellow
	case interfaces.SeverityLow:
		return ColorBlue
	default:
		return ColorGreen
	}
}

// generateHeader creates an enhanced header
func (crg *ConsoleReportGenerator) generateHeader() string {
	var builder strings.Builder

	headerText := `
================================================================================
                           DRIFT DETECTION REPORT
================================================================================
`
	builder.WriteString(crg.colorize(headerText, ColorCyan+ColorBold))
	builder.WriteString(crg.colorize(fmt.Sprintf("Generated: %s\n", time.Now().Format("2006-01-02 15:04:05 MST")), ColorDim))

	return builder.String()
}

// generateCustomHeader creates a header with custom title and color settings
func (crg *ConsoleReportGenerator) generateCustomHeader(title string, colorEnabled bool) string {
	var builder strings.Builder

	// Create separator line
	separator := strings.Repeat("=", len(title)+4)
	builder.WriteString(separator + "\n")
	builder.WriteString(fmt.Sprintf("  %s  \n", title))
	builder.WriteString(separator + "\n")

	if colorEnabled {
		return crg.colorize(builder.String(), ColorCyan+ColorBold)
	}
	return builder.String()
}

// generateSummarySection creates a summary section with custom color settings
func (crg *ConsoleReportGenerator) generateSummarySection(results map[string]*interfaces.DriftResult, colorEnabled bool) string {
	var builder strings.Builder

	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0
	highestSeverity := interfaces.SeverityNone

	for _, result := range results {
		if result.IsDrifted {
			resourcesWithDrift++
			totalDifferences += len(result.DriftDetails)
			if result.Severity > highestSeverity {
				highestSeverity = result.Severity
			}
		}
	}

	builder.WriteString("\nSUMMARY:\n")
	builder.WriteString(fmt.Sprintf("Total Resources: %d\n", totalResources))
	builder.WriteString(fmt.Sprintf("Resources with Drift: %d\n", resourcesWithDrift))
	builder.WriteString(fmt.Sprintf("Resources without Drift: %d\n", totalResources-resourcesWithDrift))
	builder.WriteString(fmt.Sprintf("Total Differences: %d\n", totalDifferences))
	if highestSeverity != interfaces.SeverityNone {
		builder.WriteString(fmt.Sprintf("Highest Severity: %s\n", string(highestSeverity)))
	}

	if colorEnabled {
		return crg.colorize(builder.String(), ColorWhite)
	}
	return builder.String()
}

// generateColoredSummary creates a summary with color coding
func (crg *ConsoleReportGenerator) generateColoredSummary(results map[string]*interfaces.DriftResult) string {
	if len(results) == 0 {
		var builder strings.Builder
		builder.WriteString(crg.colorize("\n📊 SUMMARY:\n", ColorBold+ColorWhite))
		builder.WriteString(fmt.Sprintf("   Total Resources: %s\n", crg.colorize("0", ColorCyan)))
		builder.WriteString(fmt.Sprintf("   Resources with Drift: %s\n", crg.colorize("0", ColorGreen)))
		builder.WriteString(fmt.Sprintf("   %s\n", crg.colorize("✅ No drift detected!", ColorGreen+ColorBold)))
		return builder.String()
	}

	var builder strings.Builder

	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0
	severityCounts := make(map[interfaces.SeverityLevel]int)

	for _, result := range results {
		if result.IsDrifted {
			resourcesWithDrift++
			totalDifferences += len(result.DriftDetails)
		}
		severityCounts[result.Severity]++
	}

	builder.WriteString(crg.colorize("\n📊 SUMMARY:\n", ColorBold+ColorWhite))
	builder.WriteString(fmt.Sprintf("   Total Resources: %s\n", crg.colorize(fmt.Sprintf("%d", totalResources), ColorCyan)))

	if resourcesWithDrift > 0 {
		builder.WriteString(fmt.Sprintf("   Resources with Drift: %s\n", crg.colorize(fmt.Sprintf("%d", resourcesWithDrift), ColorRed)))
		builder.WriteString(fmt.Sprintf("   Total Differences: %s\n", crg.colorize(fmt.Sprintf("%d", totalDifferences), ColorYellow)))
	} else {
		builder.WriteString(fmt.Sprintf("   Resources with Drift: %s\n", crg.colorize("0", ColorGreen)))
		builder.WriteString(fmt.Sprintf("   %s\n", crg.colorize("✅ No drift detected!", ColorGreen+ColorBold)))
	}

	// Severity breakdown
	if resourcesWithDrift > 0 {
		builder.WriteString("\n🔍 SEVERITY BREAKDOWN:\n")
		// Show severity breakdown
		if count := severityCounts[interfaces.SeverityCritical]; count > 0 {
			severityText := fmt.Sprintf("   Critical: %d", count)
			builder.WriteString(crg.colorize(severityText, crg.getSeverityColor(interfaces.SeverityCritical)) + "\n")
		}
		if count := severityCounts[interfaces.SeverityHigh]; count > 0 {
			severityText := fmt.Sprintf("   High: %d", count)
			builder.WriteString(crg.colorize(severityText, crg.getSeverityColor(interfaces.SeverityHigh)) + "\n")
		}
		if count := severityCounts[interfaces.SeverityMedium]; count > 0 {
			severityText := fmt.Sprintf("   Medium: %d", count)
			builder.WriteString(crg.colorize(severityText, crg.getSeverityColor(interfaces.SeverityMedium)) + "\n")
		}
		if count := severityCounts[interfaces.SeverityLow]; count > 0 {
			severityText := fmt.Sprintf("   Low: %d", count)
			builder.WriteString(crg.colorize(severityText, crg.getSeverityColor(interfaces.SeverityLow)) + "\n")
		}
	}

	return builder.String()
}

// formatResourceResult formats a single resource result with colors
func (crg *ConsoleReportGenerator) formatResourceResult(resourceKey string, result *interfaces.DriftResult) string {
	var builder strings.Builder

	// Resource header
	resourceHeader := fmt.Sprintf("\n🔧 Resource: %s", resourceKey)
	if result.IsDrifted {
		resourceHeader = crg.colorize(resourceHeader, ColorRed+ColorBold)
	} else {
		resourceHeader = crg.colorize(resourceHeader, ColorGreen+ColorBold)
	}
	builder.WriteString(resourceHeader + "\n")

	if result.ResourceID != "" {
		builder.WriteString(fmt.Sprintf("   Instance ID: %s\n", crg.colorize(result.ResourceID, ColorCyan)))
	}

	// Status
	status := "✅ No Drift"
	statusColor := ColorGreen
	if result.IsDrifted {
		status = fmt.Sprintf("❌ Drift Detected (%d differences)", len(result.DriftDetails))
		statusColor = crg.getSeverityColor(result.Severity)
	}
	builder.WriteString(fmt.Sprintf("   Status: %s\n", crg.colorize(status, statusColor)))
	builder.WriteString(fmt.Sprintf("   Severity: %s\n", crg.colorize(string(result.Severity), crg.getSeverityColor(result.Severity))))
	builder.WriteString(fmt.Sprintf("   Checked: %s ago\n", time.Since(result.DetectionTime).Round(time.Second)))

	// Differences
	if result.IsDrifted {
		builder.WriteString(fmt.Sprintf("   %s:\n", crg.colorize("Differences", ColorYellow+ColorBold)))
		for i, diff := range result.DriftDetails {
			builder.WriteString(fmt.Sprintf("     %d. %s\n", i+1, crg.colorize(diff.Attribute, ColorWhite+ColorBold)))
			builder.WriteString(fmt.Sprintf("        Expected: %s\n", crg.colorize(fmt.Sprintf("%v", diff.ExpectedValue), ColorGreen)))
			builder.WriteString(fmt.Sprintf("        Actual:   %s\n", crg.colorize(fmt.Sprintf("%v", diff.ActualValue), ColorRed)))
			builder.WriteString(fmt.Sprintf("        Severity: %s\n", crg.colorize(string(diff.Severity), crg.getSeverityColor(diff.Severity))))
			if diff.Description != "" {
				builder.WriteString(fmt.Sprintf("        Description: %s\n", crg.colorize(diff.Description, ColorDim)))
			}
		}
	}

	builder.WriteString(crg.colorize(strings.Repeat("─", 80), ColorDim) + "\n")
	return builder.String()
}

// generateProgressIndicator creates a simple progress indicator
func (crg *ConsoleReportGenerator) generateProgressIndicator(totalResources int) string {
	var builder strings.Builder

	builder.WriteString(crg.colorize("\n⏳ Processing Resources...\n", ColorYellow))
	progressBar := "[" + strings.Repeat("█", 20) + "]"
	builder.WriteString(crg.colorize(fmt.Sprintf("   %s 100%% (%d resources)\n", progressBar, totalResources), ColorGreen))

	return builder.String()
}

// generateCustomProgressIndicator creates a progress indicator with custom parameters
func (crg *ConsoleReportGenerator) generateCustomProgressIndicator(progress, total int, colorEnabled bool) string {
	var builder strings.Builder

	progressWidth := 20
	filledWidth := int(float64(progress) / float64(total) * float64(progressWidth))
	emptyWidth := progressWidth - filledWidth

	progressBar := "[" + strings.Repeat("█", filledWidth) + strings.Repeat("░", emptyWidth) + "]"
	percentage := int(float64(progress) / float64(total) * 100)

	builder.WriteString(fmt.Sprintf("%s %d%% (%d/%d)", progressBar, percentage, progress, total))

	if colorEnabled {
		return crg.colorize(builder.String(), ColorGreen)
	}
	return builder.String()
}

// generateResultsBySeverity groups and displays results by severity
func (crg *ConsoleReportGenerator) generateResultsBySeverity(results map[string]*interfaces.DriftResult) string {
	var builder strings.Builder

	// Group by severity
	severityGroups := make(map[interfaces.SeverityLevel][]*interfaces.DriftResult)
	for _, result := range results {
		if result.IsDrifted {
			severityGroups[result.Severity] = append(severityGroups[result.Severity], result)
		}
	}

	if len(severityGroups) == 0 {
		return ""
	}

	builder.WriteString(crg.colorize("\n🎯 RESULTS BY SEVERITY:\n", ColorBold+ColorWhite))

	// Show results by severity in order
	severities := []interfaces.SeverityLevel{interfaces.SeverityCritical, interfaces.SeverityHigh, interfaces.SeverityMedium, interfaces.SeverityLow}
	for _, severity := range severities {
		if resources := severityGroups[severity]; len(resources) > 0 {
			severityHeader := fmt.Sprintf("\n   %s (%d resources):", strings.ToUpper(string(severity)), len(resources))
			builder.WriteString(crg.colorize(severityHeader, crg.getSeverityColor(severity)+ColorBold) + "\n")
			for _, result := range resources {
				builder.WriteString(fmt.Sprintf("     • %s (%d differences)\n", result.ResourceID, len(result.DriftDetails)))
			}
		}
	}

	return builder.String()
}

// GenerateSimpleReport generates a simple console report without colors
func (crg *ConsoleReportGenerator) GenerateSimpleReport(results map[string]*interfaces.DriftResult) (string, error) {
	if results == nil {
		return "", NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	var builder strings.Builder

	// Header
	builder.WriteString("DRIFT DETECTION REPORT\n")
	builder.WriteString("Generated: " + time.Now().Format("2006-01-02 15:04:05 MST") + "\n\n")

	// Summary
	builder.WriteString(crg.generateSummarySection(results, false))

	// Individual results
	for _, result := range results {
		builder.WriteString(fmt.Sprintf("\nResource: %s\n", result.ResourceID))
		if result.IsDrifted {
			builder.WriteString(fmt.Sprintf("Status: Drift Detected (%d differences)\n", len(result.DriftDetails)))
			builder.WriteString(fmt.Sprintf("Severity: %s\n", string(result.Severity)))
			for i, diff := range result.DriftDetails {
				builder.WriteString(fmt.Sprintf("  %d. %s: %v -> %v\n", i+1, diff.Attribute, diff.ExpectedValue, diff.ActualValue))
			}
		} else {
			builder.WriteString("Status: No Drift\n")
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
