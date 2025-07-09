package interfaces

import (
	"encoding/json"
	"time"
)

// EC2Instance represents an AWS EC2 instance configuration
type EC2Instance struct {
	// InstanceID is the unique identifier for the EC2 instance
	InstanceID string `json:"instance_id"`

	// InstanceType is the type of the EC2 instance (e.g., t3.micro, m5.large)
	InstanceType string `json:"instance_type"`

	// State is the current state of the instance (running, stopped, etc.)
	State string `json:"state"`

	// PublicIPAddress is the public IP address of the instance
	PublicIPAddress *string `json:"public_ip_address,omitempty"`

	// PrivateIPAddress is the private IP address of the instance
	PrivateIPAddress *string `json:"private_ip_address,omitempty"`

	// PublicDNSName is the public DNS name of the instance
	PublicDNSName *string `json:"public_dns_name,omitempty"`

	// PrivateDNSName is the private DNS name of the instance
	PrivateDNSName *string `json:"private_dns_name,omitempty"`

	// SecurityGroups is the list of security groups associated with the instance
	SecurityGroups []SecurityGroup `json:"security_groups"`

	// Tags is a map of tags associated with the instance
	Tags map[string]string `json:"tags"`

	// SubnetID is the ID of the subnet where the instance is located
	SubnetID *string `json:"subnet_id,omitempty"`

	// VPCID is the ID of the VPC where the instance is located
	VPCID *string `json:"vpc_id,omitempty"`

	// AvailabilityZone is the availability zone where the instance is located
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// LaunchTime is when the instance was launched
	LaunchTime *time.Time `json:"launch_time,omitempty"`

	// ImageID is the ID of the AMI used to launch the instance
	ImageID *string `json:"image_id,omitempty"`

	// KeyName is the name of the key pair used for the instance
	KeyName *string `json:"key_name,omitempty"`

	// Platform is the platform of the instance (e.g., windows)
	Platform *string `json:"platform,omitempty"`

	// Architecture is the architecture of the instance (e.g., x86_64, arm64)
	Architecture *string `json:"architecture,omitempty"`

	// Monitoring indicates if detailed monitoring is enabled
	Monitoring bool `json:"monitoring"`

	// EBSOptimized indicates if the instance is EBS-optimized
	EBSOptimized bool `json:"ebs_optimized"`
}

// SecurityGroup represents a security group associated with an EC2 instance
type SecurityGroup struct {
	// GroupID is the unique identifier of the security group
	GroupID string `json:"group_id"`

	// GroupName is the name of the security group
	GroupName string `json:"group_name"`
}

// S3Bucket represents an AWS S3 bucket configuration
type S3Bucket struct {
	// Name is the name of the S3 bucket
	Name string `json:"name"`

	// Region is the AWS region where the bucket is located
	Region string `json:"region"`

	// CreationDate is when the bucket was created
	CreationDate *time.Time `json:"creation_date,omitempty"`

	// Tags is a map of tags associated with the bucket
	Tags map[string]string `json:"tags"`

	// Versioning indicates if versioning is enabled
	Versioning bool `json:"versioning"`

	// Encryption indicates if encryption is enabled
	Encryption bool `json:"encryption"`

	// PublicAccessBlocked indicates if public access is blocked
	PublicAccessBlocked bool `json:"public_access_blocked"`
}

// TerraformConfig represents a Terraform configuration
type TerraformConfig struct {
	// ResourceID is the unique identifier for the resource
	ResourceID string `json:"resource_id"`

	// ResourceType is the type of the Terraform resource
	ResourceType string `json:"resource_type"`

	// ResourceName is the name of the Terraform resource
	ResourceName string `json:"resource_name"`

	// Attributes is a map of resource attributes
	Attributes map[string]interface{} `json:"attributes"`

	// Dependencies is a list of resource dependencies
	Dependencies []string `json:"dependencies,omitempty"`

	// Provider is the provider configuration
	Provider string `json:"provider,omitempty"`

	// Module is the module path if the resource is in a module
	Module string `json:"module,omitempty"`

	// TerraformVersion is the version of Terraform used
	TerraformVersion string `json:"terraform_version,omitempty"`

	// ProviderVersion is the version of the provider used
	ProviderVersion string `json:"provider_version,omitempty"`
}

// Clone creates a deep copy of the TerraformConfig
func (c *TerraformConfig) Clone() *TerraformConfig {
	if c == nil {
		return nil
	}
	newConfig := &TerraformConfig{
		ResourceID:       c.ResourceID,
		ResourceType:     c.ResourceType,
		ResourceName:     c.ResourceName,
		Provider:         c.Provider,
		TerraformVersion: c.TerraformVersion,
	}
	if c.Attributes != nil {
		newConfig.Attributes = make(map[string]interface{}, len(c.Attributes))
		for k, v := range c.Attributes {
			newConfig.Attributes[k] = v
		}
	}
	if c.Dependencies != nil {
		newConfig.Dependencies = make([]string, len(c.Dependencies))
		copy(newConfig.Dependencies, c.Dependencies)
	}
	return newConfig
}

