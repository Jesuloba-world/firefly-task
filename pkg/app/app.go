package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"firefly-task/config"
	"firefly-task/pkg/container"
	"firefly-task/pkg/interfaces"
)

// Default attributes to check for drift detection
var DefaultAttributes = []string{"instance_type", "state", "subnet_id", "vpc_id", "security_groups", "tags"}

// Application represents the main application with all its dependencies
type Application struct {
	// Dependencies
	awsClient       interfaces.EC2Client
	terraformParser interfaces.TerraformParser
	driftDetector   interfaces.DriftDetector
	reportGenerator interfaces.ReportGenerator

	// Configuration
	config *config.Config

	// Lifecycle management
	ctx          context.Context
	cancelFunc   context.CancelFunc
	wg           sync.WaitGroup
	running      bool
	shuttingDown bool
	mu           sync.Mutex
}

// New creates a new application instance with the provided dependencies
func New(cfg *config.Config, awsClient interfaces.EC2Client, terraformParser interfaces.TerraformParser,
	driftDetector interfaces.DriftDetector, reportGenerator interfaces.ReportGenerator) *Application {
	ctx, cancel := context.WithCancel(context.Background())
	return &Application{
		config:          cfg,
		awsClient:       awsClient,
		terraformParser: terraformParser,
		driftDetector:   driftDetector,
		reportGenerator: reportGenerator,
		ctx:             ctx,
		cancelFunc:      cancel,
	}
}

// NewFromContainer creates a new application instance using dependency injection container
func NewFromContainer(cfg *config.Config) (*Application, error) {
	// Initialize dependency injection container
	diContainer := container.NewContainer()
	if err := diContainer.RegisterDefaults(); err != nil {
		return nil, fmt.Errorf("failed to register defaults: %w", err)
	}

	// Get dependencies from container
	ec2Client, err := diContainer.GetEC2Client()
	if err != nil {
		return nil, fmt.Errorf("failed to get EC2 client: %w", err)
	}

	tfParser, err := diContainer.GetTerraformParser()
	if err != nil {
		return nil, fmt.Errorf("failed to get Terraform parser: %w", err)
	}

	driftDetector, err := diContainer.GetDriftDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to get drift detector: %w", err)
	}

	reportGenerator, err := diContainer.GetReportGenerator()
	if err != nil {
		return nil, fmt.Errorf("failed to get report generator: %w", err)
	}

	// Create application instance
	return New(cfg, ec2Client, tfParser, driftDetector, reportGenerator), nil
}

// ReadInstanceIDsFromFile reads instance IDs from a file
func (a *Application) ReadInstanceIDsFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	var instanceIDs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			instanceIDs = append(instanceIDs, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filename, err)
	}

	return instanceIDs, nil
}

// ValidateCheckParameters validates parameters for single instance check
func (a *Application) ValidateCheckParameters(instanceID, terraformPath string) error {
	if instanceID == "" {
		return fmt.Errorf("instance-id is required")
	}
	if terraformPath == "" {
		return fmt.Errorf("tf-path is required")
	}
	return nil
}

// ValidateBatchParameters validates parameters for batch instance check
func (a *Application) ValidateBatchParameters(inputFile, terraformPath string) error {
	if inputFile == "" {
		return fmt.Errorf("input-file is required")
	}
	if terraformPath == "" {
		return fmt.Errorf("tf-path is required")
	}
	return nil
}

// ValidateAttributeParameters validates parameters for attribute-specific check
func (a *Application) ValidateAttributeParameters(instanceID, terraformPath, attribute string) error {
	if instanceID == "" {
		return fmt.Errorf("instance-id is required")
	}
	if terraformPath == "" {
		return fmt.Errorf("tf-path is required")
	}
	if attribute == "" {
		return fmt.Errorf("attribute is required")
	}
	return nil
}

