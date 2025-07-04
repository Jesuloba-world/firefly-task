package report

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"firefly-task/drift"
)

func TestConsoleReportGenerator_GenerateConsole(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()
	config := NewReportConfig().WithFormat(FormatConsole).WithColor(false)

	data, err := generator.GenerateReport(results, *config)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	consoleOutput := string(data)

	// Check header
	assert.Contains(t, consoleOutput, "DRIFT DETECTION REPORT")
	assert.Contains(t, consoleOutput, "=")

	// Check summary section
	assert.Contains(t, consoleOutput, "SUMMARY")
	assert.Contains(t, consoleOutput, "Total Resources:")
	assert.Contains(t, consoleOutput, "Resources with Drift:")
	assert.Contains(t, consoleOutput, "Total Differences:")

	// Check severity breakdown
	assert.Contains(t, consoleOutput, "SEVERITY BREAKDOWN")
	assert.Contains(t, consoleOutput, "Critical:")
	assert.Contains(t, consoleOutput, "Medium:")
	assert.Contains(t, consoleOutput, "Low:")

	// Check detailed results section
	assert.Contains(t, consoleOutput, "DETAILED RESULTS")
	assert.Contains(t, consoleOutput, "aws_instance.web-server-1")
	assert.Contains(t, consoleOutput, "aws_db_instance.database")

	// Check drift indicators
	assert.Contains(t, consoleOutput, "Drift Detected")
	assert.Contains(t, consoleOutput, "No Drift")
}

func TestConsoleReportGenerator_GenerateConsoleWithColor(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()
	config := NewReportConfig().WithFormat(FormatConsole).WithColor(true)

	data, err := generator.GenerateReport(results, *config)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	consoleOutput := string(data)

	// Check for ANSI color codes
	assert.Contains(t, consoleOutput, ColorRed)   // For critical/high severity
	assert.Contains(t, consoleOutput, ColorGreen) // For no drift
	assert.Contains(t, consoleOutput, ColorReset) // Reset codes

	// Check that colors are properly reset
	colorCount := strings.Count(consoleOutput, ColorReset)
	assert.Greater(t, colorCount, 0, "Should contain color reset codes")
}

/*
func TestConsoleReportGenerator_GenerateTable(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()
	config := NewReportConfig().WithFormat(ReportFormatTable).WithColor(false)

	data, err := generator.GenerateTable(results, config)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	tableOutput := string(data)

	// Check table headers
	assert.Contains(t, tableOutput, "Resource ID")
	assert.Contains(t, tableOutput, "Status")
	assert.Contains(t, tableOutput, "Severity")
	assert.Contains(t, tableOutput, "Differences")

	// Check table borders
	assert.Contains(t, tableOutput, "+")
	assert.Contains(t, tableOutput, "|")
	assert.Contains(t, tableOutput, "-")

	// Check resource data
	assert.Contains(t, tableOutput, "aws_instance.web-server-1")
	assert.Contains(t, tableOutput, "aws_instance.web-server-2")
	assert.Contains(t, tableOutput, "aws_instance.database")

	// Check status indicators
	assert.Contains(t, tableOutput, "DRIFT")
	assert.Contains(t, tableOutput, "NO DRIFT")

	// Check severity levels
	assert.Contains(t, tableOutput, "CRITICAL")
	assert.Contains(t, tableOutput, "HIGH")
	assert.Contains(t, tableOutput, "MEDIUM")
}
*/

/*
func TestConsoleReportGenerator_GenerateTableWithColor(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()
	config := NewReportConfig().WithFormat(ReportFormatTable).WithColor(true)

	data, err := generator.GenerateTable(results, config)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	tableOutput := string(data)

	// Check for color codes in table
	assert.Contains(t, tableOutput, "\033[") // ANSI escape sequence
	assert.Contains(t, tableOutput, ColorReset)
}
*/

