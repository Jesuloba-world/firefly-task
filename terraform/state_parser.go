package terraform

import (
	"encoding/json"
	"fmt"
	"os"

	tfjson "github.com/hashicorp/terraform-json"
)

// ParseTerraformState parses a Terraform state file using terraform-json
func ParseTerraformState(statePath string) (*tfjson.State, error) {
	if statePath == "" {
		return nil, fmt.Errorf("state file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("state file does not exist: %s", statePath)
	}

	// Read the state file
	stateData, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Parse the state using terraform-json
	var state tfjson.State
	err = json.Unmarshal(stateData, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// ExtractEC2InstancesFromState extracts EC2 instance configurations from Terraform state
func ExtractEC2InstancesFromState(state *tfjson.State) ([]EC2InstanceConfig, error) {
	var instances []EC2InstanceConfig

	if state.Values == nil || state.Values.RootModule == nil {
		return instances, nil
	}

	// Process resources in the root module
	instances = append(instances, extractInstancesFromModule(state.Values.RootModule)...)

	// Process child modules
	for _, childModule := range state.Values.RootModule.ChildModules {
		instances = append(instances, extractInstancesFromModule(childModule)...)
	}

	return instances, nil
}

// extractInstancesFromModule extracts EC2 instances from a specific module
func extractInstancesFromModule(module *tfjson.StateModule) []EC2InstanceConfig {
	var instances []EC2InstanceConfig

	for _, resource := range module.Resources {
		if resource.Type == "aws_instance" {
			instance := EC2InstanceConfig{
				ResourceName: resource.Name,
			}

			// Extract values from the resource attributes
			if resource.AttributeValues != nil {
				// Instance type
				if instanceType, ok := resource.AttributeValues["instance_type"].(string); ok {
					instance.InstanceType = instanceType
				}

				// AMI
				if ami, ok := resource.AttributeValues["ami"].(string); ok {
					instance.AMI = ami
				}

				// Subnet ID
				if subnetID, ok := resource.AttributeValues["subnet_id"].(string); ok {
					instance.SubnetID = subnetID
				}

				// Key name
				if keyName, ok := resource.AttributeValues["key_name"].(string); ok {
					instance.KeyName = keyName
				}

				// User data
				if userData, ok := resource.AttributeValues["user_data"].(string); ok {
					instance.UserData = userData
				}

				// VPC Security Groups
				if secGroups, ok := resource.AttributeValues["vpc_security_group_ids"].([]interface{}); ok {
					for _, sg := range secGroups {
						if sgStr, ok := sg.(string); ok {
							instance.VPCSecurityGroups = append(instance.VPCSecurityGroups, sgStr)
						}
					}
				}

				// Tags
				if tags, ok := resource.AttributeValues["tags"].(map[string]interface{}); ok {
					instance.Tags = make(map[string]string)
					for k, v := range tags {
						if tagValue, ok := v.(string); ok {
							instance.Tags[k] = tagValue
						}
					}
				}
			}

			instances = append(instances, instance)
		}
	}

	// Process child modules recursively
	for _, childModule := range module.ChildModules {
		instances = append(instances, extractInstancesFromModule(childModule)...)
	}

	return instances
}

// GetResourceByAddress finds a specific resource by its address in the state
func GetResourceByAddress(state *tfjson.State, address string) (*tfjson.StateResource, error) {
	if state.Values == nil || state.Values.RootModule == nil {
		return nil, fmt.Errorf("state has no values")
	}

	// Search in root module
	if resource := findResourceInModule(state.Values.RootModule, address); resource != nil {
		return resource, nil
	}

	// Search in child modules
	for _, childModule := range state.Values.RootModule.ChildModules {
		if resource := findResourceInModule(childModule, address); resource != nil {
			return resource, nil
		}
	}

	return nil, fmt.Errorf("resource not found: %s", address)
}

// findResourceInModule searches for a resource in a specific module
func findResourceInModule(module *tfjson.StateModule, address string) *tfjson.StateResource {
	for _, resource := range module.Resources {
		if resource.Address == address {
			return resource
		}
	}

	// Search in child modules recursively
	for _, childModule := range module.ChildModules {
		if resource := findResourceInModule(childModule, address); resource != nil {
			return resource
		}
	}

	return nil
}

// ValidateStateFile validates that a file is a valid Terraform state file
func ValidateStateFile(statePath string) error {
	state, err := ParseTerraformState(statePath)
	if err != nil {
		return err
	}

	// Basic validation checks
	if state.FormatVersion == "" {
		return fmt.Errorf("invalid state file: missing format_version")
	}

	if state.TerraformVersion == "" {
		return fmt.Errorf("invalid state file: missing terraform_version")
	}

	return nil
}

// GetStateMetadata extracts metadata from the Terraform state
func GetStateMetadata(state *tfjson.State) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["format_version"] = state.FormatVersion
	metadata["terraform_version"] = state.TerraformVersion

	if state.Values != nil && state.Values.RootModule != nil {
		resourceCount := countResources(state.Values.RootModule)
		metadata["resource_count"] = resourceCount
	}

	return metadata
}

// countResources counts the total number of resources in a module and its children
func countResources(module *tfjson.StateModule) int {
	count := len(module.Resources)

	for _, childModule := range module.ChildModules {
		count += countResources(childModule)
	}

	return count
}
