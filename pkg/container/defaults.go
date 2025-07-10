package container

import (
	"context"

	"github.com/sirupsen/logrus"

	"firefly-task/drift"
	"firefly-task/pkg/interfaces"
	"firefly-task/report"
	"firefly-task/terraform"
)

// StubEC2Client is a simple stub implementation of the interfaces.EC2Client for testing
type StubEC2Client struct{}

func (s *StubEC2Client) GetEC2Instance(ctx context.Context, instanceID string) (*interfaces.EC2Instance, error) {
	return &interfaces.EC2Instance{
		InstanceID:   instanceID,
		InstanceType: "t2.micro",
		State:        "running",
		Tags:         make(map[string]string),
	}, nil
}

func (s *StubEC2Client) GetMultipleEC2Instances(ctx context.Context, instanceIDs []string) (map[string]*interfaces.EC2Instance, error) {
	result := make(map[string]*interfaces.EC2Instance)
	for _, id := range instanceIDs {
		result[id] = &interfaces.EC2Instance{
			InstanceID:   id,
			InstanceType: "t2.micro",
			State:        "running",
			Tags:         make(map[string]string),
		}
	}
	return result, nil
}

func (s *StubEC2Client) GetEC2InstancesByTags(ctx context.Context, tags map[string]string) ([]*interfaces.EC2Instance, error) {
	return []*interfaces.EC2Instance{}, nil
}

func (s *StubEC2Client) ListEC2Instances(ctx context.Context) ([]*interfaces.EC2Instance, error) {
	return []*interfaces.EC2Instance{}, nil
}

// RegisterDefaults registers the default services in the container
func (c *Container) RegisterDefaults() error {
	c.RegisterFactory("logger", func() interface{} {
		logger := logrus.New()
		logger.SetLevel(logrus.InfoLevel)
		return logger
	})

	c.RegisterFactory("ec2Client", func() interface{} {
		// Return a stub EC2Client for testing/development
		return &StubEC2Client{}
	})

	c.RegisterFactory("terraformParser", func() interface{} {
		return terraform.NewParser()
	})

	c.RegisterFactory("driftDetector", func() interface{} {
		logger, _ := c.GetLogger()
		return drift.NewConcreteDriftDetector(logger)
	})

	c.RegisterFactory("reportGenerator", func() interface{} {
		loggerInterface, err := c.Get("logger")
		if err != nil {
			return nil
		}
		logger, ok := loggerInterface.(*logrus.Logger)
		if !ok {
			return nil
		}
		return report.NewConcreteReportGenerator(logger)
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
