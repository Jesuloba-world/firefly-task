# Firefly Examples and Tutorials

## Quick Start Tutorial

### Prerequisites

1. **AWS CLI configured**:
   ```bash
   aws configure
   # or
   export AWS_ACCESS_KEY_ID="your-access-key"
   export AWS_SECRET_ACCESS_KEY="your-secret-key"
   export AWS_DEFAULT_REGION="us-west-2"
   ```

2. **Terraform files ready**:
   ```bash
   # Ensure you have .tf files in your project directory
   ls *.tf
   # main.tf  variables.tf  outputs.tf
   ```

3. **Firefly installed**:
   ```bash
   # Download and install firefly
   go install github.com/Jesuloba-world/firefly-task@latest
   # or use pre-built binary
   ```

### Basic Workflow

#### Step 1: Check Single Instance

```bash
# Basic drift check for one instance
firefly check --instance-id i-0123456789abcdef0 --tf-path ./terraform/
```

**Expected Output:**
```json
{
  "summary": {
    "total_instances": 1,
    "drift_detected": false,
    "drift_count": 0
  },
  "instances": {
    "i-0123456789abcdef0": {
      "drift_detected": false,
      "differences": []
    }
  }
}
```

#### Step 2: Check Multiple Instances

```bash
# Check multiple instances by ID
firefly check --instance-ids i-0123456789abcdef0,i-fedcba9876543210f --tf-path ./terraform/

# Check all instances with specific tags
firefly check --tags "Environment=production,Team=backend" --tf-path ./terraform/
```

#### Step 3: Generate Detailed Report

```bash
# Generate HTML report
firefly check --instance-id i-0123456789abcdef0 --tf-path ./terraform/ --format html --output drift-report.html

# Generate YAML report with metadata
firefly check --instance-id i-0123456789abcdef0 --tf-path ./terraform/ --format yaml --include-metadata --output report.yaml
```

## Configuration Examples

### Basic Configuration File

Create `firefly.yaml` in your project root:

```yaml
# Basic firefly.yaml configuration
aws:
  region: "us-west-2"
  profile: "default"
  timeout: "30s"
  max_retries: 3

logging:
  level: "info"
  format: "json"
  output: "stdout"
  timestamp: true

reporting:
  format: "json"
  output_dir: "./reports"
  include_metadata: true

drift:
  ignore_tags: ["LastModified", "CreatedBy", "ManagedBy"]
  strict_mode: false
  timeout: "60s"
```

### Advanced Configuration

```yaml
# Advanced firefly.yaml configuration
aws:
  region: "us-west-2"
  profile: "production"
  timeout: "45s"
  max_retries: 5
  endpoint_url: ""  # For LocalStack or custom endpoints

logging:
  level: "debug"
  format: "structured"
  output: "/var/log/firefly/firefly.log"
  timestamp: true
  caller: true
  max_size: "100MB"
  max_backups: 5

reporting:
  format: "html"
  output_dir: "./reports"
  include_metadata: true
  template_path: "./templates/custom-report.html"
  compress: true
  retention_days: 30

drift:
  ignore_tags: 
    - "LastModified"
    - "CreatedBy"
    - "ManagedBy"
    - "aws:*"  # Ignore all AWS-managed tags
  ignore_attributes:
    - "tags.LastModified"
    - "launch_time"
  strict_mode: true
  timeout: "120s"
  severity_rules:
    instance_type: "high"
    security_groups: "critical"
    subnet_id: "medium"
    tags: "low"

parallel:
  max_workers: 10
  batch_size: 50
  rate_limit: "10/s"

cache:
  enabled: true
  ttl: "5m"
  directory: "/tmp/firefly-cache"
```

### Environment-Specific Configurations

#### Development Environment

```yaml
# firefly-dev.yaml
aws:
  region: "us-east-1"
  profile: "dev"
  timeout: "15s"

logging:
  level: "debug"
  format: "text"
  output: "stdout"

reporting:
  format: "json"
  output_dir: "./dev-reports"

drift:
  ignore_tags: ["*"]  # Ignore all tags in dev
  strict_mode: false
```

