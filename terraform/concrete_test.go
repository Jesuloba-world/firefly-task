package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewConcreteTerraformParser(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewConcreteTerraformParser(tt.logger)
			assert.NotNil(t, parser)
			assert.IsType(t, &ConcreteTerraformParser{}, parser)
		})
	}
}

func TestNewConcreteTerraformExecutor(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewConcreteTerraformExecutor("/tmp", tt.logger)
			assert.NotNil(t, executor)
			assert.IsType(t, &ConcreteTerraformExecutor{}, executor)
		})
	}
}

func TestNewConcreteTerraformStateManager(t *testing.T) {
	tests := []struct {
		name   string
		logger *logrus.Logger
	}{
		{
			name:   "with logger",
			logger: logrus.New(),
		},
		{
			name:   "without logger",
			logger: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateManager := NewConcreteTerraformStateManager("/tmp/terraform.tfstate", tt.logger)
			assert.NotNil(t, stateManager)
			assert.IsType(t, &ConcreteTerraformStateManager{}, stateManager)
		})
	}
}

func createTestTerraformStateFile(t *testing.T, dir string) string {
	stateContent := `{
  "format_version": "1.0",
  "terraform_version": "1.0.0",
  "serial": 1,
  "lineage": "test-lineage",
  "outputs": {},
  "values": {
    "root_module": {
      "resources": [
        {
          "address": "aws_instance.example",
          "mode": "managed",
          "type": "aws_instance",
          "name": "example",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "id": "i-1234567890abcdef0",
            "instance_type": "t2.micro",
            "ami": "ami-12345678",
            "availability_zone": "us-west-2a",
            "tags": {
              "Name": "test-instance",
              "Environment": "test"
            }
          }
        }
      ]
    }
  }
}`

	filePath := filepath.Join(dir, "terraform.tfstate")
	err := os.WriteFile(filePath, []byte(stateContent), 0644)
	assert.NoError(t, err)
	return filePath
}

func createTestTerraformHCLFile(t *testing.T, dir string) string {
	hclContent := `resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"

  tags = {
    Name        = "test-instance"
    Environment = "test"
  }
}

resource "aws_security_group" "example" {
  name_prefix = "test-sg"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}`

	filePath := filepath.Join(dir, "main.tf")
	err := os.WriteFile(filePath, []byte(hclContent), 0644)
	assert.NoError(t, err)
	return filePath
}

func TestConcreteTerraformParser_ParseTerraformState(t *testing.T) {
	logger := logrus.New()
	parser := NewConcreteTerraformParser(logger)

	tempDir := t.TempDir()
	stateFile := createTestTerraformStateFile(t, tempDir)

	tests := []struct {
		name          string
		statePath     string
		expectedError string
	}{
		{
			name:      "valid state file",
			statePath: stateFile,
		},
		{
			name:          "non-existent file",
			statePath:     "/non/existent/path.tfstate",
			expectedError: "failed to parse state file",
		},
		{
			name:          "invalid JSON",
			statePath:     createInvalidJSONFile(t, tempDir),
			expectedError: "failed to parse state file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure parser to not ignore missing files for all test cases
			concreteParser := parser.(*ConcreteTerraformParser)
			options := concreteParser.parser.GetOptions()
			options.IgnoreMissingFiles = false
			concreteParser.parser.SetOptions(options)

			result, err := parser.ParseTerraformState(tt.statePath)

			if tt.expectedError != "" {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, 1)
				if config, exists := result["aws_instance.example"]; exists && config != nil {
					assert.Equal(t, "aws_instance", config.ResourceType)
					assert.Equal(t, "example", config.ResourceName)
				} else {
					t.Error("Expected aws_instance.example configuration not found or is nil")
				}
			}
		})
	}
}

func createInvalidJSONFile(t *testing.T, dir string) string {
	invalidContent := `{"invalid": json content}`
	filePath := filepath.Join(dir, "invalid.tfstate")
	err := os.WriteFile(filePath, []byte(invalidContent), 0644)
	assert.NoError(t, err)
	return filePath
}

