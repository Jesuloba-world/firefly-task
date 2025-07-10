# Firefly API Reference

## Overview

This document provides detailed API reference for Firefly's core interfaces and data structures. All interfaces are designed to be mockable for testing and extensible for future enhancements.

## Core Interfaces

### AWS Client Interfaces

#### EC2Client

Interface for EC2 operations.

```go
type EC2Client interface {
    // DescribeInstances retrieves EC2 instance details by instance IDs
    DescribeInstances(ctx context.Context, instanceIDs []string) ([]EC2Instance, error)
    
    // GetInstancesByTags retrieves EC2 instances filtered by tags
    GetInstancesByTags(ctx context.Context, tags map[string]string) ([]EC2Instance, error)
    
    // GetInstancesByRegion retrieves all EC2 instances in a specific region
    GetInstancesByRegion(ctx context.Context, region string) ([]EC2Instance, error)
    
    // GetInstanceSecurityGroups retrieves security groups for an instance
    GetInstanceSecurityGroups(ctx context.Context, instanceID string) ([]SecurityGroup, error)
}
```

**Methods:**

- **DescribeInstances**
  - **Parameters:**
    - `ctx`: Context for cancellation and timeouts
    - `instanceIDs`: Slice of EC2 instance IDs to describe
  - **Returns:** Slice of `EC2Instance` structs and error
  - **Errors:** `AWSError` if API call fails, `ValidationError` if instance IDs are invalid

- **GetInstancesByTags**
  - **Parameters:**
    - `ctx`: Context for cancellation and timeouts
    - `tags`: Map of tag key-value pairs for filtering
  - **Returns:** Slice of `EC2Instance` structs and error
  - **Errors:** `AWSError` if API call fails

#### S3Client

Interface for S3 operations.

```go
type S3Client interface {
    // GetObject retrieves an object from S3
    GetObject(ctx context.Context, bucket, key string) ([]byte, error)
    
    // PutObject stores an object in S3
    PutObject(ctx context.Context, bucket, key string, data []byte) error
    
    // ListObjects lists objects in an S3 bucket with optional prefix
    ListObjects(ctx context.Context, bucket, prefix string) ([]S3Object, error)
    
    // DeleteObject deletes an object from S3
    DeleteObject(ctx context.Context, bucket, key string) error
}
```

#### AWSClientFactory

Factory for creating AWS clients.

```go
type AWSClientFactory interface {
    // CreateEC2Client creates an EC2 client for the specified region
    CreateEC2Client(region string) (EC2Client, error)
    
    // CreateS3Client creates an S3 client
    CreateS3Client() (S3Client, error)
    
    // SetCredentials configures AWS credentials
    SetCredentials(accessKey, secretKey, sessionToken string) error
    
    // SetProfile sets the AWS profile to use
    SetProfile(profile string) error
}
```

### Terraform Interfaces

#### TerraformParser

Interface for parsing Terraform configurations.

```go
type TerraformParser interface {
    // ParseConfiguration parses Terraform files in the specified directory
    ParseConfiguration(path string) (*TerraformConfig, error)
    
    // ExtractResources extracts specific resource types from configuration
    ExtractResources(config *TerraformConfig, resourceType string) ([]TerraformResource, error)
    
    // ValidateConfiguration validates Terraform configuration syntax
    ValidateConfiguration(path string) error
    
    // GetResourceByName retrieves a specific resource by name
    GetResourceByName(config *TerraformConfig, resourceType, name string) (*TerraformResource, error)
}
```

#### TerraformExecutor

Interface for executing Terraform commands.

```go
type TerraformExecutor interface {
    // Plan generates a Terraform execution plan
    Plan(ctx context.Context, workdir string, options *PlanOptions) (*TerraformPlan, error)
    
    // Apply applies Terraform changes
    Apply(ctx context.Context, workdir string, options *ApplyOptions) error
    
    // Show displays Terraform state or plan
    Show(ctx context.Context, workdir string, target string) (*TerraformState, error)
    
    // Init initializes a Terraform working directory
    Init(ctx context.Context, workdir string) error
    
    // Validate validates Terraform configuration
    Validate(ctx context.Context, workdir string) error
}
```