#### Production Environment

```yaml
# firefly-prod.yaml
aws:
  region: "us-west-2"
  profile: "production"
  timeout: "60s"
  max_retries: 5

logging:
  level: "warn"
  format: "json"
  output: "/var/log/firefly/production.log"

reporting:
  format: "html"
  output_dir: "/var/reports/firefly"
  include_metadata: true

drift:
  strict_mode: true
  timeout: "300s"
```

## Usage Scenarios

### Scenario 1: CI/CD Pipeline Integration

#### GitHub Actions Example

```yaml
# .github/workflows/drift-detection.yml
name: Infrastructure Drift Detection

on:
  schedule:
    - cron: '0 8 * * *'  # Daily at 8 AM
  workflow_dispatch:

jobs:
  drift-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-west-2
      
      - name: Download Firefly
        run: |
          curl -L -o firefly https://github.com/your-org/firefly/releases/latest/download/firefly-linux-amd64
          chmod +x firefly
      
      - name: Run Drift Detection
        run: |
          ./firefly check \
            --tags "Environment=production" \
            --tf-path ./terraform/ \
            --format json \
            --output drift-report.json \
            --fail-on-drift
      
      - name: Upload Report
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: drift-report
          path: drift-report.json
      
      - name: Notify on Drift
        if: failure()
        uses: 8398a7/action-slack@v3
        with:
          status: failure
          text: "Infrastructure drift detected! Check the report for details."
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK }}
```

#### Jenkins Pipeline Example

```groovy
// Jenkinsfile
pipeline {
    agent any
    
    triggers {
        cron('H 8 * * *')  // Daily at 8 AM
    }
    
    environment {
        AWS_DEFAULT_REGION = 'us-west-2'
        AWS_CREDENTIALS = credentials('aws-credentials')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }
        
        stage('Download Firefly') {
            steps {
                sh '''
                    curl -L -o firefly https://github.com/your-org/firefly/releases/latest/download/firefly-linux-amd64
                    chmod +x firefly
                '''
            }
        }
        
        stage('Drift Detection') {
            steps {
                script {
                    def exitCode = sh(
                        script: '''
                            ./firefly check \
                                --tags "Environment=production" \
                                --tf-path ./terraform/ \
                                --format json \
                                --output drift-report.json \
                                --fail-on-drift
                        ''',
                        returnStatus: true
                    )
                    
                    if (exitCode != 0) {
                        currentBuild.result = 'UNSTABLE'
                        error "Infrastructure drift detected!"
                    }
                }
            }
        }
    }
    
    post {
        always {
            archiveArtifacts artifacts: 'drift-report.json', fingerprint: true
        }
        unstable {
            emailext (
                subject: "Infrastructure Drift Detected - ${env.JOB_NAME} #${env.BUILD_NUMBER}",
                body: "Infrastructure drift has been detected. Please check the attached report.",
                to: "devops@company.com",
                attachmentsPattern: "drift-report.json"
            )
        }
    }
}
```

### Scenario 2: Batch Processing Large Environments

#### Create Instance List

```bash
# Generate list of production instances
aws ec2 describe-instances \
  --filters "Name=tag:Environment,Values=production" \
  --query "Reservations[].Instances[].InstanceId" \
  --output text | tr '\t' '\n' > production-instances.txt
```

#### Batch Processing Script

```bash
#!/bin/bash
# batch-drift-check.sh

set -e

INSTANCE_FILE="production-instances.txt"
TF_PATH="./terraform/production"
REPORT_DIR="./reports/$(date +%Y-%m-%d)"
PARALLEL_WORKERS=5

# Create report directory
mkdir -p "$REPORT_DIR"

# Run batch drift detection
firefly batch \
  --input-file "$INSTANCE_FILE" \
  --tf-path "$TF_PATH" \
  --parallel $PARALLEL_WORKERS \
  --format html \
  --output "$REPORT_DIR/batch-report.html" \
  --include-metadata \
  --log-level info

# Generate summary
firefly batch \
  --input-file "$INSTANCE_FILE" \
  --tf-path "$TF_PATH" \
  --format json \
  --output "$REPORT_DIR/summary.json" \
  --summary-only

echo "Batch processing complete. Reports saved to $REPORT_DIR"
```

