package app

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"firefly-task/config"
	"firefly-task/pkg/interfaces"
)

// Mock implementations for testing
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

type MockTerraformParser struct {
	mock.Mock
}

func (m *MockTerraformParser) ParseTerraformState(statePath string) (map[string]*interfaces.TerraformConfig, error) {
	args := m.Called(statePath)
	return args.Get(0).(map[string]*interfaces.TerraformConfig), args.Error(1)
}

func (m *MockTerraformParser) ParseTerraformHCL(hclPath string) (map[string]*interfaces.TerraformConfig, error) {
	args := m.Called(hclPath)
	return args.Get(0).(map[string]*interfaces.TerraformConfig), args.Error(1)
}

func (m *MockTerraformParser) ValidateStateFile(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *MockTerraformParser) ValidateHCLDirectory(dirPath string) error {
	args := m.Called(dirPath)
	return args.Error(0)
}

func (m *MockTerraformParser) ValidateConfiguration(config *interfaces.TerraformConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

type MockDriftDetector struct {
	mock.Mock
}

func (m *MockDriftDetector) DetectDrift(actual *interfaces.EC2Instance, expected *interfaces.TerraformConfig, attributes []string) (*interfaces.DriftResult, error) {
	args := m.Called(actual, expected, attributes)
	return args.Get(0).(*interfaces.DriftResult), args.Error(1)
}

func (m *MockDriftDetector) DetectMultipleDrift(actualInstances map[string]*interfaces.EC2Instance, expectedConfigs map[string]*interfaces.TerraformConfig, attributes []string) (map[string]*interfaces.DriftResult, error) {
	args := m.Called(actualInstances, expectedConfigs, attributes)
	return args.Get(0).(map[string]*interfaces.DriftResult), args.Error(1)
}

func (m *MockDriftDetector) ValidateConfiguration(config *interfaces.TerraformConfig) error {
	args := m.Called(config)
	return args.Error(0)
}

type MockReportGenerator struct {
	mock.Mock
}

func (m *MockReportGenerator) GenerateJSONReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	args := m.Called(results)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockReportGenerator) GenerateYAMLReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	args := m.Called(results)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockReportGenerator) GenerateTableReport(results map[string]*interfaces.DriftResult) (string, error) {
	args := m.Called(results)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockReportGenerator) GenerateHTMLReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	args := m.Called(results)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockReportGenerator) GenerateMarkdownReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	args := m.Called(results)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockReportGenerator) WriteReport(results map[string]*interfaces.DriftResult, format string, writer io.Writer) error {
	args := m.Called(results, format, writer)
	return args.Error(0)
}

func (m *MockReportGenerator) WriteToFile(data []byte, filename string) error {
	args := m.Called(data, filename)
	return args.Error(0)
}

func (m *MockReportGenerator) WriteToConsole(data []byte) error {
	args := m.Called(data)
	return args.Error(0)
}

func TestNew(t *testing.T) {
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}
	cfg := &config.Config{}
	cfg.SetDefaults()

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	assert.NotNil(t, app)
	assert.NotNil(t, app.config)
	assert.NotNil(t, app.ctx)
}

func TestApplication_RunSingleInstanceCheck(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	ctx := context.Background()
	instanceID := "i-1234567890abcdef0"
	terraformPath := "/path/to/terraform"
	attributes := []string{"instance_type", "state"}

	// Mock EC2 instance
	ec2Instance := &interfaces.EC2Instance{
		InstanceID:   instanceID,
		InstanceType: "t3.micro",
		State:        "running",
	}

	// Mock Terraform config
	tfConfig := &interfaces.TerraformConfig{
		ResourceID:   instanceID,
		ResourceType: "aws_instance",
		ResourceName: "test",
		Attributes: map[string]interface{}{
			"instance_type": "t3.small",
			"state":         "running",
		},
	}

	// Mock drift result
	driftResult := &interfaces.DriftResult{
		ResourceID:     instanceID,
		ResourceType:   "aws_instance",
		IsDrifted:      true,
		DriftDetails: []*interfaces.DriftDetail{
			{
				Attribute:     "instance_type",
				ActualValue:   "t3.micro",
				ExpectedValue: "t3.small",
				Severity:      interfaces.SeverityMedium,
			},
		},
		Severity:       interfaces.SeverityMedium,
		DetectionTime:  time.Now(),
	}

	// Set up mocks
	mockEC2.On("GetEC2Instance", ctx, instanceID).Return(ec2Instance, nil)
	mockTF.On("ParseTerraformHCL", terraformPath).Return(map[string]*interfaces.TerraformConfig{instanceID: tfConfig}, nil)
	mockDrift.On("DetectDrift", ec2Instance, tfConfig, attributes).Return(driftResult, nil)

	// Execute
	result, err := app.RunSingleInstanceCheck(ctx, instanceID, terraformPath, attributes)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, driftResult, result)
	mockEC2.AssertExpectations(t)
	mockTF.AssertExpectations(t)
	mockDrift.AssertExpectations(t)
}

