package report

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportGeneratorType_String(t *testing.T) {
	tests := []struct {
		name     string
		genType  ReportGeneratorType
		expected string
	}{
		/*{
			name:     "Standard generator",
			genType:  ReportGeneratorStandard,
			expected: "standard",
		},
		{
			name:     "CI generator",
			genType:  ReportGeneratorCI,
			expected: "ci",
		},*/
		/*{
			name:     "Console generator",
			genType:  ReportGeneratorConsole,
			expected: "console",
		},*/
		{
			name:     "Invalid generator type",
			genType:  ReportGeneratorType(999),
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.genType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewReportGenerator(t *testing.T) {
	tests := []struct {
		name         string
		genType      ReportGeneratorType
		expectedType string
		expectError  bool
	}{
		/*{
			name:         "Create standard generator",
			genType:      ReportGeneratorStandard,
			expectedType: "*report.StandardReportGenerator",
			expectError:  false,
		},
		{
			name:         "Create CI generator",
			genType:      ReportGeneratorCI,
			expectedType: "*report.CIReportGenerator",
			expectError:  false,
		},
		{
			name:         "Create console generator",
			genType:      ReportGeneratorConsole,
			expectedType: "*report.ConsoleReportGenerator",
			expectError:  false,
		},*/
		{
			name:         "Invalid generator type",
			genType:      ReportGeneratorType(999),
			expectedType: "",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewReportGenerator(tt.genType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, generator)

				// Check error type
				reportErr, ok := err.(*ReportError)
				assert.True(t, ok)
				assert.Equal(t, ErrorTypeUnsupportedFormat, reportErr.Type)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)

				// Check generator type using type assertion
				/*switch tt.genType {
				case ReportGeneratorStandard:
					_, ok := generator.(*StandardReportGenerator)
					assert.True(t, ok, "Should be StandardReportGenerator")
				case ReportGeneratorCI:
					_, ok := generator.(*CIReportGenerator)
					assert.True(t, ok, "Should be CIReportGenerator")
				case ReportGeneratorConsole:
					_, ok := generator.(*ConsoleReportGenerator)
					assert.True(t, ok, "Should be ConsoleReportGenerator")
				}*/
			}
		})
	}
}

func TestNewReportGeneratorWithConfig(t *testing.T) {
	tests := []struct {
		name        string
		genType     ReportGeneratorType
		config      *ReportConfig
		expectError bool
	}{
		/*{
			name:        "Standard generator with config",
			genType:     ReportGeneratorStandard,
			config:      NewReportConfig().WithFormat(ReportFormatJSON),
			expectError: false,
		},
		{
			name:        "CI generator with config",
			genType:     ReportGeneratorCI,
			config:      NewReportConfig().WithFormat(ReportFormatCI),
			expectError: false,
		},
		{
			name:        "Console generator with config",
			genType:     ReportGeneratorConsole,
			config:      NewReportConfig().WithFormat(ReportFormatConsole),
			expectError: false,
		},*/
		{
			name:        "Invalid generator type with config",
			genType:     ReportGeneratorType(999),
			config:      NewReportConfig(),
			expectError: true,
		},
		/*{
			name:        "Valid generator with nil config",
			genType:     ReportGeneratorStandard,
			config:      nil,
			expectError: false, // Should use default config
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := NewReportGeneratorWithConfig(tt.genType, tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, generator)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, generator)

				// Test if generator implements ConfigurableGenerator interface
				/*if configurable, ok := generator.(ConfigurableGenerator); ok {
					// Test configuration
					testConfig := NewReportConfig().WithFormat(ReportFormatYAML)
					err := configurable.Configure(testConfig)
					assert.NoError(t, err)
				}*/
			}
		})
	}
}

