package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"firefly-task/drift"
)

// Test data setup
func createTestDriftResults() map[string]*drift.DriftResult {
	results := make(map[string]*drift.DriftResult)

	// Create a result with drift
	result1 := drift.NewDriftResult("aws_instance.web-server-1", "i-1234567890abcdef0")
	result1.AddDifference(drift.AttributeDifference{
		AttributeName: "instance_type",
		ExpectedValue: "t3.micro",
		ActualValue:   "t3.small",
		Severity:      drift.SeverityMedium,
		Description:   "Instance type mismatch",
	})
	result1.AddDifference(drift.AttributeDifference{
		AttributeName: "security_groups",
		ExpectedValue: []string{"sg-12345"},
		ActualValue:   []string{"sg-12345", "sg-67890"},
		Severity:      drift.SeverityHigh,
		Description:   "Additional security group attached",
	})
	results["aws_instance.web-server-1"] = result1

	// Create a result without drift
	result2 := drift.NewDriftResult("aws_instance.web-server-2", "i-0987654321fedcba0")
	results["aws_instance.web-server-2"] = result2

	// Create a result with critical drift
	result3 := drift.NewDriftResult("aws_instance.database", "i-abcdef1234567890")
	result3.AddDifference(drift.AttributeDifference{
		AttributeName: "publicly_accessible",
		ExpectedValue: false,
		ActualValue:   true,
		Severity:      drift.SeverityCritical,
		Description:   "Database is publicly accessible",
	})
	results["aws_instance.database"] = result3

	return results
}

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
	assert.Equal(t, drift.SeverityNone, config.FilterSeverity)
	assert.True(t, config.IncludeRecommendations)
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
	config := NewReportConfig().WithFilterSeverity(drift.SeverityHigh)
	assert.Equal(t, drift.SeverityHigh, config.FilterSeverity)
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
		WithFilterSeverity(drift.SeverityMedium).
		WithColorOutput(false)

	assert.Equal(t, FormatTable, config.Format)
	assert.Equal(t, "report.txt", config.OutputFile)
	assert.True(t, config.Color)
	assert.Equal(t, drift.SeverityMedium, config.FilterSeverity)
	assert.False(t, config.ColorOutput)
}

func TestReportSummary_Creation(t *testing.T) {
	results := createTestDriftResults()

	summary := &ReportSummary{
		TotalResources:     len(results),
		ResourcesWithDrift: 2,
		TotalDifferences:   3,
		SeverityCounts: map[string]int{
			"critical": 1,
			"high":     1,
			"medium":   1,
		},
		GenerationTime: time.Now().Format(time.RFC3339),
		OverallStatus:  "drift_detected",
	}

	assert.Equal(t, 3, summary.TotalResources)
	assert.Equal(t, 2, summary.ResourcesWithDrift)
	assert.Equal(t, 3, summary.TotalDifferences)
	assert.Equal(t, "drift_detected", summary.OverallStatus)
	assert.Equal(t, 1, summary.SeverityCounts["critical"])
	assert.Equal(t, 1, summary.SeverityCounts["high"])
	assert.Equal(t, 1, summary.SeverityCounts["medium"])
}

func TestRecommendation_Creation(t *testing.T) {
	rec := &Recommendation{
		ResourceID:  "aws_instance.web-server-1",
		Severity:    drift.SeverityHigh,
		Title:       "Fix instance type drift",
		Description: "Update instance type to match configuration",
		Action:      "terraform apply",
		Priority:    1,
	}

	assert.Equal(t, "aws_instance.web-server-1", rec.ResourceID)
	assert.Equal(t, "Fix instance type drift", rec.Title)
	assert.Equal(t, drift.SeverityHigh, rec.Severity)
	assert.Equal(t, "terraform apply", rec.Action)
	assert.Equal(t, 1, rec.Priority)
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
		Recommendations: []Recommendation{},
		Metadata:        map[string]interface{}{"version": "1.0"},
	}

	assert.NotNil(t, reportData.Results)
	assert.NotNil(t, reportData.Summary)
	assert.NotNil(t, reportData.Recommendations)
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
			WithFilterSeverity(drift.SeverityHigh)
	}
}

// Test edge cases
func TestReportConfig_EdgeCases(t *testing.T) {
	// Test with empty output file
	config := NewReportConfig().WithOutputFile("")
	assert.Equal(t, "", config.OutputFile)

	// Test with invalid severity (should still work)
	config = NewReportConfig().WithFilterSeverity(drift.DriftSeverity(999))
	assert.Equal(t, drift.DriftSeverity(999), config.FilterSeverity)
}

func TestReportSummary_EdgeCases(t *testing.T) {
	// Test with empty results
	emptyResults := make(map[string]*drift.DriftResult)

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
