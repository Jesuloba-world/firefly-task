# Firefly - AWS EC2 Configuration Drift Detection Tool

[![Go Version](https://img.shields.io/badge/go-1.24.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0-orange.svg)](https://github.com/Jesuloba-world/firefly-task/releases)

## Project Overview

Firefly is a powerful CLI tool designed for detecting configuration drift in AWS EC2 instances. It provides comprehensive drift detection by comparing the current state of your EC2 infrastructure with the expected configuration defined in Terraform files, helping you maintain infrastructure compliance and security.

### Key Features

- **Real-time Drift Detection**: Compare live AWS EC2 instances against Terraform configurations
- **Batch Processing**: Analyze multiple instances simultaneously for efficient operations
- **Flexible Reporting**: Generate reports in JSON, YAML, HTML, and Markdown formats
- **Attribute-specific Analysis**: Focus on specific configuration attributes
- **CI/CD Integration**: Seamlessly integrate with your deployment pipelines
- **Comprehensive Logging**: Structured logging for debugging and monitoring
- **Modular Architecture**: Extensible design for future AWS service support

### Benefits

- **Security Compliance**: Ensure your infrastructure remains secure and compliant
- **Cost Optimization**: Identify configuration changes that may impact costs
- **Operational Excellence**: Maintain consistency across your infrastructure
- **Risk Mitigation**: Detect unauthorized or accidental configuration changes
- **Audit Trail**: Track configuration changes over time

## Architecture Overview

Firefly follows a modular architecture designed for scalability and maintainability:

### Core Components

1. **AWS Client Layer** (`aws/`)
   - Handles AWS API interactions
   - Manages EC2 instance data retrieval
   - Provides abstraction for AWS services

2. **Terraform Parser** (`terraform/`)
   - Parses Terraform configuration files (.tf)
   - Extracts expected infrastructure state
   - Supports HCL and JSON formats

3. **Drift Detection Engine** (`drift/`)
   - Compares actual vs. expected configurations
   - Identifies configuration discrepancies
   - Provides detailed drift analysis

4. **Report Generator** (`report/`)
   - Formats drift detection results
   - Supports multiple output formats (JSON, YAML, HTML, Markdown)
   - Enables custom reporting templates

5. **Application Framework** (`pkg/`)
   - Provides core application infrastructure
   - Handles logging, configuration, and dependency injection
   - Manages application lifecycle

### Data Flow

1. **Configuration Loading**: Terraform files are parsed to extract expected state
2. **AWS Data Retrieval**: Current EC2 instance configurations are fetched from AWS
3. **Drift Analysis**: Configurations are compared to identify differences
4. **Report Generation**: Results are formatted and output in the desired format

For detailed architecture diagrams, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Setup and Installation Instructions

### Prerequisites

- **Go**: Version 1.24.4 or higher ([Download Go](https://golang.org/dl/))
- **AWS Account**: Active AWS account with appropriate permissions
- **AWS Credentials**: Configured AWS credentials (see [AWS Configuration](#aws-configuration))
- **Git**: For cloning the repository

### Installation

```bash
# Clone the repository
git clone https://github.com/Jesuloba-world/firefly-task.git
cd firefly-task

# Build the application
go build -o firefly cmd/main/main.go

# Make it executable (Linux/macOS)
chmod +x firefly

# Move to PATH (optional)
sudo mv firefly /usr/local/bin/
```

### AWS Configuration

Firefly requires AWS credentials to access your EC2 instances. Configure them using one of these methods:

#### Method 1: AWS CLI
```bash
aws configure
```

#### Method 2: Environment Variables
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_DEFAULT_REGION="us-west-2"
```

#### Method 3: IAM Roles (Recommended for EC2/ECS)
Use IAM roles when running on AWS infrastructure.

### Required AWS Permissions

Ensure your AWS credentials have the following permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances",
                "ec2:DescribeInstanceAttribute",
                "ec2:DescribeInstanceTypes",
                "ec2:DescribeSecurityGroups",
                "ec2:DescribeSubnets",
                "ec2:DescribeVpcs"
            ],
            "Resource": "*"
        }
    ]
}
```

### Verification

Verify your installation:

```bash
# Check version
firefly --version

# View help
firefly --help

# Test AWS connectivity
firefly check --help
```

## Configuration Guide

Firefly supports various configuration options to customize its behavior:

### Configuration File

Create a `firefly.yaml` configuration file:

```yaml
# AWS Configuration
aws:
  region: "us-west-2"
  profile: "default"
  timeout: "30s"

# Logging Configuration
logging:
  level: "info"  # debug, info, warn, error
  format: "json" # json, text
  output: "stdout" # stdout, stderr, file path

# Report Configuration
reporting:
  format: "json" # json, yaml, html, markdown
  output_dir: "./reports"
  include_metadata: true

# Drift Detection Settings
drift:
  ignore_tags: ["LastModified", "CreatedBy"]
  strict_mode: false
  timeout: "60s"
```

### Environment Variables

Alternatively, use environment variables:

```bash
export FIREFLY_AWS_REGION="us-west-2"
export FIREFLY_LOG_LEVEL="debug"
export FIREFLY_REPORT_FORMAT="yaml"
export FIREFLY_OUTPUT_DIR="./reports"
```

## Usage Examples

### Basic Usage

#### Check a single instance

```bash
# Basic drift check
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/

# With custom output format
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --format yaml

# Save report to file
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --output report.json
```

#### Check multiple instances

```bash
# Check multiple instances by ID
firefly check --instance-ids i-0123456789abcdef0,i-fedcba9876543210f --tf-path /path/to/terraform/

# Check all instances in a region
firefly check --region us-west-2 --tf-path /path/to/terraform/

# Check instances by tags
firefly check --tags "Environment=production,Team=backend" --tf-path /path/to/terraform/
```

### Batch Processing

Create a file named `instances.txt` with one instance ID per line:

```
i-0123456789abcdef0
i-fedcba9876543210f
i-1234567890abcdef1
```

Then run:

```bash
# Basic batch processing
firefly batch --input-file instances.txt --tf-path /path/to/terraform/

# Batch with parallel processing
firefly batch --input-file instances.txt --tf-path /path/to/terraform/ --parallel 5

# Batch with custom report format
firefly batch --input-file instances.txt --tf-path /path/to/terraform/ --format html --output batch-report.html
```

### Attribute-Specific Analysis

```bash
# Check specific attribute
firefly attribute --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --attribute instance_type

# Check multiple attributes
firefly attribute --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --attributes instance_type,security_groups,subnet_id

# Check all attributes with detailed output
firefly attribute --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --all --verbose
```

### Advanced Usage

#### CI/CD Integration

```bash
# Exit with non-zero code if drift detected
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --fail-on-drift

# Generate machine-readable output for CI/CD
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --format json --quiet

# Check with custom timeout for CI/CD pipelines
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --timeout 30s
```

#### Filtering and Reporting

```bash
# Ignore specific attributes
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --ignore-attributes "tags.LastModified,tags.CreatedBy"

# Generate detailed HTML report
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --format html --template custom-template.html

# Include metadata in reports
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/terraform/ --include-metadata
```

### Expected Output Examples

#### JSON Output
```json
{
  "summary": {
    "total_instances": 1,
    "drift_detected": true,
    "drift_count": 2
  },
  "instances": {
    "i-0123456789abcdef0": {
      "drift_detected": true,
      "differences": [
        {
          "attribute": "instance_type",
          "expected": "t3.medium",
          "actual": "t3.large",
          "severity": "high"
        },
        {
          "attribute": "security_groups",
          "expected": ["sg-12345"],
          "actual": ["sg-12345", "sg-67890"],
          "severity": "medium"
        }
      ]
    }
  }
}
```

#### YAML Output
```yaml
summary:
  total_instances: 1
  drift_detected: true
  drift_count: 2
instances:
  i-0123456789abcdef0:
    drift_detected: true
    differences:
    - attribute: instance_type
      expected: t3.medium
      actual: t3.large
      severity: high
    - attribute: security_groups
      expected: [sg-12345]
      actual: [sg-12345, sg-67890]
      severity: medium
```

## API Documentation

### Core Interfaces

#### AWS Client Interface
```go
type EC2Client interface {
    DescribeInstances(ctx context.Context, instanceIDs []string) ([]EC2Instance, error)
    GetInstancesByTags(ctx context.Context, tags map[string]string) ([]EC2Instance, error)
}

type S3Client interface {
    GetObject(ctx context.Context, bucket, key string) ([]byte, error)
    PutObject(ctx context.Context, bucket, key string, data []byte) error
}
```

#### Terraform Interface
```go
type TerraformParser interface {
    ParseConfiguration(path string) (*TerraformConfig, error)
    ExtractResources(config *TerraformConfig) ([]TerraformResource, error)
}

type TerraformExecutor interface {
    Plan(ctx context.Context, workdir string) (*TerraformPlan, error)
    Apply(ctx context.Context, workdir string) error
    Show(ctx context.Context, workdir string) (*TerraformState, error)
}
```

#### Drift Detection Interface
```go
type DriftDetector interface {
    DetectDrift(ctx context.Context, expected, actual interface{}) (*DriftResult, error)
    CompareAttributes(expected, actual map[string]interface{}) []AttributeDifference
}

type DriftComparator interface {
    Compare(expected, actual interface{}) (*ComparisonResult, error)
    SetIgnoreAttributes(attributes []string)
}
```

#### Report Generation Interface
```go
type ReportGenerator interface {
    GenerateReport(ctx context.Context, results []DriftResult) (*Report, error)
    SetFormat(format ReportFormat)
}

type ReportWriter interface {
    WriteReport(report *Report, output string) error
    WriteToFile(report *Report, filename string) error
}
```

### Data Structures

#### Core Types
```go
type EC2Instance struct {
    InstanceID       string            `json:"instance_id"`
    InstanceType     string            `json:"instance_type"`
    State            string            `json:"state"`
    SubnetID         string            `json:"subnet_id"`
    SecurityGroups   []string          `json:"security_groups"`
    Tags             map[string]string `json:"tags"`
    LaunchTime       time.Time         `json:"launch_time"`
}

type DriftResult struct {
    InstanceID    string                 `json:"instance_id"`
    DriftDetected bool                   `json:"drift_detected"`
    Differences   []AttributeDifference  `json:"differences"`
    Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type AttributeDifference struct {
    Attribute string      `json:"attribute"`
    Expected  interface{} `json:"expected"`
    Actual    interface{} `json:"actual"`
    Severity  string      `json:"severity"`
}
```

## Troubleshooting

### Common Issues

#### AWS Authentication Errors
```
Error: Unable to locate credentials
```
**Solution:**
- Ensure AWS credentials are configured via AWS CLI, environment variables, or IAM roles
- Verify the AWS profile exists: `aws configure list-profiles`
- Check AWS credentials: `aws sts get-caller-identity`

#### Terraform Parsing Errors
```
Error: Failed to parse Terraform configuration
```
**Solution:**
- Verify Terraform files are valid: `terraform validate`
- Ensure the specified path contains `.tf` files
- Check for syntax errors in Terraform configuration

#### Instance Not Found Errors
```
Error: Instance i-1234567890abcdef0 not found
```
**Solution:**
- Verify the instance ID is correct
- Ensure you're checking the correct AWS region
- Confirm the instance exists: `aws ec2 describe-instances --instance-ids i-1234567890abcdef0`

#### Permission Denied Errors
```
Error: User is not authorized to perform: ec2:DescribeInstances
```
**Solution:**
- Ensure your AWS user/role has the required permissions
- Attach the necessary IAM policies (see Installation section)
- Verify region-specific permissions if using resource-based policies

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
# Enable debug logging
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --log-level debug

# Save debug output to file
firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/terraform/ --log-level debug --log-output debug.log
```

### Performance Optimization

#### Large-Scale Deployments
- Use batch processing for multiple instances
- Enable parallel processing: `--parallel 10`
- Implement caching for Terraform state
- Use regional filtering to reduce API calls

#### Memory Usage
- Monitor memory usage with large Terraform states
- Consider streaming parsing for very large configurations
- Use pagination for large instance lists

### Logging Configuration

Configure logging levels and formats:

```yaml
# firefly.yaml
logging:
  level: "debug"     # trace, debug, info, warn, error
  format: "json"     # json, text, structured
  output: "stdout"   # stdout, stderr, file path
  timestamp: true
  caller: true
```

## Design Decisions and Trade-offs

### Architecture Choices

1. **Modular Design**: The application is structured with clear separation of concerns:
   - AWS client abstraction for testability
   - Terraform parsing logic isolated from drift detection
   - Report generation as a separate concern
   - Dependency injection for loose coupling

2. **Interface-Based Design**: Heavy use of interfaces allows for:
   - Easy mocking in tests
   - Swappable implementations
   - Clear contracts between components

3. **Error Handling**: Custom error types provide:
   - Better error categorization
   - More informative error messages
   - Easier debugging and troubleshooting

### Trade-offs

1. **Complexity vs Flexibility**: The modular design adds some complexity but provides significant flexibility for testing and future enhancements.

2. **Performance vs Accuracy**: The tool prioritizes accuracy over speed, making multiple AWS API calls to ensure data consistency.

3. **Memory vs Processing**: Terraform state is loaded into memory for faster processing, which may not scale for very large state files.

## Ideas for Future Improvements

*   **Caching:** Implement a caching mechanism for parsed Terraform files to improve performance.
*   **More AWS Services:** Extend the tool to support drift detection for other AWS services like S3, RDS, and VPCs.
*   **Remediation:** Add a feature to automatically remediate detected drift by applying the necessary changes.
*   **Web UI:** Create a web-based dashboard to visualize drift detection results.
