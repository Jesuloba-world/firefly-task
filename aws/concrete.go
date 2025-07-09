package aws

import (
	"context"
	"fmt"

	"firefly-task/pkg/interfaces"
	"github.com/sirupsen/logrus"
)

// ConcreteEC2Client implements the interfaces.EC2Client interface
type ConcreteEC2Client struct {
	client *Client
	logger *logrus.Logger
}

// ConcreteS3Client implements the interfaces.S3Client interface
type ConcreteS3Client struct {
	client *Client
	logger *logrus.Logger
}

// ConcreteAWSClientFactory implements the interfaces.AWSClientFactory interface
type ConcreteAWSClientFactory struct {
	config ClientConfig
	logger *logrus.Logger
}

// NewConcreteAWSClientFactory creates a new concrete AWS client factory
func NewConcreteAWSClientFactory(config ClientConfig) interfaces.AWSClientFactory {
	logger := config.Logger
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteAWSClientFactory{
		config: config,
		logger: logger,
	}
}

// CreateEC2Client creates a new EC2 client instance
func (f *ConcreteAWSClientFactory) CreateEC2Client(ctx context.Context) (interfaces.EC2Client, error) {
	client, err := NewClient(ctx, f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	return &ConcreteEC2Client{
		client: client,
		logger: f.logger,
	}, nil
}

// CreateS3Client creates a new S3 client instance
func (f *ConcreteAWSClientFactory) CreateS3Client(ctx context.Context) (interfaces.S3Client, error) {
	// For now, we'll use the same client structure
	// In the future, this could be extended to include S3-specific functionality
	client, err := NewClient(ctx, f.config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	return &ConcreteS3Client{
		client: client,
		logger: f.logger,
	}, nil
}

// EC2Client implementation methods

// GetEC2Instance retrieves a single EC2 instance by ID
func (c *ConcreteEC2Client) GetEC2Instance(ctx context.Context, instanceID string) (*interfaces.EC2Instance, error) {
	c.logger.Debugf("ConcreteEC2Client: Getting EC2 instance %s", instanceID)
	return c.client.GetEC2Instance(ctx, instanceID)
}

// ListEC2Instances retrieves all EC2 instances
func (c *ConcreteEC2Client) ListEC2Instances(ctx context.Context) ([]*interfaces.EC2Instance, error) {
	c.logger.Debug("ConcreteEC2Client: Listing all EC2 instances")
	// This method would need to be implemented in the underlying client
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("ListEC2Instances not yet implemented")
}

// GetEC2InstancesByTags retrieves EC2 instances filtered by tags
func (c *ConcreteEC2Client) GetEC2InstancesByTags(ctx context.Context, tags map[string]string) ([]*interfaces.EC2Instance, error) {
	c.logger.Debugf("ConcreteEC2Client: Getting EC2 instances by tags: %v", tags)
	// This method would need to be implemented in the underlying client
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("GetEC2InstancesByTags not yet implemented")
}

// GetMultipleEC2Instances retrieves multiple EC2 instances by IDs
func (c *ConcreteEC2Client) GetMultipleEC2Instances(ctx context.Context, instanceIDs []string) (map[string]*interfaces.EC2Instance, error) {
	c.logger.Debugf("ConcreteEC2Client: Getting multiple EC2 instances: %v", instanceIDs)
	return c.client.GetMultipleEC2Instances(ctx, instanceIDs)
}

// S3Client implementation methods (placeholder implementations)

// GetBucket retrieves information about a specific S3 bucket
func (c *ConcreteS3Client) GetBucket(ctx context.Context, bucketName string) (*interfaces.S3Bucket, error) {
	c.logger.Debugf("ConcreteS3Client: Getting S3 bucket %s", bucketName)
	// Placeholder implementation - S3 functionality not yet implemented
	return nil, fmt.Errorf("S3 functionality not yet implemented")
}

// ListBuckets retrieves all S3 buckets
func (c *ConcreteS3Client) ListBuckets(ctx context.Context) ([]*interfaces.S3Bucket, error) {
	c.logger.Debug("ConcreteS3Client: Listing all S3 buckets")
	// Placeholder implementation - S3 functionality not yet implemented
	return nil, fmt.Errorf("S3 functionality not yet implemented")
}

// GetBucketPolicy retrieves the policy for a specific S3 bucket
func (c *ConcreteS3Client) GetBucketPolicy(ctx context.Context, bucketName string) (string, error) {
	c.logger.Debugf("ConcreteS3Client: Getting S3 bucket policy for %s", bucketName)
	// Placeholder implementation - S3 functionality not yet implemented
	return "", fmt.Errorf("S3 functionality not yet implemented")
}