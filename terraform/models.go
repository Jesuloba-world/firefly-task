package terraform

import (
	"fmt"
	"strings"
)

// TerraformConfig represents the expected EC2 instance configuration
// extracted from Terraform files (state or HCL)
type TerraformConfig struct {
	// Resource identification
	ResourceID   string `json:"resource_id"`   // Terraform resource ID (e.g., "aws_instance.web")
	InstanceID   string `json:"instance_id"`   // AWS instance ID (e.g., "i-1234567890abcdef0")
	ResourceName string `json:"resource_name"` // Resource name from Terraform

	// EC2 Instance Configuration
	InstanceType     string            `json:"instance_type"`
	AMI              string            `json:"ami"`
	KeyName          string            `json:"key_name,omitempty"`
	SubnetID         string            `json:"subnet_id,omitempty"`
	VPCID            string            `json:"vpc_id,omitempty"`
	AvailabilityZone string            `json:"availability_zone,omitempty"`
	PrivateIP        string            `json:"private_ip,omitempty"`
	PublicIP         string            `json:"public_ip,omitempty"`
	EBSOptimized     *bool             `json:"ebs_optimized,omitempty"`
	Monitoring       *bool             `json:"monitoring,omitempty"`
	Tags             map[string]string `json:"tags,omitempty"`

	// Security Configuration
	SecurityGroups    []string `json:"security_groups,omitempty"`     // Security group IDs
	SecurityGroupRefs []string `json:"security_group_refs,omitempty"` // Terraform references

	// Storage Configuration
	RootBlockDevice *BlockDevice   `json:"root_block_device,omitempty"`
	EBSBlockDevices []*BlockDevice `json:"ebs_block_devices,omitempty"`

	// Network Configuration
	AssociatePublicIPAddress *bool `json:"associate_public_ip_address,omitempty"`
	SourceDestCheck          *bool `json:"source_dest_check,omitempty"`

	// Metadata
	TerraformVersion string `json:"terraform_version,omitempty"`
	ProviderVersion  string `json:"provider_version,omitempty"`
}

// BlockDevice represents EBS block device configuration
type BlockDevice struct {
	DeviceName          string            `json:"device_name"`
	VolumeType          string            `json:"volume_type,omitempty"`
	VolumeSize          int               `json:"volume_size,omitempty"`
	IOPS                int               `json:"iops,omitempty"`
	Throughput          int               `json:"throughput,omitempty"`
	Encrypted           *bool             `json:"encrypted,omitempty"`
	KMSKeyID            string            `json:"kms_key_id,omitempty"`
	DeleteOnTermination *bool             `json:"delete_on_termination,omitempty"`
	SnapshotID          string            `json:"snapshot_id,omitempty"`
	Tags                map[string]string `json:"tags,omitempty"`
}

// EC2InstanceConfig represents EC2 instance configuration extracted from Terraform
type EC2InstanceConfig struct {
	InstanceType      string            `json:"instance_type"`
	AMI               string            `json:"ami"`
	SubnetID          string            `json:"subnet_id,omitempty"`
	VPCSecurityGroups []string          `json:"vpc_security_group_ids,omitempty"`
	Tags              map[string]string `json:"tags,omitempty"`
	KeyName           string            `json:"key_name,omitempty"`
	UserData          string            `json:"user_data,omitempty"`
	ResourceName      string            `json:"resource_name"`
}

// ResourceMapping represents the mapping between Terraform resources and AWS resources
type ResourceMapping struct {
	TerraformID  string `json:"terraform_id"`  // e.g., "aws_instance.web"
	AWSID        string `json:"aws_id"`        // e.g., "i-1234567890abcdef0"
	ResourceType string `json:"resource_type"` // e.g., "aws_instance"
}

// Validate checks if the TerraformConfig has required fields
func (tc *TerraformConfig) Validate() error {
	if tc.ResourceID == "" {
		return fmt.Errorf("resource_id is required")
	}
	if tc.InstanceType == "" {
		return fmt.Errorf("instance_type is required")
	}
	if tc.AMI == "" {
		return fmt.Errorf("ami is required")
	}
	return nil
}

// GetResourceType extracts the resource type from the resource ID
// e.g., "aws_instance.web" -> "aws_instance"
func (tc *TerraformConfig) GetResourceType() string {
	parts := strings.Split(tc.ResourceID, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

// GetResourceName extracts the resource name from the resource ID
// e.g., "aws_instance.web" -> "web"
func (tc *TerraformConfig) GetResourceName() string {
	parts := strings.Split(tc.ResourceID, ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// IsEC2Instance checks if this configuration represents an EC2 instance
func (tc *TerraformConfig) IsEC2Instance() bool {
	return tc.GetResourceType() == "aws_instance"
}

// HasTag checks if the configuration has a specific tag
func (tc *TerraformConfig) HasTag(key string) bool {
	if tc.Tags == nil {
		return false
	}
	_, exists := tc.Tags[key]
	return exists
}

// GetTag returns the value of a specific tag
func (tc *TerraformConfig) GetTag(key string) string {
	if tc.Tags == nil {
		return ""
	}
	return tc.Tags[key]
}

// SetTag sets a tag value
func (tc *TerraformConfig) SetTag(key, value string) {
	if tc.Tags == nil {
		tc.Tags = make(map[string]string)
	}
	tc.Tags[key] = value
}

// MergeSecurityGroups combines security group IDs and references
func (tc *TerraformConfig) MergeSecurityGroups() []string {
	allSGs := make([]string, 0, len(tc.SecurityGroups)+len(tc.SecurityGroupRefs))
	allSGs = append(allSGs, tc.SecurityGroups...)
	allSGs = append(allSGs, tc.SecurityGroupRefs...)
	return allSGs
}

// Clone creates a deep copy of the TerraformConfig
func (tc *TerraformConfig) Clone() *TerraformConfig {
	clone := *tc

	// Deep copy tags
	if tc.Tags != nil {
		clone.Tags = make(map[string]string, len(tc.Tags))
		for k, v := range tc.Tags {
			clone.Tags[k] = v
		}
	}

	// Deep copy security groups
	if tc.SecurityGroups != nil {
		clone.SecurityGroups = make([]string, len(tc.SecurityGroups))
		copy(clone.SecurityGroups, tc.SecurityGroups)
	}

	if tc.SecurityGroupRefs != nil {
		clone.SecurityGroupRefs = make([]string, len(tc.SecurityGroupRefs))
		copy(clone.SecurityGroupRefs, tc.SecurityGroupRefs)
	}

	// Deep copy block devices
	if tc.RootBlockDevice != nil {
		rootClone := *tc.RootBlockDevice
		clone.RootBlockDevice = &rootClone
	}

	if tc.EBSBlockDevices != nil {
		clone.EBSBlockDevices = make([]*BlockDevice, len(tc.EBSBlockDevices))
		for i, bd := range tc.EBSBlockDevices {
			if bd != nil {
				bdClone := *bd
				clone.EBSBlockDevices[i] = &bdClone
			}
		}
	}

	return &clone
}

// String returns a string representation of the TerraformConfig
func (tc *TerraformConfig) String() string {
	return fmt.Sprintf("TerraformConfig{ResourceID: %s, InstanceType: %s, AMI: %s, InstanceID: %s}",
		tc.ResourceID, tc.InstanceType, tc.AMI, tc.InstanceID)
}