// DriftResult represents the result of a drift detection for a single resource
type DriftResult struct {
	// ResourceID is the unique identifier of the resource
	ResourceID string `json:"resource_id"`

	// ResourceType is the type of the resource (e.g., aws_instance, aws_s3_bucket)
	ResourceType string `json:"resource_type"`

	// IsDrifted indicates whether the resource has drifted
	IsDrifted bool `json:"is_drifted"`

	// DetectionTime is when the drift was detected
	DetectionTime time.Time `json:"detection_time"`

	// DriftDetails contains the list of attributes that have drifted
	DriftDetails []*DriftDetail `json:"drift_details"`

	// Severity is the overall severity of the drift
	Severity SeverityLevel `json:"severity"`
}

// SeverityLevel defines the severity of a drift
type SeverityLevel string

const (
	// SeverityNone indicates no drift
	SeverityNone SeverityLevel = "none"
	// SeverityLow indicates a low-severity drift
	SeverityLow SeverityLevel = "low"
	// SeverityMedium indicates a medium-severity drift
	SeverityMedium SeverityLevel = "medium"
	// SeverityHigh indicates a high-severity drift
	SeverityHigh SeverityLevel = "high"
	// SeverityCritical indicates a critical-severity drift
	SeverityCritical SeverityLevel = "critical"
)

// String returns the string representation of the severity level
func (s SeverityLevel) String() string {
	return string(s)
}

// GetHighestSeverity returns the highest severity level among the drift details
func (dr *DriftResult) GetHighestSeverity() SeverityLevel {
	highest := SeverityNone
	highestOrder := getSeverityOrder(highest)
	for _, detail := range dr.DriftDetails {
		detailOrder := getSeverityOrder(detail.Severity)
		if detailOrder > highestOrder {
			highest = detail.Severity
			highestOrder = detailOrder
		}
	}
	return highest
}

// getSeverityOrder returns the numeric order of a severity level for comparison
func getSeverityOrder(severity SeverityLevel) int {
	switch severity {
	case SeverityLow:
		return 1
	case SeverityMedium:
		return 2
	case SeverityHigh:
		return 3
	case SeverityCritical:
		return 4
	default: // SeverityNone
		return 0
	}
}

// DriftDetail represents a specific drift detected in a resource
type DriftDetail struct {
	// Attribute is the name of the attribute that has drifted
	Attribute string `json:"attribute"`

	// ExpectedValue is the expected value from Terraform configuration
	ExpectedValue interface{} `json:"expected_value"`

	// ActualValue is the actual value from the cloud provider
	ActualValue interface{} `json:"actual_value"`

	// DriftType indicates the type of drift (added, removed, changed)
	DriftType string `json:"drift_type"`

	// Description provides more context about the drift
	Description string `json:"description,omitempty"`

	// Severity is the severity of the drift for this attribute
	Severity SeverityLevel `json:"severity"`
}

// DriftStatistics represents statistics about drift detection results
type DriftStatistics struct {
	TotalResources     int            `json:"total_resources"`
	ResourcesWithDrift int            `json:"resources_with_drift"`
	TotalDrifts        int            `json:"total_drifts"`
	SeverityBreakdown  map[string]int `json:"severity_breakdown"`
	AttributeBreakdown map[string]int `json:"attribute_breakdown"`
	DriftPercentage    float64        `json:"drift_percentage"`
}



// ToJSON converts the EC2Instance to JSON string
func (e *EC2Instance) ToJSON() (string, error) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetTag returns the value of a specific tag, or empty string if not found
func (e *EC2Instance) GetTag(key string) string {
	if e.Tags == nil {
		return ""
	}
	return e.Tags[key]
}

// HasTag checks if the instance has a specific tag
func (e *EC2Instance) HasTag(key string) bool {
	if e.Tags == nil {
		return false
	}
	_, exists := e.Tags[key]
	return exists
}

// IsRunning checks if the instance is in running state
func (e *EC2Instance) IsRunning() bool {
	return e.State == "running"
}

// IsStopped checks if the instance is in stopped state
func (e *EC2Instance) IsStopped() bool {
	return e.State == "stopped"
}

// ResourceState represents the state of a Terraform resource
type ResourceState struct {
	// ResourceAddress is the full address of the resource
	ResourceAddress string `json:"resource_address"`

	// ResourceType is the type of the resource
	ResourceType string `json:"resource_type"`

	// ResourceName is the name of the resource
	ResourceName string `json:"resource_name"`

	// Attributes contains the resource attributes
	Attributes map[string]interface{} `json:"attributes"`

	// Dependencies contains resource dependencies
	Dependencies []string `json:"dependencies,omitempty"`
}