#### TerraformStateManager

Interface for managing Terraform state.

```go
type TerraformStateManager interface {
    // GetState retrieves current Terraform state
    GetState(ctx context.Context, workdir string) (*TerraformState, error)
    
    // GetRemoteState retrieves state from remote backend
    GetRemoteState(ctx context.Context, backend *BackendConfig) (*TerraformState, error)
    
    // CompareStates compares two Terraform states
    CompareStates(state1, state2 *TerraformState) (*StateComparison, error)
    
    // ExtractResourceState extracts state for specific resources
    ExtractResourceState(state *TerraformState, resourceType string) ([]ResourceState, error)
}
```

### Drift Detection Interfaces

#### DriftDetector

Main interface for drift detection.

```go
type DriftDetector interface {
    // DetectDrift compares expected vs actual resource state
    DetectDrift(ctx context.Context, expected, actual interface{}) (*DriftResult, error)
    
    // DetectInstanceDrift specifically for EC2 instances
    DetectInstanceDrift(ctx context.Context, expectedInstance *TerraformResource, actualInstance *EC2Instance) (*DriftResult, error)
    
    // SetIgnoreAttributes configures attributes to ignore during detection
    SetIgnoreAttributes(attributes []string)
    
    // SetSeverityRules configures severity assessment rules
    SetSeverityRules(rules map[string]SeverityLevel)
}
```

#### DriftComparator

Interface for detailed attribute comparison.

```go
type DriftComparator interface {
    // Compare performs detailed comparison between two objects
    Compare(expected, actual interface{}) (*ComparisonResult, error)
    
    // CompareAttributes compares specific attributes
    CompareAttributes(expected, actual map[string]interface{}, attributes []string) ([]AttributeDifference, error)
    
    // SetIgnoreAttributes sets attributes to ignore during comparison
    SetIgnoreAttributes(attributes []string)
    
    // SetCustomComparators sets custom comparison functions for specific attributes
    SetCustomComparators(comparators map[string]ComparatorFunc)
}
```

#### DriftAnalyzer

Interface for analyzing drift results.

```go
type DriftAnalyzer interface {
    // AnalyzeDrift performs analysis on drift detection results
    AnalyzeDrift(ctx context.Context, results []DriftResult) (*DriftAnalysis, error)
    
    // CalculateSeverity determines severity level for differences
    CalculateSeverity(differences []AttributeDifference) SeverityLevel
    
    // GenerateRecommendations provides recommendations based on drift
    GenerateRecommendations(analysis *DriftAnalysis) ([]Recommendation, error)
    
    // GroupByAttribute groups drift results by attribute type
    GroupByAttribute(results []DriftResult) map[string][]DriftResult
}
```

### Report Generation Interfaces

#### ReportGenerator

Main interface for report generation.

```go
type ReportGenerator interface {
    // GenerateReport creates a report from drift detection results
    GenerateReport(ctx context.Context, results []DriftResult) (*Report, error)
    
    // GenerateSummaryReport creates a summary report
    GenerateSummaryReport(ctx context.Context, results []DriftResult) (*SummaryReport, error)
    
    // SetFormat configures the output format
    SetFormat(format ReportFormat)
    
    // SetTemplate sets a custom report template
    SetTemplate(template *ReportTemplate) error
}
```

#### ReportWriter

Interface for writing reports to various outputs.

```go
type ReportWriter interface {
    // WriteReport writes a report to the specified output
    WriteReport(report *Report, output string) error
    
    // WriteToFile writes a report to a file
    WriteToFile(report *Report, filename string) error
    
    // WriteToStdout writes a report to standard output
    WriteToStdout(report *Report) error
    
    // WriteToS3 writes a report to S3
    WriteToS3(report *Report, bucket, key string) error
}
```

#### ReportFormatter

Interface for formatting reports in different formats.

```go
type ReportFormatter interface {
    // Format formats a report according to the specified format
    Format(report *Report) ([]byte, error)
    
    // GetSupportedFormats returns supported output formats
    GetSupportedFormats() []ReportFormat
    
    // SetOptions configures formatter-specific options
    SetOptions(options map[string]interface{}) error
    
    // ValidateFormat validates if the format is supported
    ValidateFormat(format ReportFormat) error
}
```

