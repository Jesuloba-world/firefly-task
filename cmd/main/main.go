package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"firefly-task/config"
	"firefly-task/errors"
	"firefly-task/pkg/container"
	"firefly-task/pkg/interfaces"
)

var cfg config.Config
var diContainer *container.Container

// initLogger creates and configures a logger based on the global config
func initLogger() (*logrus.Logger, error) {
	logger := logrus.New()
	if cfg.Verbose {
		logger.SetLevel(logrus.DebugLevel)
	} else {
		logger.SetLevel(logrus.InfoLevel)
	}
	return logger, nil
}

// initContainer initializes the dependency injection container
func initContainer() error {
	logger, err := initLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	diContainer = container.NewContainerWithLogger(logger)
	err = diContainer.RegisterDefaults()
	if err != nil {
		return fmt.Errorf("failed to register default services: %w", err)
	}

	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "firefly",
	Short: "AWS EC2 configuration drift detection tool",
	Long: `Firefly is a CLI tool for detecting configuration drift in AWS EC2 instances.
It compares the current state of EC2 instances with their expected configuration
defined in Terraform files and reports any discrepancies.`,
	Version: "1.0.0",
}

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for configuration drift in a single EC2 instance",
	Long: `Check command analyzes a single EC2 instance for configuration drift.
It compares the current instance configuration with the expected state
defined in Terraform files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize dependency injection container
		if err := initContainer(); err != nil {
			return fmt.Errorf("failed to initialize container: %w", err)
		}

		logger, err := diContainer.GetLogger()
		if err != nil {
			return fmt.Errorf("failed to get logger: %w", err)
		}

		if cfg.InstanceID == "" {
			return errors.NewValidationError("instance-id is required")
		}
		if cfg.TerraformPath == "" {
			return errors.NewValidationError("tf-path is required")
		}

		logger.Infof("Starting drift check for instance: %s", cfg.InstanceID)

		// Get dependencies from container
		ec2Client, err := diContainer.GetEC2Client()
		if err != nil {
			return fmt.Errorf("failed to get EC2 client: %w", err)
		}

		tfParser, err := diContainer.GetTerraformParser()
		if err != nil {
			return fmt.Errorf("failed to get Terraform parser: %w", err)
		}

		driftDetector, err := diContainer.GetDriftDetector()
		if err != nil {
			return fmt.Errorf("failed to get drift detector: %w", err)
		}

		reportGenerator, err := diContainer.GetReportGenerator()
		if err != nil {
			return fmt.Errorf("failed to get report generator: %w", err)
		}

		// Get actual instance state from AWS
		actualInstance, err := ec2Client.GetEC2Instance(cmd.Context(), cfg.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get ec2 instance: %w", err)
		}

		// Parse Terraform configuration to get expected state
		expectedInstances, err := tfParser.ParseTerraformHCL(cfg.TerraformPath)
		if err != nil {
			return fmt.Errorf("failed to parse terraform file: %w", err)
		}

		// Find the expected instance configuration
		var expectedInstance *interfaces.TerraformConfig
		for _, inst := range expectedInstances {
			if inst.ResourceID == cfg.InstanceID {
				expectedInstance = inst
				break
			}
		}

		if expectedInstance == nil {
			return fmt.Errorf("instance %s not found in terraform file", cfg.InstanceID)
		}

		// Detect drift
		driftResult, err := driftDetector.DetectDrift(actualInstance, expectedInstance, []string{"instance_type", "state", "subnet_id", "vpc_id", "security_groups", "tags"})
		if err != nil {
			return fmt.Errorf("failed to detect drift: %w", err)
		}

		// Generate and print report
		reportData, err := reportGenerator.GenerateJSONReport(map[string]*interfaces.DriftResult{cfg.InstanceID: driftResult})
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		fmt.Println(string(reportData))

		return nil
	},
}

// batchCmd represents the batch command
var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Check for configuration drift in multiple EC2 instances",
	Long: `Batch command analyzes multiple EC2 instances for configuration drift.
It reads a list of instance IDs from an input file and processes them
concurrently for improved performance.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize dependency injection container
		if err := initContainer(); err != nil {
			return fmt.Errorf("failed to initialize container: %w", err)
		}

		logger, err := diContainer.GetLogger()
		if err != nil {
			return fmt.Errorf("failed to get logger: %w", err)
		}

		if cfg.InputFile == "" {
			return errors.NewValidationError("input-file is required")
		}

		logger.Infof("Starting batch drift check from file: %s", cfg.InputFile)

		// Read instance IDs from input file
		instanceIDs, err := config.ReadInstanceIDs(cfg.InputFile)
		if err != nil {
			return fmt.Errorf("failed to read instance ids from file: %w", err)
		}

		// Get dependencies from container
		ec2Client, err := diContainer.GetEC2Client()
		if err != nil {
			return fmt.Errorf("failed to get EC2 client: %w", err)
		}

		tfParser, err := diContainer.GetTerraformParser()
		if err != nil {
			return fmt.Errorf("failed to get Terraform parser: %w", err)
		}

		driftDetector, err := diContainer.GetDriftDetector()
		if err != nil {
			return fmt.Errorf("failed to get drift detector: %w", err)
		}

		reportGenerator, err := diContainer.GetReportGenerator()
		if err != nil {
			return fmt.Errorf("failed to get report generator: %w", err)
		}

		// Get actual instance states from AWS
		actualInstances, err := ec2Client.GetMultipleEC2Instances(cmd.Context(), instanceIDs)
		if err != nil {
			return fmt.Errorf("failed to get ec2 instances: %w", err)
		}

		// Parse Terraform configuration to get expected state
		expectedInstances, err := tfParser.ParseTerraformHCL(cfg.TerraformPath)
		if err != nil {
			return fmt.Errorf("failed to parse terraform file: %w", err)
		}

		// Detect drift for all instances using batch detection
		driftResults, err := driftDetector.DetectMultipleDrift(actualInstances, expectedInstances, []string{"instance_type", "state", "subnet_id", "vpc_id", "security_groups", "tags"})
		if err != nil {
			return fmt.Errorf("failed to detect drift: %w", err)
		}

		// Generate and print report
		reportData, err := reportGenerator.GenerateJSONReport(driftResults)
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		fmt.Println(string(reportData))

		return nil
	},
}

