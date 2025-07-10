# Contributing to Firefly

Thank you for your interest in contributing to Firefly! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Process](#contributing-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Pull Request Process](#pull-request-process)
- [Issue Reporting](#issue-reporting)
- [Community](#community)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone, regardless of age, body size, disability, ethnicity, gender identity and expression, level of experience, nationality, personal appearance, race, religion, or sexual identity and orientation.

### Our Standards

**Positive behavior includes:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints and experiences
- Gracefully accepting constructive criticism
- Focusing on what is best for the community
- Showing empathy towards other community members

**Unacceptable behavior includes:**
- Harassment, trolling, or discriminatory comments
- Personal attacks or political arguments
- Publishing private information without permission
- Any conduct that would be inappropriate in a professional setting

## Getting Started

### Prerequisites

- **Go**: Version 1.24.4 or later
- **Git**: For version control
- **AWS CLI**: For testing AWS integrations (optional)
- **Terraform**: For testing Terraform parsing (optional)
- **Make**: For build automation

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/firefly-task.git
   cd firefly-task
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/original-org/firefly-task.git
   ```

## Development Setup

### Environment Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Verify setup:**
   ```bash
   go test ./...
   ```

3. **Build the project:**
   ```bash
   make build
   # or
   go build -o firefly ./cmd/main.go
   ```

### Development Tools

**Recommended tools:**
- **IDE**: VS Code, GoLand, or Vim with Go plugins
- **Linting**: golangci-lint
- **Formatting**: gofmt, goimports
- **Testing**: go test, testify
- **Debugging**: delve (dlv)

**Install development tools:**
```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install goimports
go install golang.org/x/tools/cmd/goimports@latest

# Install delve debugger
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Project Structure

```
firefly-task/
├── cmd/                    # Application entry points
│   └── main.go            # Main CLI application
├── pkg/                   # Public packages
│   ├── app/              # Application logic
│   ├── container/        # Dependency injection
│   ├── errors/           # Error handling
│   └── logging/          # Logging utilities
├── aws/                  # AWS client implementations
├── drift/                # Drift detection logic
├── report/               # Report generation
├── terraform/            # Terraform parsing
├── config/               # Configuration files
├── docs/                 # Documentation
├── mocks/                # Mock implementations
└── tests/                # Integration tests
```

## Contributing Process

### 1. Choose an Issue

- Browse [open issues](https://github.com/Jesuloba-world/firefly-task/issues)
- Look for issues labeled `good first issue` or `help wanted`
- Comment on the issue to indicate you're working on it
- Wait for maintainer acknowledgment before starting work

### 2. Create a Branch

```bash
# Update your main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feature/your-feature-name
# or for bug fixes
git checkout -b fix/issue-description
```

### 3. Make Changes

- Write clean, well-documented code
- Follow the coding standards (see below)
- Add tests for new functionality
- Update documentation as needed
- Commit changes with clear messages

### 4. Test Your Changes

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run linting
golangci-lint run

# Format code
gofmt -s -w .
goimports -w .
```

### 5. Submit Pull Request

- Push your branch to your fork
- Create a pull request against the main branch
- Fill out the pull request template
- Wait for review and address feedback

## Coding Standards

### Go Style Guide

We follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html) guidelines.

### Key Principles

1. **Simplicity**: Write simple, readable code
2. **Consistency**: Follow existing patterns in the codebase
3. **Documentation**: Document public APIs and complex logic
4. **Error Handling**: Handle errors explicitly and appropriately
5. **Testing**: Write tests for all new functionality

### Code Formatting

```bash
# Format code
gofmt -s -w .

# Organize imports
goimports -w .

# Run linter
golangci-lint run
```

### Naming Conventions

- **Packages**: Short, lowercase, single word when possible
- **Functions**: CamelCase, descriptive names
- **Variables**: camelCase, meaningful names
- **Constants**: CamelCase or ALL_CAPS for package-level constants
- **Interfaces**: Often end with "-er" (e.g., `DriftDetector`)

### Documentation

```go
// Package drift provides infrastructure drift detection capabilities.
// It compares live AWS resources with Terraform configurations to identify
// discrepancies and generate detailed reports.
package drift

// DriftDetector defines the interface for detecting infrastructure drift.
// Implementations should compare live infrastructure state with desired
// configuration and return detailed drift information.
type DriftDetector interface {
    // DetectDrift compares the actual instance state with the expected
    // Terraform configuration and returns any detected differences.
    //
    // Parameters:
    //   ctx: Context for cancellation and timeouts
    //   instanceID: AWS EC2 instance identifier
    //   tfConfig: Parsed Terraform configuration
    //
    // Returns:
    //   DriftResult containing detected differences, or error if detection fails
    DetectDrift(ctx context.Context, instanceID string, tfConfig *TerraformConfig) (*DriftResult, error)
}
```

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
    return nil, fmt.Errorf("failed to detect drift for instance %s: %w", instanceID, err)
}

// Good: Use custom error types for specific cases
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}
```

## Testing Guidelines

### Test Structure

```go
func TestDriftDetector_DetectDrift(t *testing.T) {
    tests := []struct {
        name           string
        instanceID     string
        tfConfig       *TerraformConfig
        mockSetup      func(*mocks.MockEC2Client)
        expectedResult *DriftResult
        expectedError  string
    }{
        {
            name:       "successful drift detection",
            instanceID: "i-1234567890abcdef0",
            tfConfig:   &TerraformConfig{/* ... */},
            mockSetup: func(m *mocks.MockEC2Client) {
                m.On("DescribeInstances", mock.Anything, mock.Anything).Return(/* ... */)
            },
            expectedResult: &DriftResult{/* ... */},
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **End-to-End Tests**: Test complete workflows
4. **Mock Tests**: Use mocks for external dependencies

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestDriftDetector_DetectDrift ./drift

# Run integration tests
go test -tags=integration ./...

# Run benchmarks
go test -bench=. ./...
```

### Test Best Practices

- **Arrange, Act, Assert**: Structure tests clearly
- **Table-driven tests**: Use for multiple similar test cases
- **Mocking**: Mock external dependencies
- **Test data**: Use realistic test data
- **Cleanup**: Clean up resources in tests
- **Parallel tests**: Use `t.Parallel()` when safe

## Documentation

### Types of Documentation

1. **Code Documentation**: Go doc comments
2. **API Documentation**: Interface and function documentation
3. **User Documentation**: README, guides, examples
4. **Architecture Documentation**: Design decisions, patterns

### Documentation Standards

- **Go doc comments**: Start with the name being documented
- **Examples**: Provide runnable examples
- **Markdown**: Use for user-facing documentation
- **Diagrams**: Use ASCII art or mermaid for simple diagrams

### Updating Documentation

When making changes:

1. **Update code comments** for API changes
2. **Update README** for new features or setup changes
3. **Update examples** if usage patterns change
4. **Update architecture docs** for design changes

## Pull Request Process

### Before Submitting

- [ ] Code follows style guidelines
- [ ] Tests pass locally
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests pass
```

### Review Process

1. **Automated checks**: CI/CD pipeline runs tests and linting
2. **Peer review**: At least one maintainer reviews the code
3. **Feedback**: Address review comments
4. **Approval**: Maintainer approves the changes
5. **Merge**: Changes are merged to main branch

### Review Criteria

- **Functionality**: Does the code work as intended?
- **Quality**: Is the code well-written and maintainable?
- **Testing**: Are there adequate tests?
- **Documentation**: Is the code properly documented?
- **Performance**: Are there any performance concerns?
- **Security**: Are there any security implications?

## Issue Reporting

### Bug Reports

When reporting bugs, include:

1. **Environment**: OS, Go version, Firefly version
2. **Steps to reproduce**: Clear, minimal reproduction steps
3. **Expected behavior**: What should happen
4. **Actual behavior**: What actually happens
5. **Logs**: Relevant log output or error messages
6. **Configuration**: Relevant configuration (sanitized)

### Feature Requests

When requesting features, include:

1. **Use case**: What problem does this solve?
2. **Proposed solution**: How should it work?
3. **Alternatives**: Other approaches considered
4. **Impact**: Who would benefit from this feature?

### Issue Templates

Use the provided issue templates for consistency:

- **Bug Report Template**
- **Feature Request Template**
- **Documentation Issue Template**
- **Question Template**

## Community

### Communication Channels

- **GitHub Issues**: Bug reports, feature requests
- **GitHub Discussions**: General questions, ideas
- **Pull Requests**: Code contributions
- **Documentation**: In-code and markdown documentation

### Getting Help

1. **Check documentation**: Start with README and docs/
2. **Search issues**: Look for existing solutions
3. **Ask questions**: Use GitHub Discussions
4. **Report bugs**: Use GitHub Issues

### Recognition

We recognize contributors through:

- **Contributors file**: Listed in CONTRIBUTORS.md
- **Release notes**: Mentioned in changelog
- **GitHub**: Contributor statistics and badges

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- **MAJOR**: Breaking changes
- **MINOR**: New features (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

### Release Steps

1. **Update version**: Update version in code and documentation
2. **Update changelog**: Move unreleased changes to new version
3. **Create tag**: Tag the release commit
4. **Build artifacts**: Create release binaries
5. **GitHub release**: Create release with changelog
6. **Announce**: Notify community of new release

## License

By contributing to Firefly, you agree that your contributions will be licensed under the same license as the project (MIT License).

---

## Questions?

If you have questions about contributing, please:

1. Check this document first
2. Search existing issues and discussions
3. Create a new discussion or issue
4. Reach out to maintainers

Thank you for contributing to Firefly! 🚀