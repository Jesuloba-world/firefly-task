package report

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"firefly-task/pkg/interfaces"
)

// convertToValueMap converts pointer map to value map
func convertToValueMap(ptrMap map[string]*interfaces.DriftResult) map[string]interfaces.DriftResult {
	valueMap := make(map[string]interfaces.DriftResult)
	for k, v := range ptrMap {
		if v != nil {
			valueMap[k] = *v
		}
	}
	return valueMap
}

// createTestReportData creates test drift results for CICD tests
func createTestReportData() map[string]*interfaces.DriftResult {
	results := make(map[string]*interfaces.DriftResult)

	// Resource with drift
	result1 := &interfaces.DriftResult{
		ResourceID:    "aws_instance.test",
		ResourceType:  "aws_instance",
		IsDrifted:     true,
		DetectionTime: time.Now(),
		Severity:      interfaces.SeverityHigh,
		DriftDetails: []*interfaces.DriftDetail{
			{
				Attribute:     "instance_type",
				ExpectedValue: "t2.micro",
				ActualValue:   "t2.small",
				DriftType:     "changed",
				Severity:      interfaces.SeverityHigh,
			},
		},
	}
	results["aws_instance.test"] = result1

	// Resource with critical drift
	result2 := &interfaces.DriftResult{
		ResourceID:    "aws_s3_bucket.data",
		ResourceType:  "aws_s3_bucket",
		IsDrifted:     true,
		DetectionTime: time.Now(),
		Severity:      interfaces.SeverityCritical,
		DriftDetails: []*interfaces.DriftDetail{
			{
				Attribute:     "public_access_block",
				ExpectedValue: "true",
				ActualValue:   "false",
				DriftType:     "changed",
				Severity:      interfaces.SeverityCritical,
			},
		},
	}
	results["aws_s3_bucket.data"] = result2

	// Clean resource
	result3 := &interfaces.DriftResult{
		ResourceID:    "aws_instance.clean",
		ResourceType:  "aws_instance",
		IsDrifted:     false,
		DetectionTime: time.Now(),
		Severity:      interfaces.SeverityNone,
		DriftDetails:  []*interfaces.DriftDetail{},
	}
	results["aws_instance.clean"] = result3

	return results
}

