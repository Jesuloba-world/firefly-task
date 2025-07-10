# Firefly Documentation

Welcome to the comprehensive documentation for Firefly, an advanced infrastructure drift detection tool for AWS EC2 instances and Terraform configurations.

## Documentation Overview

This documentation is organized into several sections to help you get started quickly and dive deep into advanced features:

### 📚 Core Documentation

- **[Architecture Guide](architecture.md)** - Detailed system architecture, design patterns, and component interactions
- **[API Reference](api-reference.md)** - Complete API documentation with interfaces, data structures, and examples
- **[Examples & Tutorials](examples.md)** - Practical examples, configuration samples, and step-by-step tutorials

### 🚀 Quick Start

For immediate setup and basic usage, see the main [README.md](../README.md) in the project root.

### 📖 What's in This Directory

| File | Description |
|------|-------------|
| `architecture.md` | System architecture, design decisions, and component relationships |
| `api-reference.md` | Complete API documentation with interfaces and data structures |
| `examples.md` | Practical examples, tutorials, and configuration samples |
| `README.md` | This overview document |

## Key Features

### 🔍 Drift Detection
- **Real-time Analysis**: Compare live AWS infrastructure with Terraform configurations
- **Attribute-level Comparison**: Detailed analysis of instance properties, security groups, tags, and more
- **Severity Assessment**: Automatic categorization of drift severity (low, medium, high, critical)
- **Custom Rules**: Configurable severity rules and ignore patterns

### 📊 Reporting
- **Multiple Formats**: JSON, YAML, HTML, Markdown, CSV, and XML output
- **Custom Templates**: Extensible templating system for custom report formats
- **Batch Processing**: Efficient processing of large instance sets
- **Historical Tracking**: Support for trend analysis and historical comparisons

### 🔧 Integration
- **CI/CD Ready**: Built for integration with GitHub Actions, Jenkins, and other CI/CD systems
- **Monitoring**: Prometheus metrics and Grafana dashboard support
- **Alerting**: Integration with Slack, email, and other notification systems
- **Caching**: Intelligent caching for improved performance

### ⚙️ Configuration
- **Flexible Config**: YAML configuration files with environment variable support
- **AWS Integration**: Multiple authentication methods (CLI, environment, IAM roles)
- **Parallel Processing**: Configurable concurrency for large-scale operations
- **Logging**: Structured logging with multiple output formats

## Getting Started

### 1. Installation

Choose your preferred installation method:

```bash
# Using Go
go install github.com/Jesuloba-world/firefly-task@latest

# Using pre-built binary
curl -L -o firefly https://github.com/your-org/firefly/releases/latest/download/firefly-linux-amd64
chmod +x firefly

# Using package manager (if available)
brew install firefly  # macOS
apt install firefly   # Ubuntu/Debian
```

### 2. Basic Configuration

Create a `firefly.yaml` configuration file:

```yaml
aws:
  region: "us-west-2"
  profile: "default"

logging:
  level: "info"
  format: "json"

reporting:
  format: "json"
  output_dir: "./reports"

drift:
  ignore_tags: ["LastModified", "CreatedBy"]
  strict_mode: false
```

### 3. First Drift Check

```bash
# Check a single instance
firefly check --instance-id i-0123456789abcdef0 --tf-path ./terraform/

# Check multiple instances by tags
firefly check --tags "Environment=production" --tf-path ./terraform/

# Generate HTML report
firefly check --instance-id i-0123456789abcdef0 --tf-path ./terraform/ --format html --output report.html
```

## Common Use Cases

### 🏢 Enterprise Environments
- **Large-scale Monitoring**: Monitor hundreds or thousands of instances across multiple regions
- **Compliance Reporting**: Generate compliance reports for auditing and governance
- **Change Management**: Track infrastructure changes and ensure they match approved configurations
- **Cost Optimization**: Identify instances that have drifted from cost-optimized configurations

### 🔄 CI/CD Integration
- **Pre-deployment Checks**: Validate infrastructure state before deployments
- **Post-deployment Verification**: Ensure deployments match expected configurations
- **Scheduled Monitoring**: Regular drift detection as part of operational procedures
- **Automated Remediation**: Trigger remediation workflows when drift is detected

### 🛠️ Development Teams
- **Environment Consistency**: Ensure development, staging, and production environments match
- **Configuration Validation**: Validate Terraform configurations against live infrastructure
- **Debugging**: Identify configuration discrepancies during troubleshooting
- **Documentation**: Generate up-to-date infrastructure documentation

## Architecture Overview

Firefly follows a modular, interface-driven architecture:

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Layer (Cobra)                        │
├─────────────────────────────────────────────────────────────────┤
│                     Application Layer                          │
├─────────────────────────────────────────────────────────────────┤
│                      Business Logic Layer                      │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Drift Detection │  │ Report Generator│  │ Config Manager  │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
├─────────────────────────────────────────────────────────────────┤
│                    Infrastructure Layer                        │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   AWS Client    │  │ Terraform Parser│  │   File System   │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

