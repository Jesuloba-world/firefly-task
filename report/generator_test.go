package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"firefly-task/pkg/interfaces"
)



func TestReportFormat_String(t *testing.T) {
	tests := []struct {
		format   ReportFormat
		expected string
	}{
		{FormatJSON, "json"},
		{FormatYAML, "yaml"},
		{FormatTable, "table"},
		{FormatConsole, "console"},
		{FormatCI, "ci"},
		{ReportFormat(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.format.String())
		})
	}
}

func TestNewReportConfig(t *testing.T) {
	config := NewReportConfig()

	assert.Equal(t, FormatJSON, config.Format)
	assert.True(t, config.IncludeTimestamp)
	assert.True(t, config.IncludeSummary)
	assert.True(t, config.ColorOutput)
	assert.Equal(t, interfaces.SeverityNone, config.FilterSeverity)

	assert.False(t, config.ShowProgressIndicator)
}

func TestReportConfig_WithFormat(t *testing.T) {
	config := NewReportConfig().WithFormat(FormatYAML)
	assert.Equal(t, FormatYAML, config.Format)
}

func TestReportConfig_WithOutputFile(t *testing.T) {
	filename := "test-report.json"
	config := NewReportConfig().WithOutputFile(filename)
	assert.Equal(t, filename, config.OutputFile)
}

func TestReportConfig_WithColor(t *testing.T) {
	config := NewReportConfig().WithColor(true)
	assert.True(t, config.Color)
}

func TestReportConfig_WithFilterSeverity(t *testing.T) {
	config := NewReportConfig().WithFilterSeverity(interfaces.SeverityHigh)
	assert.Equal(t, interfaces.SeverityHigh, config.FilterSeverity)
}

func TestReportConfig_WithColorOutput(t *testing.T) {
	config := NewReportConfig().WithColorOutput(false)
	assert.False(t, config.ColorOutput)
}

func TestReportConfig_Chaining(t *testing.T) {
	// Test method chaining
	config := NewReportConfig().
		WithFormat(FormatTable).
		WithOutputFile("report.txt").
		WithColor(true).
		WithFilterSeverity(interfaces.SeverityMedium).
		WithColorOutput(false)

	assert.Equal(t, FormatTable, config.Format)
	assert.Equal(t, "report.txt", config.OutputFile)
	assert.True(t, config.Color)
	assert.Equal(t, interfaces.SeverityMedium, config.FilterSeverity)
	assert.False(t, config.ColorOutput)
}

func TestReportSummary_Creation(t *testing.T) {
	results := createTestDriftResults()

	summary := &ReportSummary{
		TotalResources:     len(results),
		ResourcesWithDrift: 3,
		TotalDifferences:   3,
		SeverityCounts: map[string]int{
			"critical": 1,
			"high":     1,
			"medium":   1,
		},
		GenerationTime: time.Now().Format(time.RFC3339),
		OverallStatus:  "drift_detected",
	}

	assert.Equal(t, 4, summary.TotalResources)
	assert.Equal(t, 3, summary.ResourcesWithDrift)
	assert.Equal(t, 3, summary.TotalDifferences)
	assert.Equal(t, "drift_detected", summary.OverallStatus)
	assert.Equal(t, 1, summary.SeverityCounts["critical"])
	assert.Equal(t, 1, summary.SeverityCounts["high"])
	assert.Equal(t, 1, summary.SeverityCounts["medium"])
}



func TestReportData_Creation(t *testing.T) {
	results := createTestDriftResults()

	reportData := &ReportData{
		Results: results,
		Summary: ReportSummary{
			TotalResources:     len(results),
			ResourcesWithDrift: 1,
			TotalDifferences:   2,
			SeverityCounts:     map[string]int{"high": 1},
			GenerationTime:     time.Now().Format(time.RFC3339),
			OverallStatus:      "drift_detected",
		},
		Metadata:        map[string]interface{}{"version": "1.0"},
	}

	assert.NotNil(t, reportData.Results)
	assert.NotNil(t, reportData.Summary)
	assert.NotNil(t, reportData.Metadata)
}

// Benchmark tests
func BenchmarkReportConfig_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewReportConfig()
	}
}

func BenchmarkReportConfig_Chaining(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewReportConfig().
			WithFormat(FormatYAML).
			WithColor(true).
			WithFilterSeverity(interfaces.SeverityHigh)
	}
}

// Test edge cases
func TestReportConfig_EdgeCases(t *testing.T) {
	// Test with empty output file
	config := NewReportConfig().WithOutputFile("")
	assert.Equal(t, "", config.OutputFile)

	// Test with invalid severity (should still work)
	config = NewReportConfig().WithFilterSeverity(interfaces.SeverityLevel("invalid"))
	assert.Equal(t, interfaces.SeverityLevel("invalid"), config.FilterSeverity)
}

func TestReportSummary_EdgeCases(t *testing.T) {
	// Test with empty results
	emptyResults := make(map[string]*interfaces.DriftResult)

	summary := &ReportSummary{
		TotalResources:     0,
		ResourcesWithDrift: 0,
		TotalDifferences:   0,
		SeverityCounts:     make(map[string]int),
		GenerationTime:     time.Now().Format(time.RFC3339),
		OverallStatus:      "no_drift",
	}

	assert.Equal(t, 0, summary.TotalResources)
	assert.Equal(t, 0, summary.ResourcesWithDrift)
	assert.Equal(t, "no_drift", summary.OverallStatus)
	assert.NotNil(t, summary.SeverityCounts)

	_ = emptyResults // Use the variable to avoid unused variable error
}
