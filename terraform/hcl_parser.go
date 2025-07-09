package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// ParsedTerraformConfig represents the parsed Terraform configuration from terraform-config-inspect
type ParsedTerraformConfig struct {
	Resources []TerraformResource    `json:"resources"`
	Variables map[string]interface{} `json:"variables,omitempty"`
	Outputs   map[string]interface{} `json:"outputs,omitempty"`
}

// TerraformResource represents a Terraform resource
type TerraformResource struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider,omitempty"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

// ParseTerraformHCL parses Terraform configuration files using terraform-config-inspect
func ParseTerraformHCL(configPath string) (*ParsedTerraformConfig, error) {
	// Check if path is a file or directory
	info, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to access path %s: %w", configPath, err)
	}

	var modulePath string
	if info.IsDir() {
		modulePath = configPath
	} else {
		// If it's a file, use its directory
		modulePath = filepath.Dir(configPath)
	}

	// Load the module using terraform-config-inspect
	module, diags := tfconfig.LoadModule(modulePath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to load Terraform module: %s", diags.Error())
	}

	// Convert to our internal format
	config := &ParsedTerraformConfig{
		Resources: make([]TerraformResource, 0),
		Variables: make(map[string]interface{}),
		Outputs:   make(map[string]interface{}),
	}

	// Process managed resources
	for _, resource := range module.ManagedResources {
		tfResource := TerraformResource{
			Type:         resource.Type,
			Name:         resource.Name,
			Provider:     resource.Provider.Name,
			Config:       make(map[string]interface{}), // terraform-config-inspect doesn't expose config details
			Dependencies: []string{},                   // terraform-config-inspect doesn't expose dependencies
		}
		config.Resources = append(config.Resources, tfResource)
	}

	// Process variables
	for name, variable := range module.Variables {
		config.Variables[name] = map[string]interface{}{
			"description": variable.Description,
			"type":        variable.Type,
			"default":     variable.Default,
			// Note: terraform-config-inspect doesn't provide a Required field
			// Required status is inferred from presence of default value
			"required": variable.Default == nil,
		}
	}

	// Process outputs
	for name, output := range module.Outputs {
		config.Outputs[name] = map[string]interface{}{
			"description": output.Description,
			"sensitive":   output.Sensitive,
		}
	}

	return config, nil
}

// Note: terraform-config-inspect provides high-level metadata only
// It doesn't expose detailed configuration attributes like the old HCL parser
// This is a limitation of the library's design for broad compatibility

// ExtractEC2Instances extracts EC2 instance configurations from parsed Terraform config
func ExtractEC2Instances(config *ParsedTerraformConfig) ([]EC2InstanceConfig, error) {
	var instances []EC2InstanceConfig

	for _, resource := range config.Resources {
		if resource.Type == "aws_instance" {
			// Note: terraform-config-inspect doesn't expose detailed config attributes
			// We create a basic instance config with the resource name
			instance := EC2InstanceConfig{
				ResourceName: resource.Name,
				// Set default values since config details aren't available
				InstanceType: "unknown", // Will be populated from state file if available
				AMI:          "unknown", // Will be populated from state file if available
				Tags:         make(map[string]string),
			}

			instances = append(instances, instance)
		}
	}

	return instances, nil
}

// ParseTerraformFile is a convenience function for parsing a single Terraform file
func ParseTerraformFile(filePath string) (*ParsedTerraformConfig, error) {
	if !strings.HasSuffix(filePath, ".tf") && !strings.HasSuffix(filePath, ".tf.json") {
		return nil, fmt.Errorf("file %s is not a Terraform configuration file", filePath)
	}

	return ParseTerraformHCL(filePath)
}

// ToJSON converts the ParsedTerraformConfig to JSON string
func (tc *ParsedTerraformConfig) ToJSON() (string, error) {
	data, err := json.MarshalIndent(tc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(data), nil
}