For detailed architecture information, see [architecture.md](architecture.md).

## API Reference

Firefly provides comprehensive APIs for integration and extension:

### Core Interfaces
- **DriftDetector**: Main drift detection interface
- **ReportGenerator**: Report generation and formatting
- **EC2Client**: AWS EC2 operations
- **TerraformParser**: Terraform configuration parsing

For complete API documentation, see [api-reference.md](api-reference.md).

## Examples and Tutorials

Practical examples for common scenarios:

### Quick Examples
- **Basic Usage**: Simple drift detection workflows
- **CI/CD Integration**: GitHub Actions and Jenkins examples
- **Batch Processing**: Large-scale environment monitoring
- **Custom Reports**: Creating custom report templates

For detailed examples and tutorials, see [examples.md](examples.md).

## Configuration Reference

### Configuration File Structure

```yaml
# Complete configuration example
aws:
  region: string              # AWS region
  profile: string             # AWS profile name
  timeout: duration           # API timeout
  max_retries: int           # Maximum retry attempts

logging:
  level: string              # Log level (debug, info, warn, error)
  format: string             # Log format (json, text, structured)
  output: string             # Output destination
  timestamp: bool            # Include timestamps
  caller: bool               # Include caller information

reporting:
  format: string             # Report format (json, yaml, html, etc.)
  output_dir: string         # Output directory
  include_metadata: bool     # Include metadata in reports
  template_path: string      # Custom template path

drift:
  ignore_tags: []string      # Tags to ignore during comparison
  ignore_attributes: []string # Attributes to ignore
  strict_mode: bool          # Enable strict comparison mode
  timeout: duration          # Drift detection timeout
  severity_rules: map        # Custom severity rules

parallel:
  max_workers: int           # Maximum parallel workers
  batch_size: int            # Batch processing size
  rate_limit: string         # API rate limiting

cache:
  enabled: bool              # Enable caching
  ttl: duration              # Cache time-to-live
  directory: string          # Cache directory
```

### Environment Variables

All configuration options can be overridden with environment variables:

```bash
# AWS Configuration
export FIREFLY_AWS_REGION="us-west-2"
export FIREFLY_AWS_PROFILE="production"

# Logging Configuration
export FIREFLY_LOG_LEVEL="debug"
export FIREFLY_LOG_FORMAT="json"

# Reporting Configuration
export FIREFLY_REPORT_FORMAT="html"
export FIREFLY_OUTPUT_DIR="./reports"

# Drift Detection Configuration
export FIREFLY_STRICT_MODE="true"
export FIREFLY_IGNORE_TAGS="LastModified,CreatedBy"
```

## Troubleshooting

### Common Issues

#### AWS Authentication
```bash
# Verify AWS credentials
aws sts get-caller-identity

# Check AWS profile
aws configure list-profiles
```

#### Terraform Configuration
```bash
# Validate Terraform files
terraform validate

# Check Terraform formatting
terraform fmt -check
```

#### Debug Mode
```bash
# Enable debug logging
firefly check --instance-id i-xxx --tf-path ./terraform/ --log-level debug
```

### Performance Optimization

- **Parallel Processing**: Use `--parallel` flag for large batches
- **Caching**: Enable caching for repeated operations
- **Regional Filtering**: Limit scope to specific regions
- **Attribute Filtering**: Focus on specific attributes of interest

## Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/Jesuloba-world/firefly-task.git
cd firefly-task

# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o firefly ./cmd/main.go
```

### Documentation Updates

When updating documentation:

1. **Architecture changes**: Update `architecture.md`
2. **API changes**: Update `api-reference.md`
3. **New examples**: Add to `examples.md`
4. **Configuration changes**: Update this README

### Testing

```bash
# Run unit tests
go test ./...

# Run integration tests
go test -tags=integration ./...

# Run with coverage
go test -cover ./...
```

## Support and Community

### Getting Help

- **Documentation**: Start with this documentation
- **Issues**: Report bugs and feature requests on GitHub
- **Discussions**: Join community discussions
- **Examples**: Check the examples directory for common patterns

### Reporting Issues

When reporting issues, please include:

1. **Version**: Firefly version (`firefly version`)
2. **Configuration**: Relevant configuration (sanitized)
3. **Command**: Exact command that failed
4. **Output**: Error messages and logs
5. **Environment**: OS, Go version, AWS region

### Feature Requests

For feature requests, please describe:

1. **Use Case**: What you're trying to accomplish
2. **Current Limitation**: What's not possible today
3. **Proposed Solution**: How you envision it working
4. **Alternatives**: Other approaches you've considered

## License

Firefly is released under the [MIT License](../LICENSE).

## Changelog

See [CHANGELOG.md](../CHANGELOG.md) for version history and release notes.