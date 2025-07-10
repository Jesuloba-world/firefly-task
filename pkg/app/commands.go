package app

import (
	"context"
	"fmt"
	"os"


	"github.com/spf13/cobra"
)

// CommandHandler handles all CLI commands for the application
type CommandHandler struct {
	app *Application
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(app *Application) *CommandHandler {
	return &CommandHandler{app: app}
}

// CreateRootCommand creates the root cobra command
func (h *CommandHandler) CreateRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "firefly-task",
		Short: "A tool for detecting drift in EC2 instances",
		Long: `Firefly Task is a CLI tool that helps detect configuration drift
between actual EC2 instances and their Terraform configurations.`,
	}

	// Add subcommands
	rootCmd.AddCommand(h.CreateCheckCommand())
	rootCmd.AddCommand(h.CreateBatchCommand())
	rootCmd.AddCommand(h.CreateAttributeCommand())

	return rootCmd
}

// CreateCheckCommand creates the check command for single instance drift detection
func (h *CommandHandler) CreateCheckCommand() *cobra.Command {
	var instanceID, terraformPath, outputFile string
	var attributes []string

	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Check drift for a single EC2 instance",
		Long:  `Check configuration drift for a single EC2 instance against its Terraform configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.handleCheckCommand(cmd.Context(), instanceID, terraformPath, outputFile, attributes)
		},
	}

	// Add flags
	checkCmd.Flags().StringVarP(&instanceID, "instance-id", "i", "", "EC2 instance ID to check (required)")
	checkCmd.Flags().StringVarP(&terraformPath, "tf-path", "t", "", "Path to Terraform configuration file (required)")
	checkCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (optional, prints to stdout if not specified)")
	checkCmd.Flags().StringSliceVarP(&attributes, "attributes", "a", DefaultAttributes, "Attributes to check for drift")

	// Mark required flags
	checkCmd.MarkFlagRequired("instance-id")
	checkCmd.MarkFlagRequired("tf-path")

	return checkCmd
}

// CreateBatchCommand creates the batch command for multiple instance drift detection
func (h *CommandHandler) CreateBatchCommand() *cobra.Command {
	var inputFile, terraformPath, outputFile string
	var attributes []string

	batchCmd := &cobra.Command{
		Use:   "batch",
		Short: "Check drift for multiple EC2 instances",
		Long:  `Check configuration drift for multiple EC2 instances listed in a file against their Terraform configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.handleBatchCommand(cmd.Context(), inputFile, terraformPath, outputFile, attributes)
		},
	}

	// Add flags
	batchCmd.Flags().StringVarP(&inputFile, "input-file", "f", "", "File containing list of instance IDs (required)")
	batchCmd.Flags().StringVarP(&terraformPath, "tf-path", "t", "", "Path to Terraform configuration file (required)")
	batchCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (optional, prints to stdout if not specified)")
	batchCmd.Flags().StringSliceVarP(&attributes, "attributes", "a", DefaultAttributes, "Attributes to check for drift")

	// Mark required flags
	batchCmd.MarkFlagRequired("input-file")
	batchCmd.MarkFlagRequired("tf-path")

	return batchCmd
}

// CreateAttributeCommand creates the attribute command for attribute-specific drift detection
func (h *CommandHandler) CreateAttributeCommand() *cobra.Command {
	var instanceID, terraformPath, attribute, outputFile string

	attributeCmd := &cobra.Command{
		Use:   "attribute",
		Short: "Check drift for a specific attribute of an EC2 instance",
		Long:  `Check configuration drift for a specific attribute of an EC2 instance against its Terraform configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.handleAttributeCommand(cmd.Context(), instanceID, terraformPath, attribute, outputFile)
		},
	}

	// Add flags
	attributeCmd.Flags().StringVarP(&instanceID, "instance-id", "i", "", "EC2 instance ID to check (required)")
	attributeCmd.Flags().StringVarP(&terraformPath, "tf-path", "t", "", "Path to Terraform configuration file (required)")
	attributeCmd.Flags().StringVarP(&attribute, "attribute", "a", "", "Specific attribute to check for drift (required)")
	attributeCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (optional, prints to stdout if not specified)")

	// Mark required flags
	attributeCmd.MarkFlagRequired("instance-id")
	attributeCmd.MarkFlagRequired("tf-path")
	attributeCmd.MarkFlagRequired("attribute")

	return attributeCmd
}

// handleCheckCommand handles the check command execution
func (h *CommandHandler) handleCheckCommand(ctx context.Context, instanceID, terraformPath, outputFile string, attributes []string) error {
	// Start the application
	if err := h.app.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer h.app.Shutdown()

	// Run single check
	reportData, err := h.app.RunSingleCheck(ctx, instanceID, terraformPath, attributes)
	if err != nil {
		return err
	}

	// Output result
	return h.outputResult(reportData, outputFile)
}

// handleBatchCommand handles the batch command execution
func (h *CommandHandler) handleBatchCommand(ctx context.Context, inputFile, terraformPath, outputFile string, attributes []string) error {
	// Start the application
	if err := h.app.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer h.app.Shutdown()

	// Run batch check
	reportData, err := h.app.RunBatchCheck(ctx, inputFile, terraformPath, attributes)
	if err != nil {
		return err
	}

	// Output result
	return h.outputResult(reportData, outputFile)
}

// handleAttributeCommand handles the attribute command execution
func (h *CommandHandler) handleAttributeCommand(ctx context.Context, instanceID, terraformPath, attribute, outputFile string) error {
	// Start the application
	if err := h.app.Start(); err != nil {
		return fmt.Errorf("failed to start application: %w", err)
	}
	defer h.app.Shutdown()

	// Run attribute check
	reportData, err := h.app.RunAttributeCheck(ctx, instanceID, terraformPath, attribute)
	if err != nil {
		return err
	}

	// Output result
	return h.outputResult(reportData, outputFile)
}

// outputResult outputs the result to file or stdout
func (h *CommandHandler) outputResult(data []byte, outputFile string) error {
	if outputFile != "" {
		// Write to file
		if err := os.WriteFile(outputFile, data, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		// Print to stdout
		fmt.Print(string(data))
	}
	return nil
}

// ExecuteCommand executes the root command with the provided arguments
func (h *CommandHandler) ExecuteCommand(args []string) error {
	rootCmd := h.CreateRootCommand()
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

// ExecuteRootCommand executes the root command (used by main.go)
func (h *CommandHandler) ExecuteRootCommand() error {
	rootCmd := h.CreateRootCommand()
	return rootCmd.Execute()
}