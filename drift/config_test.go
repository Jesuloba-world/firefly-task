package drift

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewConfigManager(t *testing.T) {
	configPath := "/tmp/test-config.json"
	cm := NewConfigManager(configPath)

	if cm == nil {
		t.Fatal("NewConfigManager returned nil")
	}

	if cm.configPath != configPath {
		t.Errorf("Expected configPath %s, got %s", configPath, cm.configPath)
	}
}

func TestConfigManager_LoadConfig_DefaultWhenFileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent-config.json")
	cm := NewConfigManager(configPath)

	config, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Should return default config
	defaultConfig := DefaultDetectionConfig()
	if config.MaxConcurrency != defaultConfig.MaxConcurrency {
		t.Errorf("Expected MaxConcurrency %d, got %d", defaultConfig.MaxConcurrency, config.MaxConcurrency)
	}

	if config.StrictMode != defaultConfig.StrictMode {
		t.Errorf("Expected StrictMode %v, got %v", defaultConfig.StrictMode, config.StrictMode)
	}
}

func TestConfigManager_SaveAndLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")
	cm := NewConfigManager(configPath)

	// Create a custom config
	originalConfig := DefaultDetectionConfig()
	originalConfig.MaxConcurrency = 25
	originalConfig.StrictMode = true
	originalConfig.Timeout = 45 * time.Second
	originalConfig.IgnoredAttributes = []string{"custom_ignored"}

	// Add custom attribute config
	tolerance := 0.5
	originalConfig.AttributeConfigs["custom_attr"] = AttributeConfig{
		ComparisonType: NumericTolerance,
		CaseSensitive:  false,
		Tolerance:      &tolerance,
	}

	// Save config
	err := cm.SaveConfig(originalConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedConfig, err := cm.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify loaded config matches original
	if loadedConfig.MaxConcurrency != originalConfig.MaxConcurrency {
		t.Errorf("Expected MaxConcurrency %d, got %d", originalConfig.MaxConcurrency, loadedConfig.MaxConcurrency)
	}

	if loadedConfig.StrictMode != originalConfig.StrictMode {
		t.Errorf("Expected StrictMode %v, got %v", originalConfig.StrictMode, loadedConfig.StrictMode)
	}

	if loadedConfig.Timeout != originalConfig.Timeout {
		t.Errorf("Expected Timeout %v, got %v", originalConfig.Timeout, loadedConfig.Timeout)
	}

	if len(loadedConfig.IgnoredAttributes) != len(originalConfig.IgnoredAttributes) {
		t.Errorf("Expected %d ignored attributes, got %d", len(originalConfig.IgnoredAttributes), len(loadedConfig.IgnoredAttributes))
	}

	// Verify custom attribute config
	customAttr, exists := loadedConfig.AttributeConfigs["custom_attr"]
	if !exists {
		t.Error("Custom attribute config not found")
	} else {
		if customAttr.ComparisonType != NumericTolerance {
			t.Errorf("Expected NumericTolerance, got %v", customAttr.ComparisonType)
		}
		if customAttr.CaseSensitive {
			t.Error("Expected CaseSensitive to be false")
		}
		if customAttr.Tolerance == nil || *customAttr.Tolerance != 0.5 {
			t.Errorf("Expected tolerance 0.5, got %v", customAttr.Tolerance)
		}
	}
}