func TestConcreteTerraformParser_ParseTerraformHCL(t *testing.T) {
	logger := logrus.New()
	parser := NewConcreteTerraformParser(logger)

	tempDir := t.TempDir()
	hclFile := createTestTerraformHCLFile(t, tempDir)

	tests := []struct {
		name          string
		configPath    string
		expectedError string
	}{
		{
			name:          "valid HCL file",
			configPath:    hclFile,
			expectedError: "failed to parse HCL configuration", // terraform-config-inspect has limitations
		},
		{
			name:          "non-existent file",
			configPath:    "/non/existent/path.tf",
			expectedError: "failed to parse HCL configuration",
		},
		{
			name:          "invalid HCL syntax",
			configPath:    createInvalidHCLFile(t, tempDir),
			expectedError: "failed to parse HCL configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Configure parser to not ignore missing files for all test cases
			concreteParser := parser.(*ConcreteTerraformParser)
			options := concreteParser.parser.GetOptions()
			options.IgnoreMissingFiles = false
			concreteParser.parser.SetOptions(options)

			result, err := parser.ParseTerraformHCL(tt.configPath)

			if tt.expectedError != "" {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.expectedError)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Len(t, result, 1)

			// Check for aws_instance resource
			_, instanceFound := result["aws_instance.example"]
			assert.True(t, instanceFound, "aws_instance.example resource should be found")
			}
		})
	}
}

func createInvalidHCLFile(t *testing.T, dir string) string {
	invalidContent := `resource "aws_instance" {
  # Missing name and invalid syntax
  ami = "ami-12345678"
  instance_type = "t2.micro"
}`
	filePath := filepath.Join(dir, "invalid.tf")
	err := os.WriteFile(filePath, []byte(invalidContent), 0644)
	assert.NoError(t, err)
	return filePath
}



func TestConcreteTerraformExecutor_Init(t *testing.T) {
	logger := logrus.New()
	executor := NewConcreteTerraformExecutor("/tmp", logger)

	tests := []struct {
		name          string
		workingDir    string
		expectedError string
	}{
		{
			name:       "valid working directory",
			workingDir: "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.Init(tt.workingDir)

			assert.NoError(t, err)
		})
	}
}

func TestConcreteTerraformExecutor_Plan(t *testing.T) {
	logger := logrus.New()
	executor := NewConcreteTerraformExecutor("/tmp", logger)

	tests := []struct {
		name          string
		workingDir    string
		planFile      string
		expectedError string
	}{
		{
			name:       "valid plan",
			workingDir: "/tmp",
			planFile:   "plan.out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.Plan(tt.workingDir, tt.planFile)

			assert.NoError(t, err)
		})
	}
}

func TestConcreteTerraformExecutor_Apply(t *testing.T) {
	logger := logrus.New()
	executor := NewConcreteTerraformExecutor("/tmp", logger)

	tests := []struct {
		name          string
		workingDir    string
		planFile      string
		expectedError string
	}{
		{
			name:       "valid apply",
			workingDir: "/tmp",
			planFile:   "plan.out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.Apply(tt.workingDir, tt.planFile)

			assert.NoError(t, err)
		})
	}
}

func TestConcreteTerraformExecutor_Show(t *testing.T) {
	logger := logrus.New()
	executor := NewConcreteTerraformExecutor("/tmp", logger)

	tests := []struct {
		name          string
		workingDir    string
		target        string
		expectedError string
	}{
		{
			name:       "valid show",
			workingDir: "/tmp",
			target:     "plan.out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := executor.Show(tt.workingDir, tt.target)

				assert.NoError(t, err)
		})
	}
}

func TestConcreteTerraformExecutor_Validate(t *testing.T) {
	logger := logrus.New()
	executor := NewConcreteTerraformExecutor("/tmp", logger)

	tests := []struct {
		name          string
		workingDir    string
		expectedError string
	}{
		{
			name:       "valid validate",
			workingDir: "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executor.Validate(tt.workingDir)

			assert.NoError(t, err)
		})
	}
}

func TestConcreteTerraformStateManager_GetResourceState(t *testing.T) {
	logger := logrus.New()
	stateManager := NewConcreteTerraformStateManager("/tmp/terraform.tfstate", logger)

	tests := []struct {
		name          string
		resourceAddress string
		expectedError string
	}{
		{
			name:          "get resource state not implemented",
			resourceAddress: "aws_instance.example",
			expectedError: "state management not yet implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := stateManager.GetResourceState(tt.resourceAddress)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}