// attributeCmd represents the attribute command
var attributeCmd = &cobra.Command{
	Use:   "attribute",
	Short: "Check for drift in a specific attribute",
	Long: `Attribute command focuses on checking drift for a specific EC2 attribute.
This is useful when you want to monitor changes in particular configuration
aspects like security groups, instance type, or tags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize dependency injection container
		if err := initContainer(); err != nil {
			return fmt.Errorf("failed to initialize container: %w", err)
		}

		logger, err := diContainer.GetLogger()
		if err != nil {
			return fmt.Errorf("failed to get logger: %w", err)
		}

		if cfg.InstanceID == "" {
			return errors.NewValidationError("instance-id is required")
		}
		if cfg.TerraformPath == "" {
			return errors.NewValidationError("tf-path is required")
		}
		if cfg.Attribute == "" {
			return errors.NewValidationError("attribute is required")
		}

		logger.Infof("Checking drift for instance %s attribute %s", cfg.InstanceID, cfg.Attribute)

		// Get dependencies from container
		ec2Client, err := diContainer.GetEC2Client()
		if err != nil {
			return fmt.Errorf("failed to get EC2 client: %w", err)
		}

		tfParser, err := diContainer.GetTerraformParser()
		if err != nil {
			return fmt.Errorf("failed to get Terraform parser: %w", err)
		}

		driftDetector, err := diContainer.GetDriftDetector()
		if err != nil {
			return fmt.Errorf("failed to get drift detector: %w", err)
		}

		reportGenerator, err := diContainer.GetReportGenerator()
		if err != nil {
			return fmt.Errorf("failed to get report generator: %w", err)
		}

		// Get actual instance state from AWS
		actualInstance, err := ec2Client.GetEC2Instance(cmd.Context(), cfg.InstanceID)
		if err != nil {
			return fmt.Errorf("failed to get ec2 instance: %w", err)
		}

		// Parse Terraform configuration to get expected state
		expectedInstances, err := tfParser.ParseTerraformHCL(cfg.TerraformPath)
		if err != nil {
			return fmt.Errorf("failed to parse terraform file: %w", err)
		}

		// Find the expected instance configuration
		var expectedInstance *interfaces.TerraformConfig
		for _, inst := range expectedInstances {
			if inst.ResourceID == cfg.InstanceID {
				expectedInstance = inst
				break
			}
		}

		if expectedInstance == nil {
			return fmt.Errorf("instance %s not found in terraform file", cfg.InstanceID)
		}

		// Detect drift for specific attribute
		driftResult, err := driftDetector.DetectDrift(actualInstance, expectedInstance, []string{cfg.Attribute})
		if err != nil {
			return fmt.Errorf("failed to detect drift: %w", err)
		}

		// Generate and print report
		reportData, err := reportGenerator.GenerateJSONReport(map[string]*interfaces.DriftResult{cfg.InstanceID: driftResult})
		if err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		fmt.Println(string(reportData))

		return nil
	},
}

func init() {
	// Add global flags
	rootCmd.PersistentFlags().StringVar(&cfg.AWSProfile, "aws-profile", "", "AWS profile to use for authentication")
	rootCmd.PersistentFlags().StringVar(&cfg.AWSRegion, "aws-region", "", "AWS region to operate in")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&cfg.Output, "output", "text", "Output format (json|text|table)")

	// Add command-specific flags for check command
	checkCmd.Flags().StringVar(&cfg.TerraformPath, "tf-path", "", "Path to Terraform configuration file")
	checkCmd.Flags().StringVar(&cfg.InstanceID, "instance-id", "", "EC2 instance ID to check")

	// Add command-specific flags for batch command
	batchCmd.Flags().StringVar(&cfg.InputFile, "input-file", "", "Input file containing instances to process")
	batchCmd.Flags().IntVar(&cfg.Concurrency, "concurrency", 5, "Number of concurrent operations (1-50)")

	// Add command-specific flags for attribute command
	attributeCmd.Flags().StringVar(&cfg.TerraformPath, "tf-path", "", "Path to Terraform configuration file")
	attributeCmd.Flags().StringVar(&cfg.InstanceID, "instance-id", "", "EC2 instance ID to check")
	attributeCmd.Flags().StringVar(&cfg.Attribute, "attribute", "", "Specific attribute to check for drift")

	// Add subcommands
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(batchCmd)
	rootCmd.AddCommand(attributeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