func TestConfigManager_LoadConfig_InvalidJSON(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "invalid-config.json")

	// Write invalid JSON
	err := ioutil.WriteFile(configPath, []byte("invalid json content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cm := NewConfigManager(configPath)
	_, err = cm.LoadConfig()
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}

func TestDetectionConfigFile_ToDetectionConfig(t *testing.T) {
	tolerance := 0.1
	configFile := DetectionConfigFile{
		AttributeConfigs: map[string]AttributeConfigFile{
			"test_attr": {
				ComparisonType: "numeric_tolerance",
				CaseSensitive:  false,
				Tolerance:      &tolerance,
			},
		},
		DefaultConfig: AttributeConfigFile{
			ComparisonType: "exact_match",
			CaseSensitive:  true,
		},
		IgnoredAttributes: []string{"ignored1", "ignored2"},
		StrictMode:        true,
		MaxConcurrency:    15,
		TimeoutSeconds:    60,
	}

	config := configFile.ToDetectionConfig()

	if config.StrictMode != true {
		t.Error("Expected StrictMode to be true")
	}

	if config.MaxConcurrency != 15 {
		t.Errorf("Expected MaxConcurrency 15, got %d", config.MaxConcurrency)
	}

	if config.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout 60s, got %v", config.Timeout)
	}

	if len(config.IgnoredAttributes) != 2 {
		t.Errorf("Expected 2 ignored attributes, got %d", len(config.IgnoredAttributes))
	}

	testAttr, exists := config.AttributeConfigs["test_attr"]
	if !exists {
		t.Error("test_attr not found in AttributeConfigs")
	} else {
		if testAttr.ComparisonType != NumericTolerance {
			t.Errorf("Expected NumericTolerance, got %v", testAttr.ComparisonType)
		}
		if testAttr.CaseSensitive {
			t.Error("Expected CaseSensitive to be false")
		}
		if testAttr.Tolerance == nil || *testAttr.Tolerance != 0.1 {
			t.Errorf("Expected tolerance 0.1, got %v", testAttr.Tolerance)
		}
	}

	if config.DefaultConfig.ComparisonType != ExactMatch {
		t.Errorf("Expected default ExactMatch, got %v", config.DefaultConfig.ComparisonType)
	}
}

func TestDetectionConfigFileFromConfig(t *testing.T) {
	tolerance := 0.2
	config := DetectionConfig{
		AttributeConfigs: map[string]AttributeConfig{
			"test_attr": {
				ComparisonType: FuzzyMatch,
				CaseSensitive:  false,
				Tolerance:      &tolerance,
			},
		},
		DefaultConfig: AttributeConfig{
			ComparisonType: ArrayOrdered,
			CaseSensitive:  true,
		},
		IgnoredAttributes: []string{"ignored1"},
		StrictMode:        false,
		MaxConcurrency:    20,
		Timeout:           45 * time.Second,
	}

	configFile := DetectionConfigFileFromConfig(config)

	if configFile.StrictMode != false {
		t.Error("Expected StrictMode to be false")
	}

	if configFile.MaxConcurrency != 20 {
		t.Errorf("Expected MaxConcurrency 20, got %d", configFile.MaxConcurrency)
	}

	if configFile.TimeoutSeconds != 45 {
		t.Errorf("Expected TimeoutSeconds 45, got %d", configFile.TimeoutSeconds)
	}

	testAttr, exists := configFile.AttributeConfigs["test_attr"]
	if !exists {
		t.Error("test_attr not found in AttributeConfigs")
	} else {
		if testAttr.ComparisonType != "fuzzy_match" {
			t.Errorf("Expected fuzzy_match, got %s", testAttr.ComparisonType)
		}
		if testAttr.CaseSensitive {
			t.Error("Expected CaseSensitive to be false")
		}
		if testAttr.Tolerance == nil || *testAttr.Tolerance != 0.2 {
			t.Errorf("Expected tolerance 0.2, got %v", testAttr.Tolerance)
		}
	}

	if configFile.DefaultConfig.ComparisonType != "array_ordered" {
		t.Errorf("Expected array_ordered, got %s", configFile.DefaultConfig.ComparisonType)
	}
}

