package interfaces



// TerraformParser defines the interface for parsing Terraform configurations
type TerraformParser interface {
	// ParseTerraformState parses a Terraform state file and returns the configuration
	ParseTerraformState(filePath string) (map[string]*TerraformConfig, error)

	// ParseTerraformHCL parses Terraform HCL files in a directory and returns the configuration
	ParseTerraformHCL(dirPath string) (map[string]*TerraformConfig, error)

	// ValidateStateFile validates that the state file is valid and readable
	ValidateStateFile(filePath string) error

	// ValidateHCLDirectory validates that the HCL directory contains valid Terraform files
	ValidateHCLDirectory(dirPath string) error
}

// TerraformExecutor defines the interface for executing Terraform commands
type TerraformExecutor interface {
	// Init initializes a Terraform working directory
	Init(workingDir string) error

	// Plan creates an execution plan
	Plan(workingDir string, planFile string) error

	// Apply applies the changes required to reach the desired state
	Apply(workingDir string, planFile string) error

	// Show displays the current state or a saved plan
	Show(workingDir string, target string) (string, error)

	// Validate validates the configuration files
	Validate(workingDir string) error
}

// TerraformStateManager defines the interface for managing Terraform state
type TerraformStateManager interface {
	// GetResourceState retrieves the state of a specific resource
	GetResourceState(resourceAddress string) (*ResourceState, error)

	// ListResources lists all resources in the state
	ListResources() ([]*ResourceState, error)

	// RefreshState refreshes the state to match real infrastructure
	RefreshState(workingDir string) error
}