### Scenario 3: Monitoring and Alerting

#### Prometheus Metrics Integration

```bash
#!/bin/bash
# drift-metrics.sh - Export metrics to Prometheus

RESULT=$(firefly check --tags "Environment=production" --tf-path ./terraform/ --format json --quiet)
DRIFT_COUNT=$(echo "$RESULT" | jq '.summary.drift_count')
TOTAL_INSTANCES=$(echo "$RESULT" | jq '.summary.total_instances')

# Export metrics
cat << EOF > /var/lib/prometheus/node-exporter/firefly.prom
# HELP firefly_drift_count Number of instances with detected drift
# TYPE firefly_drift_count gauge
firefly_drift_count $DRIFT_COUNT

# HELP firefly_total_instances Total number of instances checked
# TYPE firefly_total_instances gauge
firefly_total_instances $TOTAL_INSTANCES

# HELP firefly_drift_ratio Ratio of instances with drift
# TYPE firefly_drift_ratio gauge
firefly_drift_ratio $(echo "scale=4; $DRIFT_COUNT / $TOTAL_INSTANCES" | bc)
EOF
```

#### Grafana Dashboard Query Examples

```promql
# Drift count over time
firefly_drift_count

# Drift ratio percentage
firefly_drift_ratio * 100

# Alert when drift ratio exceeds 10%
firefly_drift_ratio > 0.1
```

### Scenario 4: Custom Report Templates

#### HTML Report Template

```html
<!-- templates/custom-report.html -->
<!DOCTYPE html>
<html>
<head>
    <title>Infrastructure Drift Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .summary { margin: 20px 0; }
        .drift-item { border: 1px solid #ddd; margin: 10px 0; padding: 15px; }
        .severity-high { border-left: 5px solid #ff4444; }
        .severity-medium { border-left: 5px solid #ffaa00; }
        .severity-low { border-left: 5px solid #44ff44; }
        .no-drift { color: #008000; }
        .has-drift { color: #ff0000; }
    </style>
</head>
<body>
    <div class="header">
        <h1>Infrastructure Drift Report</h1>
        <p>Generated: {{.Metadata.GeneratedAt}}</p>
        <p>Environment: {{.Metadata.Environment}}</p>
    </div>
    
    <div class="summary">
        <h2>Summary</h2>
        <ul>
            <li>Total Instances: {{.Summary.TotalInstances}}</li>
            <li>Drift Detected: {{if .Summary.DriftDetected}}<span class="has-drift">Yes</span>{{else}}<span class="no-drift">No</span>{{end}}</li>
            <li>Instances with Drift: {{.Summary.DriftCount}}</li>
            <li>Execution Time: {{.Summary.ExecutionTime}}</li>
        </ul>
    </div>
    
    <div class="results">
        <h2>Detailed Results</h2>
        {{range .Results}}
        <div class="drift-item {{if .DriftDetected}}severity-{{.Severity}}{{end}}">
            <h3>Instance: {{.InstanceID}}</h3>
            {{if .DriftDetected}}
                <p><strong>Drift Detected:</strong> <span class="has-drift">Yes</span></p>
                <p><strong>Severity:</strong> {{.Severity}}</p>
                <h4>Differences:</h4>
                <ul>
                {{range .Differences}}
                    <li>
                        <strong>{{.Attribute}}:</strong> 
                        Expected: <code>{{.Expected}}</code>, 
                        Actual: <code>{{.Actual}}</code>
                        ({{.Severity}})
                    </li>
                {{end}}
                </ul>
            {{else}}
                <p><strong>Drift Detected:</strong> <span class="no-drift">No</span></p>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>
```

#### Markdown Report Template

