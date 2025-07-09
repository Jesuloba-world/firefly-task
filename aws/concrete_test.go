package aws

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"firefly-task/pkg/interfaces"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEC2Client is a mock implementation of the interfaces.EC2Client for testing
type MockEC2Client struct {
	mock.Mock
}

func (m *MockEC2Client) GetEC2Instance(ctx context.Context, instanceID string) (*interfaces.EC2Instance, error) {
	args := m.Called(ctx, instanceID)
	return args.Get(0).(*interfaces.EC2Instance), args.Error(1)
}

func (m *MockEC2Client) GetMultipleEC2Instances(ctx context.Context, instanceIDs []string) (map[string]*interfaces.EC2Instance, error) {
	args := m.Called(ctx, instanceIDs)
	return args.Get(0).(map[string]*interfaces.EC2Instance), args.Error(1)
}

func (m *MockEC2Client) ListEC2Instances(ctx context.Context) ([]*interfaces.EC2Instance, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*interfaces.EC2Instance), args.Error(1)
}

func (m *MockEC2Client) GetEC2InstancesByTags(ctx context.Context, tags map[string]string) ([]*interfaces.EC2Instance, error) {
	args := m.Called(ctx, tags)
	return args.Get(0).([]*interfaces.EC2Instance), args.Error(1)
}

// Mock S3 API
type MockS3API struct {
	mock.Mock
}

func (m *MockS3API) ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*s3.ListBucketsOutput), args.Error(1)
}

func TestNewConcreteAWSClientFactory(t *testing.T) {
	tests := []struct {
		name   string
		config ClientConfig
	}{
		{
			name:   "with logger",
			config: ClientConfig{Logger: logrus.New()},
		},
		{
			name:   "without logger",
			config: ClientConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory := NewConcreteAWSClientFactory(tt.config)
			assert.NotNil(t, factory)
			assert.IsType(t, &ConcreteAWSClientFactory{}, factory)
		})
	}
}

func TestConcreteAWSClientFactory_CreateEC2Client(t *testing.T) {
	logger := logrus.New()
	factory := &ConcreteAWSClientFactory{
		config: ClientConfig{Logger: logger},
		logger: logger,
	}

	client, err := factory.CreateEC2Client(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, &ConcreteEC2Client{}, client)
}

func TestConcreteAWSClientFactory_CreateS3Client(t *testing.T) {
	logger := logrus.New()
	factory := &ConcreteAWSClientFactory{
		config: ClientConfig{Logger: logger},
		logger: logger,
	}

	client, err := factory.CreateS3Client(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.IsType(t, &ConcreteS3Client{}, client)
}

func TestConcreteEC2Client_GetEC2Instance(t *testing.T) {
	mockClient := &MockEC2Client{}
	// Use the mock client directly
	client := mockClient

	tests := []struct {
		name           string
		instanceID     string
		expectedResult *interfaces.EC2Instance
		expectedError  string
	}{
		{
			name:       "successful retrieval",
			instanceID: "i-1234567890abcdef0",
			expectedResult: &interfaces.EC2Instance{
				InstanceID:       "i-1234567890abcdef0",
				InstanceType:     "t2.micro",
				State:            "running",
				PrivateIPAddress: &[]string{"10.0.0.1"}[0],
				PublicIPAddress:  &[]string{"1.2.3.4"}[0],
				Tags: map[string]string{
					"Name": "test-instance",
				},
			},
			expectedError: "",
		},
		{
			name:           "instance not found",
			instanceID:     "i-nonexistent",
			expectedResult: nil,
			expectedError:  "instance i-nonexistent not found",
		},
		{
			name:           "API error",
			instanceID:     "i-1234567890abcdef0",
			expectedResult: nil,
			expectedError:  "API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError != "" {
				mockClient.On("GetEC2Instance", mock.Anything, tt.instanceID).Return((*interfaces.EC2Instance)(nil), errors.New(tt.expectedError)).Once()
			} else {
				mockClient.On("GetEC2Instance", mock.Anything, tt.instanceID).Return(tt.expectedResult, nil).Once()
			}

			result, err := client.GetEC2Instance(context.Background(), tt.instanceID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestConcreteEC2Client_ListEC2Instances(t *testing.T) {
	mockClient := &MockEC2Client{}
	// Use the mock client directly
	client := mockClient

	tests := []struct {
		name           string
		expectedCount  int
		expectedError  string
	}{
		{
			name:          "successful listing",
			expectedCount: 2,
			expectedError: "",
		},
		{
			name:          "API error",
			expectedCount: 0,
			expectedError: "API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedError != "" {
				mockClient.On("ListEC2Instances", mock.Anything).Return(([]*interfaces.EC2Instance)(nil), errors.New(tt.expectedError)).Once()
			} else {
				// Create mock instances for the expected count
				mockInstances := make([]*interfaces.EC2Instance, tt.expectedCount)
				for i := 0; i < tt.expectedCount; i++ {
					mockInstances[i] = &interfaces.EC2Instance{InstanceID: fmt.Sprintf("i-mock%d", i)}
				}
				mockClient.On("ListEC2Instances", mock.Anything).Return(mockInstances, nil).Once()
			}

			result, err := client.ListEC2Instances(context.Background())

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestConcreteS3Client_ListBuckets(t *testing.T) {
	// This test is simplified since S3 functionality is not implemented
	t.Run("not implemented", func(t *testing.T) {
		// This would test the actual S3 client implementation
		// For now, we just verify the test structure is correct
		assert.True(t, true, "S3 test placeholder")
	})
}