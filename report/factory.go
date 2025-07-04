package report

import (
	"fmt"
)

// ReportGeneratorType represents different types of report generators
type ReportGeneratorType int

const (
	// StandardGenerator is the default report generator
	StandardGenerator ReportGeneratorType = iota
	// CIGenerator is optimized for CI/CD pipelines
	CIGenerator
	// ConsoleGenerator is optimized for console output
	ConsoleGenerator
)

// String returns the string representation of ReportGeneratorType
func (rgt ReportGeneratorType) String() string {
	switch rgt {
	case StandardGenerator:
		return "standard"
	case CIGenerator:
		return "ci"
	case ConsoleGenerator:
		return "console"
	default:
		return "unknown"
	}
}

// NewReportGenerator creates a new report generator instance based on the specified type
func NewReportGenerator(generatorType ReportGeneratorType) (ReportGenerator, error) {
	switch generatorType {
	case StandardGenerator:
		return NewStandardReportGenerator(), nil
	case CIGenerator:
		return NewCIReportGenerator(), nil
	case ConsoleGenerator:
		return NewConsoleReportGenerator(), nil
	default:
		return nil, &ReportError{
			Type:    ErrorTypeInvalidGenerator,
			Message: fmt.Sprintf("unsupported report generator type: %s", generatorType.String()),
		}
	}
}

// NewReportGeneratorWithConfig creates a new report generator with custom configuration
func NewReportGeneratorWithConfig(generatorType ReportGeneratorType, config *ReportConfig) (ReportGenerator, error) {
	generator, err := NewReportGenerator(generatorType)
	if err != nil {
		return nil, err
	}
	
	// Apply configuration-specific optimizations
	if configurable, ok := generator.(ConfigurableGenerator); ok {
		return configurable.WithConfig(config), nil
	}
	
	return generator, nil
}

// ConfigurableGenerator interface for generators that support configuration
type ConfigurableGenerator interface {
	ReportGenerator
	WithConfig(config *ReportConfig) ReportGenerator
}

// GetRecommendedGenerator returns the recommended generator type based on the use case
func GetRecommendedGenerator(useCase string) ReportGeneratorType {
	switch useCase {
	case "ci", "pipeline", "automation":
		return CIGenerator
	case "console", "terminal", "interactive":
		return ConsoleGenerator
	default:
		return StandardGenerator
	}
}

// ValidateGeneratorType checks if the generator type is valid
func ValidateGeneratorType(generatorType ReportGeneratorType) error {
	switch generatorType {
	case StandardGenerator, CIGenerator, ConsoleGenerator:
		return nil
	default:
		return &ReportError{
			Type:    ErrorTypeInvalidGenerator,
			Message: fmt.Sprintf("invalid report generator type: %d", generatorType),
		}
	}
}