#### ReportFilter

Interface for filtering report data.

```go
type ReportFilter interface {
    // FilterByAttribute filters results by specific attributes
    FilterByAttribute(results []DriftResult, attribute string, values []interface{}) []DriftResult
    
    // FilterBySeverity filters results by severity level
    FilterBySeverity(results []DriftResult, minSeverity SeverityLevel) []DriftResult
    
    // FilterByInstance filters results by instance IDs
    FilterByInstance(results []DriftResult, instanceIDs []string) []DriftResult
    
    // ApplyCustomFilter applies a custom filter function
    ApplyCustomFilter(results []DriftResult, filterFunc FilterFunc) []DriftResult
}
```

## Data Structures

### Core Types

#### EC2Instance

Represents an EC2 instance.

```go
type EC2Instance struct {
    InstanceID       string            `json:"instance_id"`
    InstanceType     string            `json:"instance_type"`
    State            string            `json:"state"`
    SubnetID         string            `json:"subnet_id"`
    VpcID            string            `json:"vpc_id"`
    AvailabilityZone string            `json:"availability_zone"`
    SecurityGroups   []string          `json:"security_groups"`
    Tags             map[string]string `json:"tags"`
    LaunchTime       time.Time         `json:"launch_time"`
    PublicIP         string            `json:"public_ip,omitempty"`
    PrivateIP        string            `json:"private_ip,omitempty"`
    KeyName          string            `json:"key_name,omitempty"`
    IamInstanceProfile string          `json:"iam_instance_profile,omitempty"`
}
```

#### TerraformResource

Represents a Terraform resource.

```go
type TerraformResource struct {
    Type         string                 `json:"type"`
    Name         string                 `json:"name"`
    Provider     string                 `json:"provider"`
    Attributes   map[string]interface{} `json:"attributes"`
    Dependencies []string               `json:"dependencies,omitempty"`
    Count        int                    `json:"count,omitempty"`
    ForEach      map[string]interface{} `json:"for_each,omitempty"`
}
```

#### DriftResult

Represents the result of drift detection.

```go
type DriftResult struct {
    InstanceID      string                 `json:"instance_id"`
    ResourceName    string                 `json:"resource_name"`
    DriftDetected   bool                   `json:"drift_detected"`
    Differences     []AttributeDifference  `json:"differences"`
    Severity        SeverityLevel          `json:"severity"`
    Timestamp       time.Time              `json:"timestamp"`
    Metadata        map[string]interface{} `json:"metadata,omitempty"`
    Recommendations []Recommendation       `json:"recommendations,omitempty"`
}
```

#### AttributeDifference

Represents a difference in a specific attribute.

```go
type AttributeDifference struct {
    Attribute   string        `json:"attribute"`
    Expected    interface{}   `json:"expected"`
    Actual      interface{}   `json:"actual"`
    Severity    SeverityLevel `json:"severity"`
    Description string        `json:"description,omitempty"`
    Path        string        `json:"path,omitempty"`
}
```

#### Report

Represents a generated report.

```go
type Report struct {
    Summary     ReportSummary  `json:"summary"`
    Results     []DriftResult  `json:"results"`
    Metadata    ReportMetadata `json:"metadata"`
    Format      ReportFormat   `json:"format"`
    GeneratedAt time.Time      `json:"generated_at"`
}
```

#### ReportSummary

Summary information for a report.

```go
type ReportSummary struct {
    TotalInstances    int                        `json:"total_instances"`
    DriftDetected     bool                       `json:"drift_detected"`
    DriftCount        int                        `json:"drift_count"`
    SeverityBreakdown map[SeverityLevel]int      `json:"severity_breakdown"`
    AttributeBreakdown map[string]int            `json:"attribute_breakdown"`
    ExecutionTime     time.Duration              `json:"execution_time"`
}
```

### Enums and Constants

#### SeverityLevel

```go
type SeverityLevel string

const (
    SeverityLow      SeverityLevel = "low"
    SeverityMedium   SeverityLevel = "medium"
    SeverityHigh     SeverityLevel = "high"
    SeverityCritical SeverityLevel = "critical"
)
```

