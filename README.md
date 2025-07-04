# Firefly - AWS EC2 Configuration Drift Detection Tool

## Project Overview

Firefly is a CLI tool for detecting configuration drift in AWS EC2 instances. It compares the current state of EC2 instances with their expected configuration defined in Terraform files and reports any discrepancies. This helps ensure that your infrastructure remains compliant with its intended state, reducing the risk of unexpected issues and security vulnerabilities.

## Setup and Installation Instructions

1.  **Prerequisites:**
    *   Go (version 1.18 or higher)
    *   AWS account with configured credentials

2.  **Installation:**
    ```bash
    git clone https://github.com/your-repo/firefly-task.git
    go build -o firefly
    ```

## Usage Examples

### Check a single instance

```bash
./firefly check --instance-id i-0123456789abcdef0 --tf-path /path/to/your/terraform.tf
```

### Check multiple instances in batch

Create a file named `instances.txt` with one instance ID per line:

```
i-0123456789abcdef0
i-fedcba9876543210f
```

Then run:

```bash
./firefly batch --input-file instances.txt --tf-path /path/to/your/terraform.tf
```

### Check a specific attribute

```bash
./firefly attribute --instance-id i-0123456789abcdef0 --tf-path /path/to/your/terraform.tf --attribute instance_type
```

## Design Decisions and Trade-offs

*   **Cobra for CLI:** We use the `cobra` library to create a powerful and user-friendly CLI interface. It simplifies command parsing, flag handling, and help text generation.
*   **Structured Logging:** We use a custom logging package to provide structured, leveled logging. This makes it easier to debug issues and monitor the application's behavior.
*   **Modular Design:** The application is divided into several packages (`aws`, `terraform`, `drift`, `report`), each with a specific responsibility. This promotes code reusability and maintainability.
*   **Trade-offs:** The current implementation parses the entire Terraform file for each command execution. For large Terraform configurations, this could be optimized by caching the parsed results.

## Ideas for Future Improvements

*   **Caching:** Implement a caching mechanism for parsed Terraform files to improve performance.
*   **More AWS Services:** Extend the tool to support drift detection for other AWS services like S3, RDS, and VPCs.
*   **Remediation:** Add a feature to automatically remediate detected drift by applying the necessary changes.
*   **Web UI:** Create a web-based dashboard to visualize drift detection results.
