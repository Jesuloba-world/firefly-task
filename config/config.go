package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all configuration values for the application
type Config struct {
	// Global flags
	AWSProfile string
	AWSRegion  string
	Verbose    bool
	Output     string

	// Command-specific flags
	TerraformPath string
	InstanceID    string
	InputFile     string
	Concurrency   int
	Attribute     string
}

// OutputFormat represents valid output formats
type OutputFormat string

const (
	OutputText  OutputFormat = "text"
	OutputJSON  OutputFormat = "json"
	OutputTable OutputFormat = "table"
)

// ValidateOutputFormat checks if the output format is valid
func ValidateOutputFormat(format string) error {
	switch OutputFormat(format) {
	case OutputText, OutputJSON, OutputTable:
		return nil
	default:
		return fmt.Errorf("invalid output format '%s'. Valid formats: text, json, table", format)
	}
}

// NormalizePath normalizes file paths for Windows compatibility
func NormalizePath(path string) string {
	if path == "" {
		return path
	}
	
	// Convert forward slashes to backslashes on Windows
	normalized := filepath.Clean(path)
	
	// Handle relative paths
	if !filepath.IsAbs(normalized) {
		if wd, err := os.Getwd(); err == nil {
			normalized = filepath.Join(wd, normalized)
		}
	}
	
	return normalized
}

// ValidateConfig validates the configuration values
func (c *Config) ValidateConfig() error {
	// Validate output format
	if err := ValidateOutputFormat(c.Output); err != nil {
		return err
	}
	
	// Validate concurrency
	if c.Concurrency < 1 {
		c.Concurrency = 1 // Default to 1 if not set or invalid
	}
	if c.Concurrency > 50 {
		return fmt.Errorf("concurrency cannot exceed 50, got %d", c.Concurrency)
	}
	
	// Normalize paths
	if c.TerraformPath != "" {
		c.TerraformPath = NormalizePath(c.TerraformPath)
	}
	if c.InputFile != "" {
		c.InputFile = NormalizePath(c.InputFile)
	}
	
	return nil
}

// GetAWSRegionFromEnv gets AWS region from environment variables if not set
func (c *Config) GetAWSRegionFromEnv() {
	if c.AWSRegion == "" {
		if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
			c.AWSRegion = region
		} else if region := os.Getenv("AWS_REGION"); region != "" {
			c.AWSRegion = region
		}
	}
}

// GetAWSProfileFromEnv gets AWS profile from environment variables if not set
func (c *Config) GetAWSProfileFromEnv() {
	if c.AWSProfile == "" {
		if profile := os.Getenv("AWS_PROFILE"); profile != "" {
			c.AWSProfile = profile
		}
	}
}

// SetDefaults sets default values for configuration
func (c *Config) SetDefaults() {
	if c.Output == "" {
		c.Output = string(OutputText)
	}
	if c.Concurrency == 0 {
		c.Concurrency = 5
	}
	
	// Get AWS settings from environment
	c.GetAWSRegionFromEnv()
	c.GetAWSProfileFromEnv()
}

// String returns a string representation of the config (for debugging)
func (c *Config) String() string {
	var parts []string
	
	if c.AWSProfile != "" {
		parts = append(parts, fmt.Sprintf("AWS Profile: %s", c.AWSProfile))
	}
	if c.AWSRegion != "" {
		parts = append(parts, fmt.Sprintf("AWS Region: %s", c.AWSRegion))
	}
	parts = append(parts, fmt.Sprintf("Output: %s", c.Output))
	parts = append(parts, fmt.Sprintf("Verbose: %t", c.Verbose))
	
	if c.TerraformPath != "" {
		parts = append(parts, fmt.Sprintf("Terraform Path: %s", c.TerraformPath))
	}
	if c.InstanceID != "" {
		parts = append(parts, fmt.Sprintf("Instance ID: %s", c.InstanceID))
	}
	if c.InputFile != "" {
		parts = append(parts, fmt.Sprintf("Input File: %s", c.InputFile))
	}
	if c.Attribute != "" {
		parts = append(parts, fmt.Sprintf("Attribute: %s", c.Attribute))
	}
	parts = append(parts, fmt.Sprintf("Concurrency: %d", c.Concurrency))
	
	return strings.Join(parts, ", ")
}