// RunSingleCheck performs a complete single instance drift check workflow
func (a *Application) RunSingleCheck(ctx context.Context, instanceID, terraformPath string, attributes []string) ([]byte, error) {
	// Validate parameters
	if err := a.ValidateCheckParameters(instanceID, terraformPath); err != nil {
		return nil, err
	}

	// Use default attributes if none provided
	if len(attributes) == 0 {
		attributes = DefaultAttributes
	}

	// Run single instance check
	driftResult, err := a.RunSingleInstanceCheck(ctx, instanceID, terraformPath, attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to check instance drift: %w", err)
	}

	if driftResult == nil {
		return nil, fmt.Errorf("instance %s not found in terraform file", instanceID)
	}

	// Generate report
	reportData, err := a.GenerateReport(
			map[string]*interfaces.DriftResult{instanceID: driftResult},
			a.config.Output,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return reportData, nil
}

// RunBatchCheck performs a complete batch instance drift check workflow
func (a *Application) RunBatchCheck(ctx context.Context, inputFile, terraformPath string, attributes []string) ([]byte, error) {
	// Validate parameters
	if err := a.ValidateBatchParameters(inputFile, terraformPath); err != nil {
		return nil, err
	}

	// Use default attributes if none provided
	if len(attributes) == 0 {
		attributes = DefaultAttributes
	}

	// Read instance IDs from input file
	instanceIDs, err := a.ReadInstanceIDsFromFile(inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read instance ids from file: %w", err)
	}

	// Run batch instance check
	driftResults, err := a.RunBatchInstanceCheck(ctx, instanceIDs, terraformPath, attributes)
	if err != nil {
		return nil, fmt.Errorf("failed to check batch instance drift: %w", err)
	}

	// Generate report
	reportData, err := a.GenerateReport(driftResults, a.config.Output)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return reportData, nil
}

// RunAttributeCheck performs a complete attribute-specific drift check workflow
func (a *Application) RunAttributeCheck(ctx context.Context, instanceID, terraformPath, attribute string) ([]byte, error) {
	// Validate parameters
	if err := a.ValidateAttributeParameters(instanceID, terraformPath, attribute); err != nil {
		return nil, err
	}

	// Run single instance check for specific attribute
	driftResult, err := a.RunSingleInstanceCheck(ctx, instanceID, terraformPath, []string{attribute})
	if err != nil {
		return nil, fmt.Errorf("failed to check instance drift: %w", err)
	}

	if driftResult == nil {
		return nil, fmt.Errorf("instance %s not found in terraform file", instanceID)
	}

	// Generate report
	reportData, err := a.GenerateReport(
			map[string]*interfaces.DriftResult{instanceID: driftResult},
			a.config.Output,
		)
	if err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}

	return reportData, nil
}

// RunSingleInstanceCheck performs drift detection on a single EC2 instance
func (a *Application) IsShuttingDown() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.shuttingDown
}

func (a *Application) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

func (a *Application) RunSingleInstanceCheck(ctx context.Context, instanceID, terraformPath string, attributes []string) (*interfaces.DriftResult, error) {
	a.wg.Add(1)
	defer a.wg.Done()

	// Get actual instance state from AWS
	actualInstance, err := a.awsClient.GetEC2Instance(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	// Parse Terraform configuration to get expected state
	expectedInstances, err := a.terraformParser.ParseTerraformHCL(terraformPath)
	if err != nil {
		return nil, err
	}

	// Find the expected instance configuration
	var expectedInstance *interfaces.TerraformConfig
	for _, inst := range expectedInstances {
		if inst.ResourceID == instanceID {
			expectedInstance = inst
			break
		}
	}

	if expectedInstance == nil {
		return nil, nil // Instance not found in terraform file
	}

	// Detect drift
	driftResult, err := a.driftDetector.DetectDrift(actualInstance, expectedInstance, attributes)
	if err != nil {
		return nil, err
	}

	return driftResult, nil
}

// RunBatchInstanceCheck performs drift detection on multiple EC2 instances
func (a *Application) RunBatchInstanceCheck(ctx context.Context, instanceIDs []string, terraformPath string, attributes []string) (map[string]*interfaces.DriftResult, error) {
	a.wg.Add(1)
	defer a.wg.Done()

	// Get actual instance states from AWS
	actualInstances, err := a.awsClient.GetMultipleEC2Instances(ctx, instanceIDs)
	if err != nil {
		return nil, err
	}

	// Parse Terraform configuration to get expected state
	expectedInstances, err := a.terraformParser.ParseTerraformHCL(terraformPath)
	if err != nil {
		return nil, err
	}

	// Detect drift for all instances using batch detection
	driftResults, err := a.driftDetector.DetectMultipleDrift(actualInstances, expectedInstances, attributes)
	if err != nil {
		return nil, err
	}

	return driftResults, nil
}

// GenerateReport generates a report from drift results
func (a *Application) GenerateReport(driftResults map[string]*interfaces.DriftResult, format string) ([]byte, error) {
	switch format {
	case "json":
		return a.reportGenerator.GenerateJSONReport(driftResults)
	case "yaml":
		return a.reportGenerator.GenerateYAMLReport(driftResults)
	default:
		return a.reportGenerator.GenerateJSONReport(driftResults)
	}
}

// Context returns the application context
func (a *Application) Context() context.Context {
	return a.ctx
}

// Config returns the application configuration
func (a *Application) Config() *config.Config {
	return a.config
}