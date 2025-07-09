package container

import (
	"github.com/sirupsen/logrus"

	"firefly-task/drift"
	"firefly-task/logging"
	"firefly-task/pkg/interfaces"
	"firefly-task/report"
	"firefly-task/terraform"
)

// RegisterDefaults registers the default services in the container
func (c *Container) RegisterDefaults() error {
	c.RegisterFactory("logger", func() interface{} {
		logger, _ := logging.NewLogger(logging.DefaultLogConfig())
		return logger
	})

	c.RegisterFactory("ec2Client", func() interface{} {
		// Return a placeholder or nil since NewEC2Client is undefined
		return nil
	})

	c.RegisterFactory("terraformParser", func() interface{} {
		return terraform.NewParser()
	})

	c.RegisterFactory("driftDetector", func() interface{} {
		return drift.NewDriftDetector(drift.DefaultDetectionConfig())
	})

	c.RegisterFactory("reportGenerator", func() interface{} {
		generator, _ := report.NewReportGenerator(report.StandardGenerator)
		return generator
	})

	return nil
}

// GetLogger retrieves the logger from the container
func (c *Container) GetLogger() (*logrus.Logger, error) {
	return GetTyped[*logrus.Logger](c, "logger")
}

// GetEC2Client retrieves the EC2 client from the container
func (c *Container) GetEC2Client() (interfaces.EC2Client, error) {
	return GetTyped[interfaces.EC2Client](c, "ec2Client")
}

// GetTerraformParser retrieves the Terraform parser from the container
func (c *Container) GetTerraformParser() (interfaces.TerraformParser, error) {
	return GetTyped[interfaces.TerraformParser](c, "terraformParser")
}

// GetDriftDetector retrieves the drift detector from the container
func (c *Container) GetDriftDetector() (interfaces.DriftDetector, error) {
	return GetTyped[interfaces.DriftDetector](c, "driftDetector")
}

// GetReportGenerator retrieves the report generator from the container
func (c *Container) GetReportGenerator() (interfaces.ReportGenerator, error) {
	return GetTyped[interfaces.ReportGenerator](c, "reportGenerator")
}