```markdown
<!-- templates/report.md -->
# Infrastructure Drift Report

**Generated:** {{.Metadata.GeneratedAt}}  
**Environment:** {{.Metadata.Environment}}  
**Region:** {{.Metadata.Region}}

## Summary

| Metric | Value |
|--------|-------|
| Total Instances | {{.Summary.TotalInstances}} |
| Drift Detected | {{if .Summary.DriftDetected}}✅ Yes{{else}}❌ No{{end}} |
| Instances with Drift | {{.Summary.DriftCount}} |
| Execution Time | {{.Summary.ExecutionTime}} |

{{if .Summary.DriftDetected}}
## ⚠️ Drift Detection Results

{{range .Results}}
{{if .DriftDetected}}
### Instance: `{{.InstanceID}}`

**Severity:** {{.Severity | upper}}  
**Resource:** {{.ResourceName}}

#### Differences:

{{range .Differences}}
- **{{.Attribute}}**
  - Expected: `{{.Expected}}`
  - Actual: `{{.Actual}}`
  - Severity: {{.Severity}}
  {{if .Description}}- Description: {{.Description}}{{end}}

{{end}}

{{if .Recommendations}}
#### Recommendations:

{{range .Recommendations}}
- {{.Description}}
{{end}}
{{end}}

---

{{end}}
{{end}}
{{else}}
## ✅ No Drift Detected

All instances match their Terraform configuration.
{{end}}

## Instance Details

{{range .Results}}
- **{{.InstanceID}}**: {{if .DriftDetected}}❌ Drift{{else}}✅ No Drift{{end}}
{{end}}
```

### Scenario 5: Advanced Filtering and Analysis

#### Filter by Specific Attributes

```bash
# Check only instance type and security groups
firefly attribute \
  --instance-id i-0123456789abcdef0 \
  --tf-path ./terraform/ \
  --attributes instance_type,security_groups \
  --format yaml
```

#### Ignore Specific Attributes

```bash
# Ignore tags and launch time (common for dynamic attributes)
firefly check \
  --instance-id i-0123456789abcdef0 \
  --tf-path ./terraform/ \
  --ignore-attributes "tags.LastModified,tags.CreatedBy,launch_time" \
  --format json
```

#### Custom Severity Rules

```yaml
# firefly.yaml with custom severity rules
drift:
  severity_rules:
    instance_type: "critical"      # Instance type changes are critical
    security_groups: "high"        # Security group changes are high priority
    subnet_id: "high"              # Subnet changes are high priority
    vpc_id: "critical"             # VPC changes are critical
    tags: "low"                    # Tag changes are low priority
    iam_instance_profile: "high"   # IAM profile changes are high priority
```

## Troubleshooting Examples

### Debug Mode

```bash
# Enable debug logging
firefly check \
  --instance-id i-0123456789abcdef0 \
  --tf-path ./terraform/ \
  --log-level debug \
  --log-output debug.log

# Check the debug log
tail -f debug.log
```

### Validate Configuration

```bash
# Validate Terraform configuration
terraform validate ./terraform/

# Test AWS connectivity
aws sts get-caller-identity

# Test instance access
aws ec2 describe-instances --instance-ids i-0123456789abcdef0
```

### Performance Optimization

```bash
# Use parallel processing for large batches
firefly batch \
  --input-file large-instance-list.txt \
  --tf-path ./terraform/ \
  --parallel 10 \
  --timeout 300s

# Enable caching for repeated runs
export FIREFLY_CACHE_ENABLED=true
export FIREFLY_CACHE_TTL=300s
```

## Best Practices

### 1. Configuration Management

- Use environment-specific configuration files
- Store sensitive data in environment variables
- Version control your configuration files
- Use configuration validation

### 2. CI/CD Integration

- Run drift detection on schedule
- Fail builds on critical drift
- Archive reports for historical analysis
- Set up notifications for drift detection

### 3. Performance Optimization

- Use batch processing for large environments
- Enable parallel processing
- Implement caching for repeated runs
- Use regional filtering to reduce API calls

### 4. Security

- Use IAM roles instead of access keys when possible
- Implement least privilege access
- Rotate credentials regularly
- Audit access logs

### 5. Monitoring

- Set up alerts for drift detection
- Monitor execution time and performance
- Track drift trends over time
- Implement automated remediation where appropriate