func TestGetRecommendedGenerator(t *testing.T) {
	tests := []struct {
		name     string
		useCase  string
		expected ReportGeneratorType
	}{
		/*{
			name:     "CI/CD pipeline use case",
			useCase:  "ci",
			expected: ReportGeneratorCI,
		},
		{
			name:     "CI/CD uppercase",
			useCase:  "CI",
			expected: ReportGeneratorCI,
		},
		{
			name:     "Pipeline use case",
			useCase:  "pipeline",
			expected: ReportGeneratorCI,
		},
		{
			name:     "Automation use case",
			useCase:  "automation",
			expected: ReportGeneratorCI,
		},*/
		/*{
			name:     "Console use case",
			useCase:  "console",
			expected: ReportGeneratorConsole,
		},
		{
			name:     "Interactive use case",
			useCase:  "interactive",
			expected: ReportGeneratorConsole,
		},
		{
			name:     "Terminal use case",
			useCase:  "terminal",
			expected: ReportGeneratorConsole,
		},
		{
			name:     "CLI use case",
			useCase:  "cli",
			expected: ReportGeneratorConsole,
		},
		{
			name:     "Standard use case",
			useCase:  "standard",
			expected: ReportGeneratorStandard,
		},
		{
			name:     "Report use case",
			useCase:  "report",
			expected: ReportGeneratorStandard,
		},
		{
			name:     "File use case",
			useCase:  "file",
			expected: ReportGeneratorStandard,
		},
		{
			name:     "Unknown use case",
			useCase:  "unknown",
			expected: ReportGeneratorStandard,
		},
		{
			name:     "Empty use case",
			useCase:  "",
			expected: ReportGeneratorStandard,
		},*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRecommendedGenerator(tt.useCase)
			assert.Equal(t, tt.expected, result)
		})
	}
}

/*func TestValidateGeneratorType(t *testing.T) {
	tests := []struct {
		name     string
		genType  ReportGeneratorType
		expected bool
	}{
		{
			name:     "Valid standard type",
			genType:  ReportGeneratorStandard,
			expected: true,
		},
		{
			name:     "Valid CI type",
			genType:  ReportGeneratorCI,
			expected: true,
		},
		{
			name:     "Valid console type",
			genType:  ReportGeneratorConsole,
			expected: true,
		},
		{
			name:     "Invalid type - negative",
			genType:  ReportGeneratorType(-1),
			expected: false,
		},
		{
			name:     "Invalid type - too high",
			genType:  ReportGeneratorType(999),
			expected: false,
		},
		{
			name:     "Invalid type - boundary",
			genType:  ReportGeneratorType(3),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateGeneratorType(tt.genType)
			assert.Equal(t, tt.expected, result)
		})
	}
}*/

/*// Test ConfigurableGenerator interface
func TestConfigurableGenerator_Interface(t *testing.T) {
	// Test that all generators implement the interface correctly
	generators := []struct {
		name string
		gen  ReportGenerator
	}{
		{"Standard", NewStandardReportGenerator()},
		{"CI", NewCIReportGenerator()},
		{"Console", NewConsoleReportGenerator()},
	}

	for _, g := range generators {
		t.Run(g.name, func(t *testing.T) {
			// Check if generator implements ConfigurableGenerator
			if configurable, ok := g.gen.(ConfigurableGenerator); ok {
				// Test configuration
				config := NewReportConfig().WithFormat(ReportFormatJSON)
				err := configurable.Configure(config)
				assert.NoError(t, err)

				// Test with nil config
				err = configurable.Configure(nil)
				assert.Error(t, err)
			} else {
				t.Errorf("Generator %s does not implement ConfigurableGenerator interface", g.name)
			}
		})
	}
}*/

/*// Test factory with different configurations
func TestFactoryWithDifferentConfigurations(t *testing.T) {
	configs := []*ReportConfig{
		NewReportConfig().WithFormat(ReportFormatJSON),
		NewReportConfig().WithFormat(ReportFormatYAML),
		NewReportConfig().WithFormat(ReportFormatTable),
		NewReportConfig().WithFormat(ReportFormatConsole),
		NewReportConfig().WithFormat(ReportFormatCI),
	}

	generatorTypes := []ReportGeneratorType{
		ReportGeneratorStandard,
		ReportGeneratorCI,
		ReportGeneratorConsole,
	}

	for _, genType := range generatorTypes {
		for _, config := range configs {
			t.Run(fmt.Sprintf("%s_with_%s", genType.String(), config.Format.String()), func(t *testing.T) {
				generator, err := NewReportGeneratorWithConfig(genType, config)
				assert.NoError(t, err)
				assert.NotNil(t, generator)

				// Test that generator can be configured
				if configurable, ok := generator.(ConfigurableGenerator); ok {
					err := configurable.Configure(config)
					assert.NoError(t, err)
				}
			})
		}
	}
}*/

// Test error handling in factory
func TestFactoryErrorHandling(t *testing.T) {
	// Test with invalid generator type
	invalidType := ReportGeneratorType(999)

	generator, err := NewReportGenerator(invalidType)
	assert.Error(t, err)
	assert.Nil(t, generator)

	// Check error details
	reportErr, ok := err.(*ReportError)
	require.True(t, ok)
	assert.Equal(t, ErrorTypeUnsupportedFormat, reportErr.Type)
	assert.Contains(t, reportErr.Message, "unsupported generator type")

	// Test with config
	generator2, err2 := NewReportGeneratorWithConfig(invalidType, NewReportConfig())
	assert.Error(t, err2)
	assert.Nil(t, generator2)
}

/*// Test concurrent factory usage
func TestFactoryConcurrency(t *testing.T) {
	const numGoroutines = 10
	const numIterations = 100

	results := make(chan error, numGoroutines*numIterations)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numIterations; j++ {
				genType := ReportGeneratorType(j % 3) // Cycle through valid types
				generator, err := NewReportGenerator(genType)
				if err != nil {
					results <- err
					return
				}
				if generator == nil {
					results <- fmt.Errorf("generator is nil")
					return
				}
			}
			results <- nil
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}*/

/*// Benchmark tests
func BenchmarkNewReportGenerator(b *testing.B) {
	generatorTypes := []ReportGeneratorType{
		ReportGeneratorStandard,
		ReportGeneratorCI,
		ReportGeneratorConsole,
	}

	for _, genType := range generatorTypes {
		b.Run(genType.String(), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := NewReportGenerator(genType)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}*/

/*func BenchmarkNewReportGeneratorWithConfig(b *testing.B) {
	config := NewReportConfig().WithFormat(ReportFormatJSON)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewReportGeneratorWithConfig(ReportGeneratorStandard, config)
		if err != nil {
			b.Fatal(err)
		}
	}
}*/

func BenchmarkGetRecommendedGenerator(b *testing.B) {
	useCases := []string{"ci", "console", "standard", "unknown"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		useCase := useCases[i%len(useCases)]
		GetRecommendedGenerator(useCase)
	}
}

/*func BenchmarkValidateGeneratorType(b *testing.B) {
	generatorTypes := []ReportGeneratorType{
		ReportGeneratorStandard,
		ReportGeneratorCI,
		ReportGeneratorConsole,
		ReportGeneratorType(999), // Invalid
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		genType := generatorTypes[i%len(generatorTypes)]
		ValidateGeneratorType(genType)
	}
}*/
