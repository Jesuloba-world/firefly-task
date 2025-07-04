package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"firefly-task/config"
	"firefly-task/errors"
	"firefly-task/logging"
)

var cfg config.Config

// initLogger creates and configures a logger based on the global config
func initLogger() (*logging.Logger, error) {
	logConfig := logging.DefaultLogConfig()
	if cfg.Verbose {
		logConfig.Level = "debug"
	}
	return logging.NewLogger(logConfig)
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
		logger, err := initLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		if cfg.InstanceID == "" {
			return errors.NewValidationError("instance-id is required")
		}
		if cfg.TerraformPath == "" {
			return errors.NewValidationError("tf-path is required")
		}

		logger.Infof("Starting drift check for instance: %s", cfg.InstanceID)
		// TODO: Implement drift detection logic
		return fmt.Errorf("drift detection functionality not yet implemented")
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
		logger, err := initLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		if cfg.InputFile == "" {
			return errors.NewValidationError("input-file is required")
		}

		logger.Infof("Starting batch drift check from file: %s", cfg.InputFile)
		// TODO: Implement batch drift detection logic
		return fmt.Errorf("batch drift detection functionality not yet implemented")
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
		logger, err := initLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		if cfg.Attribute == "" {
			return errors.NewValidationError("attribute is required")
		}

		logger.Infof("Starting attribute drift check for: %s", cfg.Attribute)
		// TODO: Implement attribute drift detection logic
		return fmt.Errorf("attribute drift detection functionality not yet implemented")
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