#### ReportFormat

```go
type ReportFormat string

const (
    FormatJSON     ReportFormat = "json"
    FormatYAML     ReportFormat = "yaml"
    FormatHTML     ReportFormat = "html"
    FormatMarkdown ReportFormat = "markdown"
    FormatCSV      ReportFormat = "csv"
    FormatXML      ReportFormat = "xml"
)
```

#### ErrorType

```go
type ErrorType string

const (
    ValidationError ErrorType = "validation"
    AWSError       ErrorType = "aws"
    TerraformError ErrorType = "terraform"
    ConfigError    ErrorType = "config"
    InternalError  ErrorType = "internal"
    NetworkError   ErrorType = "network"
    AuthError      ErrorType = "authentication"
)
```

## Error Handling

### Custom Error Types

#### FireflyError

Base error type for all Firefly errors.

```go
type FireflyError struct {
    Type      ErrorType              `json:"type"`
    Message   string                 `json:"message"`
    Code      string                 `json:"code,omitempty"`
    Details   map[string]interface{} `json:"details,omitempty"`
    Cause     error                  `json:"-"`
    Timestamp time.Time              `json:"timestamp"`
}

func (e *FireflyError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *FireflyError) Unwrap() error {
    return e.Cause
}
```

### Error Creation Functions

```go
// NewValidationError creates a validation error
func NewValidationError(message string, details map[string]interface{}) *FireflyError

// NewAWSError creates an AWS-related error
func NewAWSError(message string, cause error) *FireflyError

// NewTerraformError creates a Terraform-related error
func NewTerraformError(message string, cause error) *FireflyError

// NewConfigError creates a configuration error
func NewConfigError(message string, details map[string]interface{}) *FireflyError
```

## Context and Cancellation

All long-running operations accept a `context.Context` parameter for:

- **Cancellation**: Graceful shutdown of operations
- **Timeouts**: Automatic timeout handling
- **Request Tracing**: Distributed tracing support
- **Metadata**: Passing request-scoped data

### Example Usage

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := driftDetector.DetectDrift(ctx, expected, actual)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        // Handle timeout
    }
    // Handle other errors
}
```

## Configuration

### Configuration Structures

#### Config

Main configuration structure.

```go
type Config struct {
    AWS       AWSConfig       `yaml:"aws"`
    Logging   LoggingConfig   `yaml:"logging"`
    Reporting ReportingConfig `yaml:"reporting"`
    Drift     DriftConfig     `yaml:"drift"`
}
```

#### AWSConfig

```go
type AWSConfig struct {
    Region     string        `yaml:"region"`
    Profile    string        `yaml:"profile"`
    Timeout    time.Duration `yaml:"timeout"`
    MaxRetries int           `yaml:"max_retries"`
}
```

#### LoggingConfig

```go
type LoggingConfig struct {
    Level     string `yaml:"level"`
    Format    string `yaml:"format"`
    Output    string `yaml:"output"`
    Timestamp bool   `yaml:"timestamp"`
    Caller    bool   `yaml:"caller"`
}
```

## Testing Utilities

### Mock Implementations

Firefly provides mock implementations for all interfaces:

```go
// MockEC2Client for testing
type MockEC2Client struct {
    DescribeInstancesFunc func(ctx context.Context, instanceIDs []string) ([]EC2Instance, error)
    GetInstancesByTagsFunc func(ctx context.Context, tags map[string]string) ([]EC2Instance, error)
}

// MockDriftDetector for testing
type MockDriftDetector struct {
    DetectDriftFunc func(ctx context.Context, expected, actual interface{}) (*DriftResult, error)
}
```

### Test Helpers

```go
// CreateTestEC2Instance creates a test EC2 instance
func CreateTestEC2Instance(instanceID string) *EC2Instance

// CreateTestTerraformResource creates a test Terraform resource
func CreateTestTerraformResource(resourceType, name string) *TerraformResource

// CreateTestDriftResult creates a test drift result
func CreateTestDriftResult(instanceID string, hasDrift bool) *DriftResult
```