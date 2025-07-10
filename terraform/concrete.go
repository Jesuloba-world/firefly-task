package terraform

import (
	"fmt"
	"os"
	"path/filepath"

	"firefly-task/pkg/interfaces"
	"github.com/sirupsen/logrus"
)

// ConcreteTerraformParser implements the interfaces.TerraformParser interface
type ConcreteTerraformParser struct {
	parser *TerraformParser
	logger *logrus.Logger
}

// ConcreteTerraformExecutor implements the interfaces.TerraformExecutor interface
type ConcreteTerraformExecutor struct {
	workingDir string
	logger     *logrus.Logger
}

// ConcreteTerraformStateManager implements the interfaces.TerraformStateManager interface
type ConcreteTerraformStateManager struct {
	statePath string
	logger    *logrus.Logger
}

// NewConcreteTerraformParser creates a new concrete Terraform parser
func NewConcreteTerraformParser(logger *logrus.Logger) interfaces.TerraformParser {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	parser := NewParser()
	return &ConcreteTerraformParser{
		parser: parser,
		logger: logger,
	}
}

// NewConcreteTerraformExecutor creates a new concrete Terraform executor
func NewConcreteTerraformExecutor(workingDir string, logger *logrus.Logger) interfaces.TerraformExecutor {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteTerraformExecutor{
		workingDir: workingDir,
		logger:     logger,
	}
}

// NewConcreteTerraformStateManager creates a new concrete Terraform state manager
func NewConcreteTerraformStateManager(statePath string, logger *logrus.Logger) interfaces.TerraformStateManager {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteTerraformStateManager{
		statePath: statePath,
		logger:    logger,
	}
}

// TerraformParser implementation methods

// ParseTerraformState parses a Terraform state file and returns the configuration
func (p *ConcreteTerraformParser) ParseTerraformState(filePath string) (map[string]*interfaces.TerraformConfig, error) {
	p.logger.Debugf("ConcreteTerraformParser: Parsing state file %s", filePath)
	return p.parser.ParseTerraformState(filePath)
}

func (p *ConcreteTerraformParser) ParseTerraformHCL(dirPath string) (map[string]*interfaces.TerraformConfig, error) {
	p.logger.Debugf("ConcreteTerraformParser: Parsing HCL files in %s", dirPath)
	return p.parser.ParseTerraformHCL(dirPath)
}

// ValidateStateFile validates that the state file is valid and readable
func (p *ConcreteTerraformParser) ValidateStateFile(filePath string) error {
	p.logger.Debugf("ConcreteTerraformParser: Validating state file %s", filePath)
	return p.parser.ValidateStateFile(filePath)
}

// ValidateHCLDirectory validates that the HCL directory contains valid Terraform files
func (p *ConcreteTerraformParser) ValidateHCLDirectory(dirPath string) error {
	p.logger.Debugf("ConcreteTerraformParser: Validating HCL directory %s", dirPath)
	return p.parser.ValidateHCLDirectory(dirPath)
}

// ParsePlanFile parses a Terraform plan file and returns configurations
func (p *ConcreteTerraformParser) ParsePlanFile(filePath string) (map[string]*interfaces.TerraformConfig, error) {
	p.logger.Debugf("ConcreteTerraformParser: Parsing plan file %s", filePath)
	// This method would need to be implemented in the underlying parser
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("ParsePlanFile not yet implemented")
}

// ValidateConfiguration validates a Terraform configuration
func (p *ConcreteTerraformParser) ValidateConfiguration(configPath string) error {
	p.logger.Debugf("ConcreteTerraformParser: Validating configuration at %s", configPath)
	// Basic validation - check if path exists and contains .tf files
	info, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("configuration path does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("configuration path must be a directory")
	}

	// Check for .tf files
	hasTerraformFiles := false
	err = filepath.Walk(configPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".tf" {
			hasTerraformFiles = true
			return filepath.SkipDir // Stop walking once we find a .tf file
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error scanning configuration directory: %w", err)
	}

	if !hasTerraformFiles {
		return fmt.Errorf("no Terraform files (.tf) found in configuration directory")
	}

	return nil
}

// TerraformExecutor implementation methods

// Init initializes a Terraform working directory
func (e *ConcreteTerraformExecutor) Init(workingDir string) error {
	e.logger.Debug("ConcreteTerraformExecutor: Initializing Terraform")
	// Placeholder implementation - actual terraform execution not yet implemented
	return nil
}

// Plan creates an execution plan
func (e *ConcreteTerraformExecutor) Plan(workingDir string, planFile string) error {
	e.logger.Debugf("ConcreteTerraformExecutor: Creating plan file %s", planFile)
	// Placeholder implementation - actual terraform execution not yet implemented
	return nil
}

// Apply applies the changes required to reach the desired state
func (e *ConcreteTerraformExecutor) Apply(workingDir string, planFile string) error {
	e.logger.Debugf("ConcreteTerraformExecutor: Applying plan file %s", planFile)
	// Placeholder implementation - actual terraform execution not yet implemented
	return nil
}

// Show displays the current state or a saved plan
func (e *ConcreteTerraformExecutor) Show(workingDir string, target string) (string, error) {
	e.logger.Debugf("ConcreteTerraformExecutor: Showing %s", target)
	// Placeholder implementation - actual terraform execution not yet implemented
	return "", nil
}

// Validate validates the configuration files
func (e *ConcreteTerraformExecutor) Validate(workingDir string) error {
	e.logger.Debug("ConcreteTerraformExecutor: Validating Terraform configuration")
	// Placeholder implementation - actual terraform execution not yet implemented
	return nil
}

// TerraformStateManager implementation methods

// GetResourceState retrieves the state of a specific resource
func (s *ConcreteTerraformStateManager) GetResourceState(resourceAddress string) (*interfaces.ResourceState, error) {
	s.logger.Debugf("ConcreteTerraformStateManager: Getting resource state for %s", resourceAddress)
	// Placeholder implementation - state management not yet implemented
	return nil, fmt.Errorf("state management not yet implemented")
}

// ListResources lists all resources in the state
func (s *ConcreteTerraformStateManager) ListResources() ([]*interfaces.ResourceState, error) {
	s.logger.Debug("ConcreteTerraformStateManager: Listing all resource states")
	// Placeholder implementation - state management not yet implemented
	return nil, fmt.Errorf("state management not yet implemented")
}

// RefreshState refreshes the state to match real infrastructure
func (s *ConcreteTerraformStateManager) RefreshState(workingDir string) error {
	s.logger.Debug("ConcreteTerraformStateManager: Refreshing state")
	// Placeholder implementation - state management not yet implemented
	return fmt.Errorf("state management not yet implemented")
}