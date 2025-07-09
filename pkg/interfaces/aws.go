package interfaces

import (
	"context"
)



// EC2Client defines the interface for AWS EC2 operations
type EC2Client interface {
	// GetEC2Instance retrieves a single EC2 instance by its ID
	GetEC2Instance(ctx context.Context, instanceID string) (*EC2Instance, error)

	// GetMultipleEC2Instances retrieves multiple EC2 instances by their IDs
	GetMultipleEC2Instances(ctx context.Context, instanceIDs []string) (map[string]*EC2Instance, error)

	// ListEC2Instances retrieves all EC2 instances in the configured region
	ListEC2Instances(ctx context.Context) ([]*EC2Instance, error)

	// GetEC2InstancesByTags retrieves EC2 instances filtered by tags
	GetEC2InstancesByTags(ctx context.Context, tags map[string]string) ([]*EC2Instance, error)
}

// S3Client defines the interface for AWS S3 operations
type S3Client interface {
	// GetBucket retrieves information about an S3 bucket
	GetBucket(ctx context.Context, bucketName string) (*S3Bucket, error)

	// ListBuckets retrieves all S3 buckets
	ListBuckets(ctx context.Context) ([]*S3Bucket, error)

	// GetBucketPolicy retrieves the policy for an S3 bucket
	GetBucketPolicy(ctx context.Context, bucketName string) (string, error)
}

// AWSClientFactory defines the interface for creating AWS service clients
type AWSClientFactory interface {
	// CreateEC2Client creates a new EC2 client
	CreateEC2Client(ctx context.Context) (EC2Client, error)

	// CreateS3Client creates a new S3 client
	CreateS3Client(ctx context.Context) (S3Client, error)
}