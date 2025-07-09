package aws

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"firefly-task/pkg/interfaces"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

var (
	// ErrInstanceNotFound is returned when the requested EC2 instance does not exist
	ErrInstanceNotFound = errors.New("ec2 instance not found")

	// ErrInvalidInstanceID is returned when the provided instance ID is invalid
	ErrInvalidInstanceID = errors.New("invalid instance id")
)

// GetEC2Instance retrieves a single EC2 instance by its ID with retry logic
func (c *Client) GetEC2Instance(ctx context.Context, instanceID string) (*interfaces.EC2Instance, error) {
	if instanceID == "" {
		return nil, ErrInvalidInstanceID
	}

	c.logger.Debugf("Retrieving EC2 instance with ID: %s", instanceID)

	// Create the input for DescribeInstances API
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}

	// Call the AWS API with retry logic
	var resp *ec2.DescribeInstancesOutput
	err := c.retryWithBackoff(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.ec2.DescribeInstances(ctx, input)
		return err
	})

	if err != nil {
		// Check for specific AWS errors
		if strings.Contains(err.Error(), ErrorPatternInstanceNotFound) {
			return nil, ErrInstanceNotFound
		}

		return nil, fmt.Errorf("failed to describe EC2 instance: %w", err)
	}

	// Check if we got any reservations and instances
	if len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
		return nil, ErrInstanceNotFound
	}

	// Convert the AWS instance to our model
	instance := convertFromAWSInstance(resp.Reservations[0].Instances[0])

	c.logger.Debugf("Successfully retrieved EC2 instance: %s", instanceID)
	return instance, nil
}

// retryWithBackoff implements exponential backoff retry logic
func (c *Client) retryWithBackoff(ctx context.Context, operation func(ctx context.Context) error) error {
	const (
		maxRetries     = 3
		baseDelay      = 100 * time.Millisecond
		maxDelay       = 5 * time.Second
		requestTimeout = 10 * time.Second // Add a request timeout
	)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Create a new context with a timeout for each attempt
		attemptCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		// Check if the parent context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the operation
		err := operation(attemptCtx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on the last attempt
		if attempt == maxRetries {
			break
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return err
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		if delay > maxDelay {
			delay = maxDelay
		}

		c.logger.Printf("Retrying operation after %v (attempt %d/%d): %v", delay, attempt+1, maxRetries, err)

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// isRetryableError determines if an error is retryable
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check for context deadline exceeded, which is retryable
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	errorStr := err.Error()

	// Don't retry client errors like InvalidInstanceID.NotFound
	if strings.Contains(errorStr, ErrorPatternInvalidInstanceID) ||
		strings.Contains(errorStr, ErrorPatternInvalidParameter) ||
		strings.Contains(errorStr, ErrorPatternUnauthorizedOperation) ||
		strings.Contains(errorStr, ErrorPatternInstanceNotFound) {
		return false
	}

	// For most other errors, especially network-related ones, retrying is safe.
	// This includes throttling errors, service unavailable, internal errors,
	// and common network issues.
	return true
}

// GetMultipleEC2Instances retrieves multiple EC2 instances by their IDs with retry logic
func (c *Client) GetMultipleEC2Instances(ctx context.Context, instanceIDs []string) (map[string]*interfaces.EC2Instance, error) {
	if len(instanceIDs) == 0 {
		return make(map[string]*interfaces.EC2Instance), nil
	}

	c.logger.Debugf("Retrieving %d EC2 instances", len(instanceIDs))

	// Create the input for DescribeInstances API
	input := &ec2.DescribeInstancesInput{
		InstanceIds: instanceIDs,
	}

	// Call the AWS API with retry logic
	var resp *ec2.DescribeInstancesOutput
	err := c.retryWithBackoff(ctx, func(ctx context.Context) error {
		var err error
		resp, err = c.ec2.DescribeInstances(ctx, input)
		return err
	})

	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}

	// Process the results
	result := make(map[string]*interfaces.EC2Instance)
	for _, reservation := range resp.Reservations {
		for _, awsInstance := range reservation.Instances {
			instance := convertFromAWSInstance(awsInstance)
			result[instance.InstanceID] = instance
		}
	}

	c.logger.Debugf("Successfully retrieved %d EC2 instances", len(result))
	return result, nil
}

// convertFromAWSInstance converts AWS SDK EC2 Instance to our EC2Instance model
func convertFromAWSInstance(awsInstance types.Instance) *interfaces.EC2Instance {
	instance := &interfaces.EC2Instance{
		InstanceID:       *awsInstance.InstanceId,
		InstanceType:     string(awsInstance.InstanceType),
		State:            string(awsInstance.State.Name),
		PublicIPAddress:  awsInstance.PublicIpAddress,
		PrivateIPAddress: awsInstance.PrivateIpAddress,
		PublicDNSName:    awsInstance.PublicDnsName,
		PrivateDNSName:   awsInstance.PrivateDnsName,
		SubnetID:         awsInstance.SubnetId,
		VPCID:            awsInstance.VpcId,
		LaunchTime:       awsInstance.LaunchTime,
		ImageID:          awsInstance.ImageId,
		KeyName:          awsInstance.KeyName,
		Tags:             make(map[string]string),
	}

	// Handle placement (availability zone)
	if awsInstance.Placement != nil {
		instance.AvailabilityZone = awsInstance.Placement.AvailabilityZone
	}

	// Handle platform conversion
	if awsInstance.Platform != "" {
		platformStr := string(awsInstance.Platform)
		instance.Platform = &platformStr
	}

	// Handle architecture conversion
	if awsInstance.Architecture != "" {
		archStr := string(awsInstance.Architecture)
		instance.Architecture = &archStr
	}

	// Handle EBS optimized (it's a pointer in AWS SDK)
	if awsInstance.EbsOptimized != nil {
		instance.EBSOptimized = *awsInstance.EbsOptimized
	}

	// Convert monitoring state
	if awsInstance.Monitoring != nil {
		instance.Monitoring = awsInstance.Monitoring.State == types.MonitoringStateEnabled
	}

	// Convert security groups
	for _, sg := range awsInstance.SecurityGroups {
		instance.SecurityGroups = append(instance.SecurityGroups, interfaces.SecurityGroup{
			GroupID:   *sg.GroupId,
			GroupName: *sg.GroupName,
		})
	}

	// Convert tags
	for _, tag := range awsInstance.Tags {
		if tag.Key != nil && tag.Value != nil {
			instance.Tags[*tag.Key] = *tag.Value
		}
	}

	return instance
}