func TestCICDPlatform_String(t *testing.T) {
	tests := []struct {
		platform CICDPlatform
		expected string
	}{
		{PlatformGitHubActions, "github-actions"},
		{PlatformGitLab, "gitlab"},
		{PlatformJenkins, "jenkins"},
		{PlatformAzureDevOps, "azure-devops"},
		{PlatformCircleCI, "circleci"},
		{PlatformGeneric, "generic"},
		{CICDPlatform("unknown-platform"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.platform.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectCICDPlatform(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "JENKINS_HOME",
		"AZURE_HTTP_USER_AGENT", "TF_BUILD", "CIRCLECI", "CI",
	}

	for _, envVar := range envVars {
		originalEnv[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	defer func() {
		// Restore original environment
		for envVar, value := range originalEnv {
			if value != "" {
				os.Setenv(envVar, value)
			} else {
				os.Unsetenv(envVar)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected CICDPlatform
	}{
		{
			name:     "GitHub Actions",
			envVars:  map[string]string{"GITHUB_ACTIONS": "true"},
			expected: PlatformGitHubActions,
		},
		{
			name:     "GitLab CI",
			envVars:  map[string]string{"GITLAB_CI": "true"},
			expected: PlatformGitLab,
		},
		{
			name:     "Jenkins with URL",
			envVars:  map[string]string{"JENKINS_URL": "http://jenkins.example.com"},
			expected: PlatformJenkins,
		},
		{
			name:     "Jenkins with HOME",
			envVars:  map[string]string{"JENKINS_HOME": "/var/jenkins_home"},
			expected: PlatformJenkins,
		},
		{
			name:     "Azure DevOps with USER_AGENT",
			envVars:  map[string]string{"AZURE_HTTP_USER_AGENT": "azure-pipelines"},
			expected: PlatformAzureDevOps,
		},
		{
			name:     "Azure DevOps with TF_BUILD",
			envVars:  map[string]string{"TF_BUILD": "True"},
			expected: PlatformAzureDevOps,
		},
		{
			name:     "CircleCI",
			envVars:  map[string]string{"CIRCLECI": "true"},
			expected: PlatformCircleCI,
		},
		{
			name:     "Generic CI",
			envVars:  map[string]string{"CI": "true"},
			expected: PlatformGeneric,
		},
		{
			name:     "No CI detected",
			envVars:  map[string]string{},
			expected: PlatformGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all CI environment variables
			for _, envVar := range envVars {
				os.Unsetenv(envVar)
			}

			// Set test environment variables
			for envVar, value := range tt.envVars {
				os.Setenv(envVar, value)
			}

			result := DetectCICDPlatform()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewCIReportGenerator(t *testing.T) {
	generator := NewCIReportGenerator()

	assert.NotNil(t, generator)
	assert.NotEqual(t, CICDPlatform(""), generator.Platform)
}

func TestCIReportGenerator_GenerateCIReport(t *testing.T) {
	generator := NewCIReportGenerator()
	data := createTestReportData()

	report, err := generator.GenerateCIReport(convertToValueMap(data))
	require.NoError(t, err)
	require.NotNil(t, report)

	// Verify report structure
	assert.NotEmpty(t, report.Summary.TotalResources)
	assert.NotEmpty(t, report.Actions)
	assert.NotEmpty(t, report.Metadata.Platform)
	assert.NotZero(t, report.Metadata.Timestamp)
	assert.NotEmpty(t, report.Metadata.Version)
}

func TestCIReportGenerator_WriteArtifacts(t *testing.T) {
	generator := NewCIReportGenerator()
	data := createTestReportData()

	artifacts, err := generator.WriteArtifacts(data)
	require.NoError(t, err)
	require.NotEmpty(t, artifacts)

	// Verify artifacts were created
	for _, artifact := range artifacts {
		assert.NotEmpty(t, artifact.Path)
		assert.NotEmpty(t, artifact.Type)
		assert.Greater(t, artifact.Size, int64(0))

		// Verify file exists
		_, err := os.Stat(artifact.Path)
		require.NoError(t, err)
	}

	// Check for expected artifact types
	artifactTypes := make(map[string]bool)
	for _, artifact := range artifacts {
		artifactTypes[artifact.Type] = true
	}

	assert.True(t, artifactTypes["json"])
	assert.True(t, artifactTypes["junit-xml"])
	assert.True(t, artifactTypes["summary"])
}

func TestCIReportGenerator_WriteJSONArtifact(t *testing.T) {
	generator := NewCIReportGenerator()
	data := createTestReportData()

	artifact, err := generator.WriteJSONArtifact(data)
	require.NoError(t, err)
	require.NotNil(t, artifact)

	assert.Equal(t, "json", artifact.Type)
	assert.True(t, strings.HasSuffix(artifact.Path, ".json"))
	assert.Greater(t, artifact.Size, int64(0))

	// Verify file content
	content, err := os.ReadFile(artifact.Path)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	require.NoError(t, err)
	assert.Contains(t, jsonData, "summary")
	assert.Contains(t, jsonData, "actions")
	assert.Contains(t, jsonData, "metadata")
}

func TestCIReportGenerator_WriteJUnitXMLArtifact(t *testing.T) {
	generator := NewCIReportGenerator()
	data := createTestReportData()

	artifact, err := generator.WriteJUnitXMLArtifact(data)
	require.NoError(t, err)
	require.NotNil(t, artifact)

	assert.Equal(t, "junit-xml", artifact.Type)
	assert.True(t, strings.HasSuffix(artifact.Path, ".xml"))
	assert.Greater(t, artifact.Size, int64(0))

	// Verify XML content
	content, err := os.ReadFile(artifact.Path)
	require.NoError(t, err)

	var testSuite struct {
		XMLName xml.Name `xml:"testsuite"`
		Name    string   `xml:"name,attr"`
		Tests   int      `xml:"tests,attr"`
	}

	err = xml.Unmarshal(content, &testSuite)
	require.NoError(t, err)
	assert.Equal(t, "Terraform Drift Detection", testSuite.Name)
	assert.Greater(t, testSuite.Tests, 0)
}

func TestCIReportGenerator_WriteSummaryArtifact(t *testing.T) {
	generator := NewCIReportGenerator()
	data := createTestReportData()

	artifact, err := generator.WriteSummaryArtifact(data)
	require.NoError(t, err)
	require.NotNil(t, artifact)

	assert.Equal(t, "summary", artifact.Type)
	assert.True(t, strings.HasSuffix(artifact.Path, ".md"))
	assert.Greater(t, artifact.Size, int64(0))

	// Verify markdown content
	content, err := os.ReadFile(artifact.Path)
	require.NoError(t, err)

	summaryText := string(content)
	assert.Contains(t, summaryText, "# Terraform Drift Detection Summary")
	assert.Contains(t, summaryText, "## Summary")
	assert.Contains(t, summaryText, "Total Resources")
}

func TestCIReportGenerator_SetExitCode(t *testing.T) {
	generator := NewCIReportGenerator()

	tests := []struct {
		name     string
		severity interfaces.SeverityLevel
		expected int
	}{
		{"No drift", interfaces.SeverityNone, 0},
		{"Low severity", interfaces.SeverityLow, 0},
		{"Medium severity", interfaces.SeverityMedium, 0},
		{"High severity", interfaces.SeverityHigh, 1},
		{"Critical severity", interfaces.SeverityCritical, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test data with the specified severity
			results := make(map[string]*interfaces.DriftResult)
			if tt.severity != interfaces.SeverityNone {
				severity := tt.severity
				results["test-resource"] = &interfaces.DriftResult{
					ResourceID:    "test-resource",
					ResourceType:  "test_resource",
					IsDrifted:     true,
					DetectionTime: time.Now(),
					Severity:      severity,
					DriftDetails: []*interfaces.DriftDetail{
						{
							Attribute:     "test_attribute",
							ExpectedValue: "expected",
							ActualValue:   "actual",
							DriftType:     "changed",
							Severity:      severity,
						},
					},
				}
			}
			result := generator.SetExitCode(results)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCIReportGenerator_SetEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalVars := map[string]string{
		"DRIFT_TOTAL_RESOURCES":      os.Getenv("DRIFT_TOTAL_RESOURCES"),
		"DRIFT_RESOURCES_WITH_DRIFT": os.Getenv("DRIFT_RESOURCES_WITH_DRIFT"),
		"DRIFT_MAX_SEVERITY":         os.Getenv("DRIFT_MAX_SEVERITY"),
		"DRIFT_CRITICAL_COUNT":       os.Getenv("DRIFT_CRITICAL_COUNT"),
		"DRIFT_HIGH_COUNT":           os.Getenv("DRIFT_HIGH_COUNT"),
		"DRIFT_MEDIUM_COUNT":         os.Getenv("DRIFT_MEDIUM_COUNT"),
		"DRIFT_LOW_COUNT":            os.Getenv("DRIFT_LOW_COUNT"),
	}

	defer func() {
		// Restore original environment
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	generator := NewCIReportGenerator()
	data := createTestReportData()

	err := generator.SetEnvironmentVariables(data)
	require.NoError(t, err)

	// Verify environment variables are set
	assert.NotEmpty(t, os.Getenv("DRIFT_TOTAL_RESOURCES"))
	assert.NotEmpty(t, os.Getenv("DRIFT_RESOURCES_WITH_DRIFT"))
	assert.NotEmpty(t, os.Getenv("DRIFT_MAX_SEVERITY"))
	assert.NotEmpty(t, os.Getenv("DRIFT_CRITICAL_COUNT"))
	assert.NotEmpty(t, os.Getenv("DRIFT_HIGH_COUNT"))
	assert.NotEmpty(t, os.Getenv("DRIFT_MEDIUM_COUNT"))
	assert.NotEmpty(t, os.Getenv("DRIFT_LOW_COUNT"))
}

func TestCIReportGenerator_SetPlatformSpecificVariables(t *testing.T) {
	// Save original environment
	originalVars := map[string]string{
		"GITHUB_OUTPUT":       os.Getenv("GITHUB_OUTPUT"),
		"GITHUB_STEP_SUMMARY": os.Getenv("GITHUB_STEP_SUMMARY"),
	}

	defer func() {
		// Restore original environment
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	tempDir := t.TempDir()
	generator := NewCIReportGenerator()
	generator.OutputDir = tempDir
	generator.Platform = PlatformGitHubActions

	// Create temporary files for GitHub Actions
	githubOutput := filepath.Join(tempDir, "github_output")
	githubSummary := filepath.Join(tempDir, "github_summary")

	// Create the files before setting them as environment variables
	f, err := os.Create(githubOutput)
	require.NoError(t, err)
	f.Close()

	f, err = os.Create(githubSummary)
	require.NoError(t, err)
	f.Close()

	os.Setenv("GITHUB_OUTPUT", githubOutput)
	os.Setenv("GITHUB_STEP_SUMMARY", githubSummary)

	data := createTestReportData()

	err = generator.SetPlatformSpecificVariables(data)
	require.NoError(t, err)

	// Verify GitHub Actions files are created
	_, err = os.Stat(githubOutput)
	assert.NoError(t, err)

	_, err = os.Stat(githubSummary)
	assert.NoError(t, err)

	// Verify content
	outputBytes, err := os.ReadFile(githubOutput)
	require.NoError(t, err)
	outputContent := string(outputBytes)

	assert.Contains(t, outputContent, "drift_has_drift=true")

	summaryContent, err := os.ReadFile(githubSummary)
	require.NoError(t, err)
	assert.Contains(t, string(summaryContent), "# Terraform Drift Detection")
}

func TestCIReportGenerator_GetArtifactInfo(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewCIReportGenerator()
	generator.OutputDir = tempDir
	data := createTestReportData()

	// Create some artifacts
	_, err := generator.WriteArtifacts(data)
	require.NoError(t, err)

	info, err := generator.GetArtifactInfo()
	require.NoError(t, err)
	assert.Greater(t, info.FileCount, 0)
	assert.Greater(t, info.TotalSize, int64(0))
	assert.NotEmpty(t, info.Files)

	// Verify file info
	for _, fileInfo := range info.Files {
		assert.NotEmpty(t, fileInfo.Name())
		assert.Greater(t, fileInfo.Size(), int64(0))
		assert.NotZero(t, fileInfo.ModTime())
	}
}

// Test edge cases
func TestCIReportGenerator_EmptyData(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewCIReportGenerator()
	generator.OutputDir = tempDir

	emptyData := make(map[string]*interfaces.DriftResult)

	// Test generating CI report with empty data
	report, err := generator.GenerateCIReport(convertToValueMap(emptyData))
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 0, report.Summary.TotalResources)

	// Test writing artifacts with empty data
	artifacts, err := generator.WriteArtifacts(emptyData)
	require.NoError(t, err)
	assert.NotEmpty(t, artifacts) // Should still create artifacts
}

func TestCIReportGenerator_NilData(t *testing.T) {
	generator := NewCIReportGenerator()

	// Test with nil data
	_, err := generator.GenerateCIReport(nil)
	assert.Error(t, err)

	_, err = generator.WriteArtifacts(nil)
	assert.Error(t, err)

	err = generator.SetPlatformSpecificVariables(nil)
	assert.Error(t, err)
}

func TestCIReportGenerator_InvalidOutputDir(t *testing.T) {
	// Test with a path that cannot be created (e.g., a file as parent directory)
	tempDir := t.TempDir()
	// Create a file where we want to create a directory
	filePath := filepath.Join(tempDir, "not-a-directory")
	err := os.WriteFile(filePath, []byte("test"), 0644)
	require.NoError(t, err)

	// Try to use the file path as a directory - this should fail
	generator := NewCIReportGenerator()
	generator.OutputDir = filepath.Join(filePath, "subdir")
	data := createTestReportData()

	_, err = generator.WriteArtifacts(data)
	assert.Error(t, err)

	_, err = generator.WriteJSONArtifact(data)
	assert.Error(t, err)

	_, err = generator.WriteJUnitXMLArtifact(data)
	assert.Error(t, err)

	_, err = generator.WriteSummaryArtifact(data)
	assert.Error(t, err)
}

type CIArtifact = Artifact

func TestCIReportGenerator_ConcurrentArtifactCreation(t *testing.T) {
	tempDir := t.TempDir()
	generator := NewCIReportGenerator()
	generator.OutputDir = tempDir
	data := createTestReportData()

	// Test concurrent artifact creation
	results := make(chan *CIArtifact, 3)
	errors := make(chan error, 3)

	go func() {
		artifact, err := generator.WriteJSONArtifact(data)
		if err != nil {
			errors <- err
		} else {
			results <- artifact
		}
	}()

	go func() {
		artifact, err := generator.WriteJUnitXMLArtifact(data)
		if err != nil {
			errors <- err
		} else {
			results <- artifact
		}
	}()

	go func() {
		artifact, err := generator.WriteSummaryArtifact(data)
		if err != nil {
			errors <- err
		} else {
			results <- artifact
		}
	}()

	// Collect results
	var artifacts []*CIArtifact
	for i := 0; i < 3; i++ {
		select {
		case artifact := <-results:
			artifacts = append(artifacts, artifact)
		case err := <-errors:
			t.Fatal(err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent artifact creation")
		}
	}

	assert.Len(t, artifacts, 3)

	// Verify all artifacts exist
	for _, artifact := range artifacts {
		_, err := os.Stat(artifact.Path)
		require.NoError(t, err)
	}
}

// Benchmark tests
func BenchmarkCIReportGenerator_GenerateCIReport(b *testing.B) {
	generator := NewCIReportGenerator()
	data := createTestReportData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.GenerateCIReport(convertToValueMap(data))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCIReportGenerator_WriteArtifacts(b *testing.B) {
	tempDir := b.TempDir()
	generator := NewCIReportGenerator()
	generator.OutputDir = tempDir
	data := createTestReportData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := generator.WriteArtifacts(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDetectCICDPlatform(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectCICDPlatform()
	}
}
