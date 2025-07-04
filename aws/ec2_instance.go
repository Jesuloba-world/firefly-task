package aws

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

// ToJSON converts the EC2Instance to JSON string
func (e *EC2Instance) ToJSON() (string, error) {
	data, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FromJSON creates an EC2Instance from JSON string
func FromJSON(jsonStr string) (*EC2Instance, error) {
	var instance EC2Instance
	err := json.Unmarshal([]byte(jsonStr), &instance)
	if err != nil {
		return nil, err
	}
	return &instance, nil
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
	return e.State == InstanceStateRunning
}

// IsStopped checks if the instance is in stopped state
func (e *EC2Instance) IsStopped() bool {
	return e.State == InstanceStateStopped
}
