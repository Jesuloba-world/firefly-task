package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/sirupsen/logrus"
)

// ClientConfig holds AWS client configuration options
type ClientConfig struct {
	// Region is the AWS region to connect to
	Region string

	// Profile is the AWS profile to use from the credentials file
	Profile string

	// AccessKeyID is the AWS access key ID for direct credential configuration
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key for direct credential configuration
	SecretAccessKey string

	// Logger is the logger to use for AWS client operations
	Logger *logrus.Logger
}

// Client represents an AWS client with EC2 service access
type Client struct {
	config *aws.Config
	ec2    *ec2.Client
	logger *logrus.Logger
}

// NewClient creates a new AWS client with the provided configuration
func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	var awsConfig aws.Config
	var err error

	// Set up logger
	logger := cfg.Logger
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	// Configure AWS SDK options
	options := []func(*config.LoadOptions) error{}

	// Set region if provided
	if cfg.Region != "" {
		options = append(options, config.WithRegion(cfg.Region))
	}

	// Set profile if provided
	if cfg.Profile != "" {
		options = append(options, config.WithSharedConfigProfile(cfg.Profile))
	}

	// Set static credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		options = append(options, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	// Load AWS configuration with credential chain
	// Order: Static credentials -> Environment variables -> Config files -> IAM roles
	awsConfig, err = config.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	// Create EC2 client
	ec2Client := ec2.NewFromConfig(awsConfig)

	return &Client{
		config: &awsConfig,
		ec2:    ec2Client,
		logger: logger,
	}, nil
}

// EC2 returns the EC2 service client
func (c *Client) EC2() *ec2.Client {
	return c.ec2
}
