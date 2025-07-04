package drift

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// ConfigManager handles loading and saving drift detection configurations
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
	}
}

// LoadConfig loads configuration from file or returns default if file doesn't exist
func (cm *ConfigManager) LoadConfig() (DetectionConfig, error) {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultDetectionConfig(), nil
	}

	data, err := ioutil.ReadFile(cm.configPath)
	if err != nil {
		return DetectionConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	var config DetectionConfigFile
	err = json.Unmarshal(data, &config)
	if err != nil {
		return DetectionConfig{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config.ToDetectionConfig(), nil
}

// SaveConfig saves configuration to file
func (cm *ConfigManager) SaveConfig(config DetectionConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configFile := DetectionConfigFileFromConfig(config)
	data, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = ioutil.WriteFile(cm.configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DetectionConfigFile represents the JSON structure for configuration files
type DetectionConfigFile struct {
	AttributeConfigs  map[string]AttributeConfigFile `json:"attribute_configs"`
	DefaultConfig     AttributeConfigFile            `json:"default_config"`
	IgnoredAttributes []string                       `json:"ignored_attributes"`
	StrictMode        bool                           `json:"strict_mode"`
	MaxConcurrency    int                            `json:"max_concurrency"`
	TimeoutSeconds    int                            `json:"timeout_seconds"`
	Extensions        ExtensionConfig                `json:"extensions,omitempty"`
}

// AttributeConfigFile represents the JSON structure for attribute configurations
type AttributeConfigFile struct {
	ComparisonType string   `json:"comparison_type"`
	CaseSensitive  bool     `json:"case_sensitive"`
	Tolerance      *float64 `json:"tolerance,omitempty"`
}

// ExtensionConfig holds configuration for extending drift detection
type ExtensionConfig struct {
	// CustomComparators maps attribute names to custom comparison function names
	CustomComparators map[string]string `json:"custom_comparators,omitempty"`

	// PreProcessors lists functions to run before drift detection
	PreProcessors []string `json:"pre_processors,omitempty"`

	// PostProcessors lists functions to run after drift detection
	PostProcessors []string `json:"post_processors,omitempty"`

	// SeverityRules maps attribute names to custom severity levels
	SeverityRules map[string]string `json:"severity_rules,omitempty"`

	// NotificationHooks lists webhook URLs for drift notifications
	NotificationHooks []string `json:"notification_hooks,omitempty"`

	// ReportTemplates maps report format names to template file paths
	ReportTemplates map[string]string `json:"report_templates,omitempty"`
}

// ToDetectionConfig converts DetectionConfigFile to DetectionConfig
func (dcf DetectionConfigFile) ToDetectionConfig() DetectionConfig {
	attributeConfigs := make(map[string]AttributeConfig)
	for name, config := range dcf.AttributeConfigs {
		attributeConfigs[name] = config.ToAttributeConfig()
	}

	timeout := time.Duration(dcf.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return DetectionConfig{
		AttributeConfigs:  attributeConfigs,
		DefaultConfig:     dcf.DefaultConfig.ToAttributeConfig(),
		IgnoredAttributes: dcf.IgnoredAttributes,
		StrictMode:        dcf.StrictMode,
		MaxConcurrency:    dcf.MaxConcurrency,
		Timeout:           timeout,
	}
}

// ToAttributeConfig converts AttributeConfigFile to AttributeConfig
func (acf AttributeConfigFile) ToAttributeConfig() AttributeConfig {
	comparisonType := parseComparisonType(acf.ComparisonType)
	return AttributeConfig{
		ComparisonType: comparisonType,
		CaseSensitive:  acf.CaseSensitive,
		Tolerance:      acf.Tolerance,
	}
}

// DetectionConfigFileFromConfig converts DetectionConfig to DetectionConfigFile
func DetectionConfigFileFromConfig(config DetectionConfig) DetectionConfigFile {
	attributeConfigs := make(map[string]AttributeConfigFile)
	for name, attrConfig := range config.AttributeConfigs {
		attributeConfigs[name] = AttributeConfigFileFromConfig(attrConfig)
	}

	timeoutSeconds := int(config.Timeout.Seconds())
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	return DetectionConfigFile{
		AttributeConfigs:  attributeConfigs,
		DefaultConfig:     AttributeConfigFileFromConfig(config.DefaultConfig),
		IgnoredAttributes: config.IgnoredAttributes,
		StrictMode:        config.StrictMode,
		MaxConcurrency:    config.MaxConcurrency,
		TimeoutSeconds:    timeoutSeconds,
	}
}

// AttributeConfigFileFromConfig converts AttributeConfig to AttributeConfigFile
func AttributeConfigFileFromConfig(config AttributeConfig) AttributeConfigFile {
	return AttributeConfigFile{
		ComparisonType: comparisonTypeToString(config.ComparisonType),
		CaseSensitive:  config.CaseSensitive,
		Tolerance:      config.Tolerance,
	}
}

// parseComparisonType converts string to ComparisonType
func parseComparisonType(s string) ComparisonType {
	switch s {
	case "exact_match":
		return ExactMatch
	case "fuzzy_match":
		return FuzzyMatch
	case "numeric_tolerance":
		return NumericTolerance
	case "array_ordered":
		return ArrayOrdered
	case "array_unordered":
		return ArrayUnordered
	case "map_comparison":
		return MapComparison
	case "nested_object":
		return NestedObject
	default:
		return ExactMatch
	}
}

// comparisonTypeToString converts ComparisonType to string
func comparisonTypeToString(ct ComparisonType) string {
	switch ct {
	case ExactMatch:
		return "exact_match"
	case FuzzyMatch:
		return "fuzzy_match"
	case NumericTolerance:
		return "numeric_tolerance"
	case ArrayOrdered:
		return "array_ordered"
	case ArrayUnordered:
		return "array_unordered"
	case MapComparison:
		return "map_comparison"
	case NestedObject:
		return "nested_object"
	default:
		return "exact_match"
	}
}

// ConfigValidator validates drift detection configurations
type ConfigValidator struct{}

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// ValidateConfig validates a DetectionConfig
func (cv *ConfigValidator) ValidateConfig(config DetectionConfig) error {
	if config.MaxConcurrency <= 0 {
		return fmt.Errorf("max_concurrency must be positive, got %d", config.MaxConcurrency)
	}

	if config.MaxConcurrency > 100 {
		return fmt.Errorf("max_concurrency too high (max 100), got %d", config.MaxConcurrency)
	}

	if config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive, got %v", config.Timeout)
	}

	if config.Timeout > 5*time.Minute {
		return fmt.Errorf("timeout too high (max 5 minutes), got %v", config.Timeout)
	}

	// Validate attribute configurations
	for attrName, attrConfig := range config.AttributeConfigs {
		if err := cv.validateAttributeConfig(attrName, attrConfig); err != nil {
			return fmt.Errorf("invalid config for attribute '%s': %w", attrName, err)
		}
	}

	// Validate default configuration
	if err := cv.validateAttributeConfig("default", config.DefaultConfig); err != nil {
		return fmt.Errorf("invalid default config: %w", err)
	}

	return nil
}

// validateAttributeConfig validates an AttributeConfig
func (cv *ConfigValidator) validateAttributeConfig(attrName string, config AttributeConfig) error {
	// Validate comparison type
	validTypes := []ComparisonType{
		ExactMatch, FuzzyMatch, NumericTolerance,
		ArrayOrdered, ArrayUnordered, MapComparison, NestedObject,
	}

	validType := false
	for _, validT := range validTypes {
		if config.ComparisonType == validT {
			validType = true
			break
		}
	}

	if !validType {
		return fmt.Errorf("invalid comparison type: %v", config.ComparisonType)
	}

	// Validate tolerance for numeric comparison
	if config.ComparisonType == NumericTolerance {
		if config.Tolerance == nil {
			return fmt.Errorf("tolerance is required for numeric_tolerance comparison")
		}
		if *config.Tolerance < 0 {
			return fmt.Errorf("tolerance must be non-negative, got %f", *config.Tolerance)
		}
		if *config.Tolerance > 1000000 {
			return fmt.Errorf("tolerance too high (max 1000000), got %f", *config.Tolerance)
		}
	}

	return nil
}

// GetStrictConfig returns a strict configuration for production environments
func GetStrictConfig() DetectionConfig {
	config := DefaultDetectionConfig()
	config.StrictMode = true
	config.MaxConcurrency = 5
	config.Timeout = 60 * time.Second
	return config
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./drift-config.json"
	}
	return filepath.Join(homeDir, ".firefly", "drift-config.json")
}

// GetConfigPathFromEnv returns configuration path from environment variable or default
func GetConfigPathFromEnv() string {
	if path := os.Getenv("FIREFLY_DRIFT_CONFIG"); path != "" {
		return path
	}
	return GetConfigPath()
}