/*
func TestConsoleReportGenerator_ColorizeText(t *testing.T) {
	generator := NewConsoleReportGenerator()

	tests := []struct {
		name     string
		text     string
		color    string
		useColor bool
		expected string
	}{
		{
			name:     "with color enabled",
			text:     "test",
			color:    ColorRed,
			useColor: true,
			expected: ColorRed + "test" + ColorReset,
		},
		{
			name:     "with color disabled",
			text:     "test",
			color:    ColorRed,
			useColor: false,
			expected: "test",
		},
		{
			name:     "empty text with color",
			text:     "",
			color:    ColorGreen,
			useColor: true,
			expected: ColorGreen + "" + ColorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.colorizeText(tt.text, tt.color, tt.useColor)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConsoleReportGenerator_GenerateHeader(t *testing.T) {
	generator := NewConsoleReportGenerator()

	header := generator.generateCustomHeader("TEST HEADER", false)

	assert.Contains(t, header, "TEST HEADER")
	assert.Contains(t, header, "=")

	// Check that header is properly formatted
	lines := strings.Split(header, "\n")
	assert.GreaterOrEqual(t, len(lines), 3) // Should have separator, title, separator
}

func TestConsoleReportGenerator_GenerateHeaderWithColor(t *testing.T) {
	generator := NewConsoleReportGenerator()

	header := generator.generateCustomHeader("TEST HEADER", true)

	assert.Contains(t, header, "TEST HEADER")
	assert.Contains(t, header, ColorBold)
	assert.Contains(t, header, ColorReset)
}

func TestConsoleReportGenerator_GenerateSummarySection(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()

	summary := generator.generateSummarySection(results, false)

	assert.Contains(t, summary, "SUMMARY")
	assert.Contains(t, summary, "Total Resources: 3")
	assert.Contains(t, summary, "Resources with Drift: 2")
	assert.Contains(t, summary, "Resources without Drift: 1")
	assert.Contains(t, summary, "Total Differences: 3")
	assert.Contains(t, summary, "Highest Severity: CRITICAL")
}

func TestConsoleReportGenerator_GenerateProgressIndicator(t *testing.T) {
	generator := NewConsoleReportGenerator()

	// Test different progress values
	tests := []struct {
		progress int
		total    int
	}{
		{0, 10},
		{5, 10},
		{10, 10},
		{3, 7},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d/%d", tt.progress, tt.total), func(t *testing.T) {
			indicator := generator.generateCustomProgressIndicator(tt.progress, tt.total, false)

			assert.Contains(t, indicator, "[")
			assert.Contains(t, indicator, "]")
			assert.Contains(t, indicator, fmt.Sprintf("%d/%d", tt.progress, tt.total))

			// Check progress bar characters
			if tt.progress > 0 {
				assert.Contains(t, indicator, "█") // Filled character
			}
			if tt.progress < tt.total {
				assert.Contains(t, indicator, "░") // Empty character
			}
		})
	}
}
*/



// Test edge cases
func TestConsoleReportGenerator_EmptyResults(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := make(map[string]*drift.DriftResult)
	config := NewReportConfig().WithFormat(FormatConsole).WithColor(false)

	data, err := generator.GenerateReport(results, *config)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	consoleOutput := string(data)

	assert.Contains(t, consoleOutput, "No drift detected!")
	assert.Regexp(t, `Total Resources: .*0`, consoleOutput)

	// Test with color enabled
	config.WithColor(true)
	data, err = generator.GenerateReport(results, *config)
	require.NoError(t, err)
	consoleOutput = string(data)
	assert.Contains(t, consoleOutput, "No drift detected!")
	assert.Regexp(t, `Total Resources: .*0`, consoleOutput)
}

func TestConsoleReportGenerator_NilResults(t *testing.T) {
	generator := NewConsoleReportGenerator()
	config := NewReportConfig()

	data, err := generator.GenerateReport(nil, *config)
	assert.Error(t, err)
	assert.Empty(t, data)

	// Check error type
	reportErr, ok := err.(*ReportError)
	assert.True(t, ok)
	assert.Equal(t, ErrorTypeInvalidInput, reportErr.Type)
}

func TestConsoleReportGenerator_NilConfig(t *testing.T) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()

	data, err := generator.GenerateReport(results, ReportConfig{})
	assert.Error(t, err)
	assert.Empty(t, data)
}

// Test severity color mapping
/*
func TestConsoleReportGenerator_GetSeverityColor(t *testing.T) {
	generator := NewConsoleReportGenerator()

	tests := []struct {
		severity drift.DriftSeverity
		expected string
	}{
		{drift.SeverityCritical, ColorRed},
		{drift.SeverityHigh, ColorRed},
		{drift.SeverityMedium, ColorYellow},
		{drift.SeverityLow, ColorBlue},
	}

	for _, tt := range tests {
		t.Run(tt.severity.String(), func(t *testing.T) {
			color := generator.getSeverityColor(tt.severity)
			assert.Equal(t, tt.expected, color)
		})
	}
}
*/

// Test status color mapping
/*
func TestConsoleReportGenerator_GetStatusColor(t *testing.T) {
	generator := NewConsoleReportGenerator()

	// Test with drift
	result1 := drift.NewDriftResult("test-resource", "i-test-resource")
	result1.AddDifference(drift.AttributeDifference{
		AttributeName: "test",
		Severity:      drift.SeverityHigh,
	})
	color1 := generator.getStatusColor(result1)
	assert.Equal(t, ColorRed, color1)

	// Test without drift
	result2 := drift.NewDriftResult("test-resource-2", "i-test-resource-2")
	color2 := generator.getStatusColor(result2)
	assert.Equal(t, ColorGreen, color2)
}
*/

// Benchmark tests
func BenchmarkConsoleReportGenerator_GenerateReport(b *testing.B) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()
	config := NewReportConfig().WithFormat(FormatConsole)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateReport(results, *config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

/*
func BenchmarkConsoleReportGenerator_GenerateTable(b *testing.B) {
	generator := NewConsoleReportGenerator()
	results := createTestDriftResults()
	config := NewReportConfig().WithFormat(ReportFormatTable)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateTable(results, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}
*/
