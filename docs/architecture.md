# Firefly Architecture Documentation

## Overview

Firefly is designed with a modular, interface-driven architecture that promotes testability, maintainability, and extensibility. The system follows clean architecture principles with clear separation of concerns.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLI Layer (Cobra)                        │
├─────────────────────────────────────────────────────────────────┤
│                     Application Layer                          │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   Commands      │  │   Handlers      │  │   Validators    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
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

## Core Components

### 1. CLI Layer
- **Framework**: Cobra CLI framework
- **Responsibility**: Command parsing, flag handling, user interaction
- **Key Files**: `cmd/`

### 2. Application Layer
- **Commands**: Define CLI commands and their behavior
- **Handlers**: Process user requests and coordinate business logic
- **Validators**: Input validation and sanitization

### 3. Business Logic Layer

#### Drift Detection Engine
- **Purpose**: Core drift detection algorithms
- **Components**:
  - `DriftDetector`: Main detection interface
  - `DriftComparator`: Attribute comparison logic
  - `DriftAnalyzer`: Analysis and severity assessment

#### Report Generator
- **Purpose**: Generate reports in various formats
- **Components**:
  - `ReportGenerator`: Main report generation interface
  - `ReportWriter`: Output handling
  - `ReportFormatter`: Format-specific rendering
  - `ReportFilter`: Data filtering and transformation

#### Configuration Manager
- **Purpose**: Handle application configuration
- **Features**: YAML/JSON config files, environment variables, defaults

### 4. Infrastructure Layer

#### AWS Client
- **Purpose**: AWS API interactions
- **Components**:
  - `EC2Client`: EC2 instance operations
  - `S3Client`: S3 operations for state storage
  - `AWSClientFactory`: Client creation and configuration

#### Terraform Parser
- **Purpose**: Parse and analyze Terraform configurations
- **Components**:
  - `TerraformParser`: Configuration parsing
  - `TerraformExecutor`: Terraform command execution
  - `TerraformStateManager`: State file management

## Data Flow

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    User     │───▶│     CLI     │───▶│  Commands   │
└─────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Reports   │◀───│  Generator  │◀───│  Handlers   │
└─────────────┘    └─────────────┘    └─────────────┘
                                              │
                                              ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    AWS      │───▶│    Drift    │◀───│ Terraform   │
│   Client    │    │  Detection  │    │   Parser    │
└─────────────┘    └─────────────┘    └─────────────┘
```

### Detailed Flow

1. **Input Processing**:
   - User executes CLI command
   - Cobra parses arguments and flags
   - Command handler validates input

2. **Data Gathering**:
   - Terraform parser reads configuration files
   - AWS client fetches current instance state
   - Data is normalized and validated

3. **Drift Detection**:
   - Expected state (from Terraform) vs Actual state (from AWS)
   - Attribute-by-attribute comparison
   - Severity assessment and categorization

4. **Report Generation**:
   - Results aggregation and analysis
   - Format-specific rendering (JSON, YAML, HTML, etc.)
   - Output to file or stdout

## Design Patterns

### 1. Dependency Injection
- **Implementation**: Constructor injection via interfaces
- **Benefits**: Testability, loose coupling, flexibility
- **Example**: Container pattern in `pkg/container/`

### 2. Strategy Pattern
- **Usage**: Report formatting, drift comparison algorithms
- **Benefits**: Pluggable implementations, extensibility

### 3. Factory Pattern
- **Usage**: AWS client creation, report generator instantiation
- **Benefits**: Centralized object creation, configuration management

### 4. Observer Pattern
- **Usage**: Logging, progress reporting
- **Benefits**: Decoupled event handling, extensible monitoring

## Interface Design

### Core Interfaces

```go
// Primary business logic interfaces
type DriftDetector interface {
    DetectDrift(ctx context.Context, expected, actual interface{}) (*DriftResult, error)
}

type ReportGenerator interface {
    GenerateReport(ctx context.Context, results []DriftResult) (*Report, error)
}

// Infrastructure interfaces
type EC2Client interface {
    DescribeInstances(ctx context.Context, instanceIDs []string) ([]EC2Instance, error)
}

type TerraformParser interface {
    ParseConfiguration(path string) (*TerraformConfig, error)
}
```

### Interface Benefits
- **Testability**: Easy mocking for unit tests
- **Flexibility**: Swappable implementations
- **Maintainability**: Clear contracts and boundaries

## Error Handling Strategy

### Error Types
```go
type ErrorType string

const (
    ValidationError   ErrorType = "validation"
    AWSError         ErrorType = "aws"
    TerraformError   ErrorType = "terraform"
    ConfigError      ErrorType = "config"
    InternalError    ErrorType = "internal"
)
```

### Error Propagation
- **Structured Errors**: Custom error types with context
- **Error Wrapping**: Preserve error chains for debugging
- **Graceful Degradation**: Continue processing when possible

## Concurrency Model

### Parallel Processing
- **Instance Processing**: Concurrent drift detection for multiple instances
- **API Calls**: Parallel AWS API requests with rate limiting
- **Report Generation**: Concurrent formatting for different output types

### Synchronization
- **Context Cancellation**: Graceful shutdown and timeout handling
- **Worker Pools**: Controlled concurrency with configurable limits
- **Rate Limiting**: AWS API throttling compliance

## Security Considerations

### Credential Management
- **AWS Credentials**: Support for multiple authentication methods
- **Secure Storage**: No credential persistence in application
- **Least Privilege**: Minimal required AWS permissions

### Data Protection
- **Sensitive Data**: Careful handling of instance metadata
- **Logging**: Sanitized logs without sensitive information
- **Temporary Files**: Secure cleanup of temporary data

## Performance Optimizations

### Caching Strategy
- **Terraform State**: Cache parsed configurations
- **AWS Metadata**: Cache instance data for batch operations
- **Report Templates**: Pre-compiled report templates

### Resource Management
- **Memory Usage**: Streaming for large datasets
- **Connection Pooling**: Reuse AWS client connections
- **Garbage Collection**: Efficient memory cleanup

## Testing Strategy

### Unit Testing
- **Interface Mocking**: Comprehensive mock implementations
- **Test Coverage**: Target >90% code coverage
- **Test Data**: Realistic test fixtures and scenarios

### Integration Testing
- **AWS Integration**: Real AWS API testing (optional)
- **End-to-End**: Complete workflow testing
- **Performance Testing**: Load and stress testing

## Extensibility Points

### Adding New Cloud Providers
1. Implement cloud-specific client interfaces
2. Add provider-specific configuration
3. Extend drift detection for provider-specific attributes

### Custom Report Formats
1. Implement `ReportFormatter` interface
2. Register new format in factory
3. Add CLI flag for new format

### Additional Terraform Resources
1. Extend parser for new resource types
2. Add resource-specific drift detection logic
3. Update report templates for new attributes

## Deployment Considerations

### Binary Distribution
- **Single Binary**: Self-contained executable
- **Cross-Platform**: Support for multiple OS/architectures
- **Version Management**: Semantic versioning and release automation

### Configuration Management
- **Default Configuration**: Sensible defaults for common use cases
- **Environment-Specific**: Support for different deployment environments
- **Validation**: Configuration validation and error reporting