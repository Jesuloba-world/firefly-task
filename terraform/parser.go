package terraform

import (
	"fmt"
	"path/filepath"
	"strings"

	"firefly-task/pkg/interfaces"
)

// Parser provides a unified interface for parsing Terraform configurations and state files
type Parser interface {
	// ParseState parses a Terraform state file and returns extracted configurations
	ParseState(statePath string) (map[string]*interfaces.TerraformConfig, error)

	// ParseHCL parses Terraform HCL configuration files and returns extracted configurations
	ParseHCL(configPath string) (map[string]*interfaces.TerraformConfig, error)

	// ParseBoth parses both state and HCL files and merges the results
	ParseBoth(statePath, configPath string) (map[string]*interfaces.TerraformConfig, error)

	// SetOptions configures parser behavior
	SetOptions(options ParserOptions)

	// GetOptions returns current parser options
	GetOptions() ParserOptions
}

// ParserOptions configures parser behavior
type ParserOptions struct {
	// StrictMode enables strict validation and error handling
	StrictMode bool

	// IgnoreMissingFiles continues parsing even if some files are missing
	IgnoreMissingFiles bool

	// MaxFileSize limits the maximum size of files to parse (in bytes)
	MaxFileSize int64

	// SupportedFormats specifies which file formats to support
	SupportedFormats []string

	// IncludeMetadata includes Terraform version and provider metadata
	IncludeMetadata bool

	// RecursiveModules enables parsing of nested modules
	RecursiveModules bool
}

// DefaultParserOptions returns sensible default options
func DefaultParserOptions() ParserOptions {
	return ParserOptions{
		StrictMode:         false,
		IgnoreMissingFiles: true,
		MaxFileSize:        10 * 1024 * 1024, // 10MB
		SupportedFormats:   []string{".tf", ".tf.json", ".tfstate"},
		IncludeMetadata:    true,
		RecursiveModules:   true,
	}
}

// TerraformParser implements the Parser interface
type TerraformParser struct {
	options ParserOptions
	errors  []error
}

// NewParser creates a new TerraformParser with default options
func NewParser() *TerraformParser {
	return &TerraformParser{
		options: DefaultParserOptions(),
		errors:  make([]error, 0),
	}
}

// NewParserWithOptions creates a new TerraformParser with custom options
func NewParserWithOptions(options ParserOptions) *TerraformParser {
	return &TerraformParser{
		options: options,
		errors:  make([]error, 0),
	}
}

// SetOptions configures parser behavior
func (p *TerraformParser) SetOptions(options ParserOptions) {
	p.options = options
}

// GetOptions returns current parser options
func (p *TerraformParser) GetOptions() ParserOptions {
	return p.options
}

// GetErrors returns accumulated parsing errors
func (p *TerraformParser) GetErrors() []error {
	return p.errors
}

// ClearErrors clears accumulated parsing errors
func (p *TerraformParser) ClearErrors() {
	p.errors = make([]error, 0)
}

// ParseState parses a Terraform state file and returns extracted configurations
func (p *TerraformParser) ParseState(statePath string) (map[string]*interfaces.TerraformConfig, error) {

	// Validate file extension if strict mode is enabled
	if p.options.StrictMode {
		if !p.isSupportedFormat(statePath) {
			return nil, fmt.Errorf("unsupported file format: %s", filepath.Ext(statePath))
		}
	}

	// Parse the state file
	state, err := ParseTerraformState(statePath)
	if err != nil {
		if !p.options.IgnoreMissingFiles {
			return nil, fmt.Errorf("failed to parse state file: %w", err)
		}
		p.errors = append(p.errors, err)
		return make(map[string]*interfaces.TerraformConfig), nil
	}

	// Extract EC2 instances
	instances, err := ExtractEC2InstancesFromState(state)
	if err != nil {
		return nil, fmt.Errorf("failed to extract instances from state: %w", err)
	}

	// Convert to TerraformConfig map
	configs := make(map[string]*interfaces.TerraformConfig)
	for _, instance := range instances {
		config := p.convertEC2InstanceToTerraformConfig(instance)

		// Add metadata if enabled
		if p.options.IncludeMetadata {
			config.TerraformVersion = state.TerraformVersion
		}

		configs[config.ResourceID] = config
	}

	return configs, nil
}

