package terraform

import (
	"testing"
)

func TestTerraformConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *TerraformConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &TerraformConfig{
				ResourceID:   "aws_instance.web",
				InstanceType: "t3.micro",
				AMI:          "ami-12345678",
			},
			wantErr: false,
		},
		{
			name: "missing resource_id",
			config: &TerraformConfig{
				InstanceType: "t3.micro",
				AMI:          "ami-12345678",
			},
			wantErr: true,
			errMsg:  "resource_id is required",
		},
		{
			name: "missing instance_type",
			config: &TerraformConfig{
				ResourceID: "aws_instance.web",
				AMI:        "ami-12345678",
			},
			wantErr: true,
			errMsg:  "instance_type is required",
		},
		{
			name: "missing ami",
			config: &TerraformConfig{
				ResourceID:   "aws_instance.web",
				InstanceType: "t3.micro",
			},
			wantErr: true,
			errMsg:  "ami is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("TerraformConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("TerraformConfig.Validate() error = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestTerraformConfig_GetResourceType(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "aws_instance",
			resourceID: "aws_instance.web",
			want:       "aws_instance",
		},
		{
			name:       "aws_security_group",
			resourceID: "aws_security_group.web_sg",
			want:       "aws_security_group",
		},
		{
			name:       "empty resource_id",
			resourceID: "",
			want:       "",
		},
		{
			name:       "no dot separator",
			resourceID: "aws_instance",
			want:       "aws_instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &TerraformConfig{ResourceID: tt.resourceID}
			if got := config.GetResourceType(); got != tt.want {
				t.Errorf("TerraformConfig.GetResourceType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerraformConfig_GetResourceName(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "aws_instance.web",
			resourceID: "aws_instance.web",
			want:       "web",
		},
		{
			name:       "aws_security_group.web_sg",
			resourceID: "aws_security_group.web_sg",
			want:       "web_sg",
		},
		{
			name:       "no dot separator",
			resourceID: "aws_instance",
			want:       "",
		},
		{
			name:       "multiple dots",
			resourceID: "aws_instance.web.server",
			want:       "web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &TerraformConfig{ResourceID: tt.resourceID}
			if got := config.GetResourceName(); got != tt.want {
				t.Errorf("TerraformConfig.GetResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerraformConfig_IsEC2Instance(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		want       bool
	}{
		{
			name:       "aws_instance",
			resourceID: "aws_instance.web",
			want:       true,
		},
		{
			name:       "aws_security_group",
			resourceID: "aws_security_group.web_sg",
			want:       false,
		},
		{
			name:       "empty",
			resourceID: "",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &TerraformConfig{ResourceID: tt.resourceID}
			if got := config.IsEC2Instance(); got != tt.want {
				t.Errorf("TerraformConfig.IsEC2Instance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTerraformConfig_TagOperations(t *testing.T) {
	config := &TerraformConfig{
		ResourceID:   "aws_instance.web",
		InstanceType: "t3.micro",
		AMI:          "ami-12345678",
	}

	// Test HasTag on empty tags
	if config.HasTag("Name") {
		t.Error("HasTag should return false for empty tags")
	}

	// Test GetTag on empty tags
	if got := config.GetTag("Name"); got != "" {
		t.Errorf("GetTag should return empty string for empty tags, got %v", got)
	}

	// Test SetTag
	config.SetTag("Name", "web-server")
	config.SetTag("Environment", "production")

	// Test HasTag after setting
	if !config.HasTag("Name") {
		t.Error("HasTag should return true after setting tag")
	}

	// Test GetTag after setting
	if got := config.GetTag("Name"); got != "web-server" {
		t.Errorf("GetTag should return 'web-server', got %v", got)
	}

	// Test GetTag for non-existent tag
	if got := config.GetTag("NonExistent"); got != "" {
		t.Errorf("GetTag should return empty string for non-existent tag, got %v", got)
	}
}

func TestTerraformConfig_MergeSecurityGroups(t *testing.T) {
	config := &TerraformConfig{
		ResourceID:        "aws_instance.web",
		InstanceType:      "t3.micro",
		AMI:               "ami-12345678",
		SecurityGroups:    []string{"sg-12345", "sg-67890"},
		SecurityGroupRefs: []string{"${aws_security_group.web.id}", "${aws_security_group.db.id}"},
	}

	merged := config.MergeSecurityGroups()
	expected := []string{"sg-12345", "sg-67890", "${aws_security_group.web.id}", "${aws_security_group.db.id}"}

	if len(merged) != len(expected) {
		t.Errorf("MergeSecurityGroups() length = %v, want %v", len(merged), len(expected))
	}

	for i, sg := range expected {
		if merged[i] != sg {
			t.Errorf("MergeSecurityGroups()[%d] = %v, want %v", i, merged[i], sg)
		}
	}
}

func TestTerraformConfig_Clone(t *testing.T) {
	original := &TerraformConfig{
		ResourceID:   "aws_instance.web",
		InstanceType: "t3.micro",
		AMI:          "ami-12345678",
		Tags: map[string]string{
			"Name":        "web-server",
			"Environment": "production",
		},
		SecurityGroups:    []string{"sg-12345", "sg-67890"},
		SecurityGroupRefs: []string{"${aws_security_group.web.id}"},
		RootBlockDevice: &BlockDevice{
			DeviceName: "/dev/sda1",
			VolumeSize: 20,
			VolumeType: "gp3",
		},
		EBSBlockDevices: []*BlockDevice{
			{
				DeviceName: "/dev/sdf",
				VolumeSize: 100,
				VolumeType: "gp3",
			},
		},
	}

	clone := original.Clone()

	// Test that clone is not the same object
	if clone == original {
		t.Error("Clone should return a different object")
	}

	// Test that basic fields are copied
	if clone.ResourceID != original.ResourceID {
		t.Errorf("Clone ResourceID = %v, want %v", clone.ResourceID, original.ResourceID)
	}

	// Test that tags are deep copied
	if &clone.Tags == &original.Tags {
		t.Error("Clone should deep copy tags map")
	}
	if clone.Tags["Name"] != original.Tags["Name"] {
		t.Error("Clone should copy tag values")
	}

	// Test that security groups are deep copied
	if &clone.SecurityGroups[0] == &original.SecurityGroups[0] {
		t.Error("Clone should deep copy security groups slice")
	}
	if clone.SecurityGroups[0] != original.SecurityGroups[0] {
		t.Error("Clone should copy security group values")
	}

	// Test that block devices are deep copied
	if clone.RootBlockDevice == original.RootBlockDevice {
		t.Error("Clone should deep copy root block device")
	}
	if clone.RootBlockDevice.DeviceName != original.RootBlockDevice.DeviceName {
		t.Error("Clone should copy root block device values")
	}

	if clone.EBSBlockDevices[0] == original.EBSBlockDevices[0] {
		t.Error("Clone should deep copy EBS block devices")
	}
	if clone.EBSBlockDevices[0].DeviceName != original.EBSBlockDevices[0].DeviceName {
		t.Error("Clone should copy EBS block device values")
	}

	// Test that modifying clone doesn't affect original
	clone.Tags["Name"] = "modified"
	if original.Tags["Name"] == "modified" {
		t.Error("Modifying clone should not affect original")
	}
}

func TestTerraformConfig_String(t *testing.T) {
	config := &TerraformConfig{
		ResourceID:   "aws_instance.web",
		InstanceType: "t3.micro",
		AMI:          "ami-12345678",
		InstanceID:   "i-1234567890abcdef0",
	}

	expected := "TerraformConfig{ResourceID: aws_instance.web, InstanceType: t3.micro, AMI: ami-12345678, InstanceID: i-1234567890abcdef0}"
	if got := config.String(); got != expected {
		t.Errorf("TerraformConfig.String() = %v, want %v", got, expected)
	}
}

func TestBlockDevice(t *testing.T) {
	bd := &BlockDevice{
		DeviceName:          "/dev/sda1",
		VolumeType:          "gp3",
		VolumeSize:          20,
		IOPS:                3000,
		Throughput:          125,
		DeleteOnTermination: boolPtr(true),
		Encrypted:           boolPtr(true),
		KMSKeyID:            "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
	}

	if bd.DeviceName != "/dev/sda1" {
		t.Errorf("BlockDevice.DeviceName = %v, want %v", bd.DeviceName, "/dev/sda1")
	}
	if bd.VolumeType != "gp3" {
		t.Errorf("BlockDevice.VolumeType = %v, want %v", bd.VolumeType, "gp3")
	}
	if bd.VolumeSize != 20 {
		t.Errorf("BlockDevice.VolumeSize = %v, want %v", bd.VolumeSize, 20)
	}
	if *bd.DeleteOnTermination != true {
		t.Errorf("BlockDevice.DeleteOnTermination = %v, want %v", *bd.DeleteOnTermination, true)
	}
}

func TestResourceMapping(t *testing.T) {
	rm := &ResourceMapping{
		TerraformID:  "aws_instance.web",
		AWSID:        "i-1234567890abcdef0",
		ResourceType: "aws_instance",
	}

	if rm.TerraformID != "aws_instance.web" {
		t.Errorf("ResourceMapping.TerraformID = %v, want %v", rm.TerraformID, "aws_instance.web")
	}
	if rm.AWSID != "i-1234567890abcdef0" {
		t.Errorf("ResourceMapping.AWSID = %v, want %v", rm.AWSID, "i-1234567890abcdef0")
	}
	if rm.ResourceType != "aws_instance" {
		t.Errorf("ResourceMapping.ResourceType = %v, want %v", rm.ResourceType, "aws_instance")
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}