func TestParseComparisonType(t *testing.T) {
	tests := []struct {
		input    string
		expected ComparisonType
	}{
		{"exact_match", ExactMatch},
		{"fuzzy_match", FuzzyMatch},
		{"numeric_tolerance", NumericTolerance},
		{"array_ordered", ArrayOrdered},
		{"array_unordered", ArrayUnordered},
		{"map_comparison", MapComparison},
		{"nested_object", NestedObject},
		{"invalid_type", ExactMatch}, // Should default to ExactMatch
		{"", ExactMatch},             // Should default to ExactMatch
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseComparisonType(tt.input)
			if result != tt.expected {
				t.Errorf("parseComparisonType(%s) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestComparisonTypeToString(t *testing.T) {
	tests := []struct {
		input    ComparisonType
		expected string
	}{
		{ExactMatch, "exact_match"},
		{FuzzyMatch, "fuzzy_match"},
		{NumericTolerance, "numeric_tolerance"},
		{ArrayOrdered, "array_ordered"},
		{ArrayUnordered, "array_unordered"},
		{MapComparison, "map_comparison"},
		{NestedObject, "nested_object"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := comparisonTypeToString(tt.input)
			if result != tt.expected {
				t.Errorf("comparisonTypeToString(%v) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConfigValidator_ValidateConfig(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name      string
		config    DetectionConfig
		wantError bool
	}{
		{
			name:      "valid default config",
			config:    DefaultDetectionConfig(),
			wantError: false,
		},
		{
			name: "invalid max concurrency - zero",
			config: DetectionConfig{
				MaxConcurrency: 0,
				Timeout:        30 * time.Second,
				DefaultConfig:  AttributeConfig{ComparisonType: ExactMatch},
			},
			wantError: true,
		},
		{
			name: "invalid max concurrency - too high",
			config: DetectionConfig{
				MaxConcurrency: 150,
				Timeout:        30 * time.Second,
				DefaultConfig:  AttributeConfig{ComparisonType: ExactMatch},
			},
			wantError: true,
		},
		{
			name: "invalid timeout - zero",
			config: DetectionConfig{
				MaxConcurrency: 10,
				Timeout:        0,
				DefaultConfig:  AttributeConfig{ComparisonType: ExactMatch},
			},
			wantError: true,
		},
		{
			name: "invalid timeout - too high",
			config: DetectionConfig{
				MaxConcurrency: 10,
				Timeout:        10 * time.Minute,
				DefaultConfig:  AttributeConfig{ComparisonType: ExactMatch},
			},
			wantError: true,
		},
		{
			name: "invalid attribute config - missing tolerance",
			config: DetectionConfig{
				MaxConcurrency: 10,
				Timeout:        30 * time.Second,
				DefaultConfig:  AttributeConfig{ComparisonType: ExactMatch},
				AttributeConfigs: map[string]AttributeConfig{
					"test_attr": {ComparisonType: NumericTolerance}, // Missing tolerance
				},
			},
			wantError: true,
		},
		{
			name: "valid attribute config with tolerance",
			config: func() DetectionConfig {
				tolerance := 0.1
				return DetectionConfig{
					MaxConcurrency: 10,
					Timeout:        30 * time.Second,
					DefaultConfig:  AttributeConfig{ComparisonType: ExactMatch},
					AttributeConfigs: map[string]AttributeConfig{
						"test_attr": {
							ComparisonType: NumericTolerance,
							Tolerance:      &tolerance,
						},
					},
				}
			}(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateConfig(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestConfigValidator_ValidateAttributeConfig(t *testing.T) {
	validator := NewConfigValidator()

	tests := []struct {
		name      string
		attrName  string
		config    AttributeConfig
		wantError bool
	}{
		{
			name:      "valid exact match",
			attrName:  "test_attr",
			config:    AttributeConfig{ComparisonType: ExactMatch},
			wantError: false,
		},
		{
			name:     "valid numeric tolerance",
			attrName: "test_attr",
			config: func() AttributeConfig {
				tolerance := 0.5
				return AttributeConfig{
					ComparisonType: NumericTolerance,
					Tolerance:      &tolerance,
				}
			}(),
			wantError: false,
		},
		{
			name:      "invalid - missing tolerance for numeric",
			attrName:  "test_attr",
			config:    AttributeConfig{ComparisonType: NumericTolerance},
			wantError: true,
		},
		{
			name:     "invalid - negative tolerance",
			attrName: "test_attr",
			config: func() AttributeConfig {
				tolerance := -0.1
				return AttributeConfig{
					ComparisonType: NumericTolerance,
					Tolerance:      &tolerance,
				}
			}(),
			wantError: true,
		},
		{
			name:     "invalid - tolerance too high",
			attrName: "test_attr",
			config: func() AttributeConfig {
				tolerance := 2000000.0
				return AttributeConfig{
					ComparisonType: NumericTolerance,
					Tolerance:      &tolerance,
				}
			}(),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateAttributeConfig(tt.attrName, tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validateAttributeConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestConfigPresets removed - NewConfigPresets function was deleted

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// Should contain .firefly directory
	if !contains(path, ".firefly") {
		t.Error("Config path should contain .firefly directory")
	}

	// Should end with drift-config.json
	if !contains(path, "drift-config.json") {
		t.Error("Config path should end with drift-config.json")
	}
}

func TestGetConfigPathFromEnv(t *testing.T) {
	// Test with environment variable set
	customPath := "/custom/path/config.json"
	os.Setenv("FIREFLY_DRIFT_CONFIG", customPath)
	defer os.Unsetenv("FIREFLY_DRIFT_CONFIG")

	path := GetConfigPathFromEnv()
	if path != customPath {
		t.Errorf("Expected custom path %s, got %s", customPath, path)
	}

	// Test without environment variable
	os.Unsetenv("FIREFLY_DRIFT_CONFIG")
	path = GetConfigPathFromEnv()
	defaultPath := GetConfigPath()
	if path != defaultPath {
		t.Errorf("Expected default path %s, got %s", defaultPath, path)
	}
}

func TestJSONRoundTrip(t *testing.T) {
	// Test that we can convert config to JSON and back without losing data
	tolerance := 0.123
	originalConfig := DetectionConfig{
		AttributeConfigs: map[string]AttributeConfig{
			"test_attr": {
				ComparisonType: NumericTolerance,
				CaseSensitive:  false,
				Tolerance:      &tolerance,
			},
			"string_attr": {
				ComparisonType: FuzzyMatch,
				CaseSensitive:  true,
			},
		},
		DefaultConfig: AttributeConfig{
			ComparisonType: ExactMatch,
			CaseSensitive:  true,
		},
		IgnoredAttributes: []string{"attr1", "attr2"},
		StrictMode:        true,
		MaxConcurrency:    25,
		Timeout:           45 * time.Second,
	}

	// Convert to file format
	configFile := DetectionConfigFileFromConfig(originalConfig)

	// Convert to JSON
	jsonData, err := json.MarshalIndent(configFile, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal to JSON: %v", err)
	}

	// Parse JSON back
	var parsedConfigFile DetectionConfigFile
	err = json.Unmarshal(jsonData, &parsedConfigFile)
	if err != nil {
		t.Fatalf("Failed to unmarshal from JSON: %v", err)
	}

	// Convert back to config
	resultConfig := parsedConfigFile.ToDetectionConfig()

	// Verify all fields match
	if resultConfig.StrictMode != originalConfig.StrictMode {
		t.Errorf("StrictMode mismatch: expected %v, got %v", originalConfig.StrictMode, resultConfig.StrictMode)
	}

	if resultConfig.MaxConcurrency != originalConfig.MaxConcurrency {
		t.Errorf("MaxConcurrency mismatch: expected %d, got %d", originalConfig.MaxConcurrency, resultConfig.MaxConcurrency)
	}

	if resultConfig.Timeout != originalConfig.Timeout {
		t.Errorf("Timeout mismatch: expected %v, got %v", originalConfig.Timeout, resultConfig.Timeout)
	}

	if len(resultConfig.IgnoredAttributes) != len(originalConfig.IgnoredAttributes) {
		t.Errorf("IgnoredAttributes length mismatch: expected %d, got %d", len(originalConfig.IgnoredAttributes), len(resultConfig.IgnoredAttributes))
	}

	// Check attribute configs
	testAttr := resultConfig.AttributeConfigs["test_attr"]
	originalTestAttr := originalConfig.AttributeConfigs["test_attr"]

	if testAttr.ComparisonType != originalTestAttr.ComparisonType {
		t.Errorf("test_attr ComparisonType mismatch: expected %v, got %v", originalTestAttr.ComparisonType, testAttr.ComparisonType)
	}

	if testAttr.CaseSensitive != originalTestAttr.CaseSensitive {
		t.Errorf("test_attr CaseSensitive mismatch: expected %v, got %v", originalTestAttr.CaseSensitive, testAttr.CaseSensitive)
	}

	if testAttr.Tolerance == nil || originalTestAttr.Tolerance == nil {
		t.Error("Tolerance should not be nil")
	} else if *testAttr.Tolerance != *originalTestAttr.Tolerance {
		t.Errorf("test_attr Tolerance mismatch: expected %f, got %f", *originalTestAttr.Tolerance, *testAttr.Tolerance)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOfSubstring(s, substr) >= 0)))
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