func TestApplication_RunBatchInstanceCheck(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	ctx := context.Background()
	instanceIDs := []string{"i-1234567890abcdef0", "i-0987654321fedcba0"}
	terraformPath := "/path/to/terraform"
	attributes := []string{"instance_type", "state"}


	// Mock EC2 instances
	ec2Instances := map[string]*interfaces.EC2Instance{
		"i-1234567890abcdef0": {
			InstanceID:   "i-1234567890abcdef0",
			InstanceType: "t3.micro",
			State:        "running",
		},
		"i-0987654321fedcba0": {
			InstanceID:   "i-0987654321fedcba0",
			InstanceType: "t3.small",
			State:        "stopped",
		},
	}

	// Mock Terraform configs
	tfConfigs := []*interfaces.TerraformConfig{
		{
			ResourceID:   "i-1234567890abcdef0",
			ResourceType: "aws_instance",
			ResourceName: "test1",
			Attributes: map[string]interface{}{
				"instance_type": "t3.micro",
				"state":         "running",
			},
		},
		{
			ResourceID:   "i-0987654321fedcba0",
			ResourceType: "aws_instance",
			ResourceName: "test2",
			Attributes: map[string]interface{}{
				"instance_type": "t3.small",
				"state":         "running",
			},
		},
	}

	// Mock drift results
	driftResults := map[string]*interfaces.DriftResult{
		"i-1234567890abcdef0": {
			ResourceID:     "i-1234567890abcdef0",
			ResourceType:   "aws_instance",
			IsDrifted:      false,
			DriftDetails:   []*interfaces.DriftDetail{},
			Severity:       interfaces.SeverityNone,
			DetectionTime:  time.Now(),
		},
		"i-0987654321fedcba0": {
			ResourceID:     "i-0987654321fedcba0",
			ResourceType:   "aws_instance",
			IsDrifted:      true,
			DriftDetails: []*interfaces.DriftDetail{
				{
					Attribute:     "state",
					ActualValue:   "stopped",
					ExpectedValue: "running",
					Severity:      interfaces.SeverityHigh,
				},
			},
			Severity:       interfaces.SeverityHigh,
			DetectionTime:  time.Now(),
		},
	}

	// Set up mocks
	mockEC2.On("GetMultipleEC2Instances", ctx, instanceIDs).Return(ec2Instances, nil)
	tfConfigsMap := map[string]*interfaces.TerraformConfig{
		"i-1234567890abcdef0": tfConfigs[0],
		"i-0987654321fedcba0": tfConfigs[1],
	}
	mockTF.On("ParseTerraformHCL", terraformPath).Return(tfConfigsMap, nil)
	mockDrift.On("DetectMultipleDrift", ec2Instances, tfConfigsMap, attributes).Return(driftResults, nil)

	// Execute
	results, err := app.RunBatchInstanceCheck(ctx, instanceIDs, terraformPath, attributes)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, driftResults, results)
	mockEC2.AssertExpectations(t)
	mockTF.AssertExpectations(t)
	mockDrift.AssertExpectations(t)
}

func TestApplication_GenerateReport(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	driftResults := map[string]*interfaces.DriftResult{
		"i-1234567890abcdef0": {
			ResourceID:     "i-1234567890abcdef0",
			ResourceType:   "aws_instance",
			IsDrifted:      true,
			DriftDetails:   []*interfaces.DriftDetail{},
			Severity:       interfaces.SeverityMedium,
			DetectionTime:  time.Now(),
		},
	}

	expectedData := []byte(`{"i-1234567890abcdef0":{"resource_id":"i-1234567890abcdef0","is_drifted":true}}`)

	t.Run("JSON format", func(t *testing.T) {
		mockReport.On("GenerateJSONReport", driftResults).Return(expectedData, nil)

		data, err := app.GenerateReport(driftResults, "json")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		mockReport.AssertExpectations(t)
	})

	t.Run("YAML format", func(t *testing.T) {
		mockReport.ExpectedCalls = nil // Reset mock
		mockReport.On("GenerateYAMLReport", driftResults).Return(expectedData, nil)

		data, err := app.GenerateReport(driftResults, "yaml")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		mockReport.AssertExpectations(t)
	})

	t.Run("Default to JSON", func(t *testing.T) {
		mockReport.ExpectedCalls = nil // Reset mock
		mockReport.On("GenerateJSONReport", driftResults).Return(expectedData, nil)

		data, err := app.GenerateReport(driftResults, "invalid")

		assert.NoError(t, err)
		assert.Equal(t, expectedData, data)
		mockReport.AssertExpectations(t)
	})
}

func TestApplication_Lifecycle(t *testing.T) {
	cfg := &config.Config{}
	cfg.SetDefaults()
	mockEC2 := &MockEC2Client{}
	mockTF := &MockTerraformParser{}
	mockDrift := &MockDriftDetector{}
	mockReport := &MockReportGenerator{}

	app := New(cfg, mockEC2, mockTF, mockDrift, mockReport)

	// Test Start
	err := app.Start()
	assert.NoError(t, err)
	assert.False(t, app.IsShuttingDown())

	// Test Shutdown
	go func() {
		time.Sleep(100 * time.Millisecond)
		app.Shutdown()
	}()

	// Test Wait
	app.Wait()
	assert.True(t, app.IsShuttingDown())
}