// ParseHCL parses Terraform HCL configuration files and returns extracted configurations
func (p *TerraformParser) ParseHCL(configPath string) (map[string]*interfaces.TerraformConfig, error) {
	p.ClearErrors()

	// Parse the HCL configuration
	parsedConfig, err := ParseTerraformHCL(configPath)
	if err != nil {
		if !p.options.IgnoreMissingFiles {
			return nil, fmt.Errorf("failed to parse HCL configuration: %w", err)
		}
		p.errors = append(p.errors, err)
		return make(map[string]*interfaces.TerraformConfig), nil
	}

	// Extract EC2 instances
	instances, err := ExtractEC2Instances(parsedConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to extract instances from HCL: %w", err)
	}

	// Convert to TerraformConfig map
	configs := make(map[string]*interfaces.TerraformConfig)
	for _, instance := range instances {
		config := p.convertEC2InstanceToTerraformConfig(instance)
		configs[config.ResourceID] = config
	}

	return configs, nil
}

// ParseBoth parses both state and HCL files and merges the results
func (p *TerraformParser) ParseBoth(statePath, configPath string) (map[string]*interfaces.TerraformConfig, error) {
	p.ClearErrors()

	// Parse state file
	stateConfigs, err := p.ParseState(statePath)
	if err != nil && !p.options.IgnoreMissingFiles {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	// Parse HCL configuration
	hclConfigs, err := p.ParseHCL(configPath)
	if err != nil && !p.options.IgnoreMissingFiles {
		return nil, fmt.Errorf("failed to parse HCL: %w", err)
	}

	// Merge configurations (state takes precedence for actual values)
	mergedConfigs := make(map[string]*interfaces.TerraformConfig)

	// Start with HCL configurations (expected values)
	for id, config := range hclConfigs {
		mergedConfigs[id] = config.Clone()
	}

	// Overlay state configurations (actual values)
	for id, stateConfig := range stateConfigs {
		if existingConfig, exists := mergedConfigs[id]; exists {
			// Merge state data into existing HCL config
			p.mergeStateIntoConfig(existingConfig, stateConfig)
		} else {
			// Add state-only configuration
			mergedConfigs[id] = stateConfig.Clone()
		}
	}

	return mergedConfigs, nil
}

// convertEC2InstanceToTerraformConfig converts EC2InstanceConfig to TerraformConfig
func (p *TerraformParser) convertEC2InstanceToTerraformConfig(instance EC2InstanceConfig) *interfaces.TerraformConfig {
	config := &interfaces.TerraformConfig{
		ResourceID:   fmt.Sprintf("aws_instance.%s", instance.ResourceName),
		ResourceName: instance.ResourceName,
		ResourceType: "aws_instance",
		Attributes: map[string]interface{}{
			"instance_type": instance.InstanceType,
			"ami":           instance.AMI,
			"key_name":      instance.KeyName,
			"subnet_id":     instance.SubnetID,
			"vpc_security_group_ids": instance.VPCSecurityGroups,
			"tags":          instance.Tags,
		},
	}

	return config
}

// mergeStateIntoConfig merges state configuration into HCL configuration
func (p *TerraformParser) mergeStateIntoConfig(hclConfig, stateConfig *interfaces.TerraformConfig) {
	// Merge attributes from state
	for key, value := range stateConfig.Attributes {
		hclConfig.Attributes[key] = value
	}

	// Merge metadata
	if stateConfig.TerraformVersion != "" {
		hclConfig.TerraformVersion = stateConfig.TerraformVersion
	}
	if stateConfig.ProviderVersion != "" {
		hclConfig.ProviderVersion = stateConfig.ProviderVersion
	}
}

// isSupportedFormat checks if the file format is supported
func (p *TerraformParser) isSupportedFormat(filePath string) bool {
	ext := filepath.Ext(filePath)
	for _, supportedExt := range p.options.SupportedFormats {
		if strings.EqualFold(ext, supportedExt) {
			return true
		}
	}
	return false
}

// ParseError represents a parsing error with context
type ParseError struct {
	Type    string
	Message string
	Path    string
	Cause   error
}

func (e *ParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s error in %s: %s (caused by: %v)", e.Type, e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s error in %s: %s", e.Type, e.Path, e.Message)
}

func (e *ParseError) Unwrap() error {
	return e.Cause
}

// NewParseError creates a new ParseError
func NewParseError(errorType, message, path string, cause error) *ParseError {
	return &ParseError{
		Type:    errorType,
		Message: message,
		Path:    path,
		Cause:   cause,
	}
}
