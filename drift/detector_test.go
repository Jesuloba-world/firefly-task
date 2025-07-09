package drift

import (
	"testing"

	"firefly-task/aws"
	"firefly-task/pkg/interfaces"
	"firefly-task/terraform"
)

func TestNewDriftDetector(t *testing.T) {
	config := DefaultDetectionConfig()
	detector := NewDriftDetector(config)

	if detector == nil {
		t.Fatal("NewDriftDetector returned nil")
	}

	gotConfig := detector.GetConfig()
	if gotConfig.MaxConcurrency != config.MaxConcurrency {
		t.Errorf("Expected MaxConcurrency %d, got %d", config.MaxConcurrency, gotConfig.MaxConcurrency)
	}
}

func TestDefaultDetectionConfig(t *testing.T) {
	config := DefaultDetectionConfig()

	if config.MaxConcurrency <= 0 {
		t.Error("MaxConcurrency should be positive")
	}

	if config.Timeout <= 0 {
		t.Error("Timeout should be positive")
	}

	if len(config.AttributeConfigs) == 0 {
		t.Error("AttributeConfigs should not be empty")
	}

	// Check some critical attribute configurations
	if config.AttributeConfigs["instance_id"].ComparisonType != ExactMatch {
		t.Error("instance_id should use ExactMatch comparison")
	}

	if config.AttributeConfigs["security_groups"].ComparisonType != ArrayUnordered {
		t.Error("security_groups should use ArrayUnordered comparison")
	}

	if config.AttributeConfigs["tags"].ComparisonType != MapComparison {
		t.Error("tags should use MapComparison")
	}
}

func TestDetectDrift_NilInputs(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	tests := []struct {
		name            string
		awsResource     interface{}
		terraformConfig interface{}
		wantError       bool
	}{
		{"both nil", nil, nil, true},
		{"aws nil", nil, &terraform.TerraformConfig{}, true},
		{"terraform nil", &aws.EC2Instance{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := detector.DetectDrift(tt.awsResource, tt.terraformConfig)
			if (err != nil) != tt.wantError {
				t.Errorf("DetectDrift() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestDetectDrift_IdenticalResources(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	publicIP := "203.0.113.12"
	privateIP := "10.0.1.55"
	subnetID := "subnet-12345678"
	vpcID := "vpc-12345678"
	availabilityZone := "us-west-2a"
	keyName := "my-key"
	imageID := "ami-0abcdef1234567890"

	awsInstance := &aws.EC2Instance{
		InstanceID:       "i-1234567890abcdef0",
		InstanceType:     "t3.micro",
		ImageID:          &imageID,
		State:            "running",
		PublicIPAddress:  &publicIP,
		PrivateIPAddress: &privateIP,
		SecurityGroups: []aws.SecurityGroup{
			{GroupID: "sg-12345678", GroupName: "test-sg"},
		},
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "dev",
		},
		SubnetID:         &subnetID,
		VPCID:            &vpcID,
		AvailabilityZone: &availabilityZone,
		KeyName:          &keyName,
		Monitoring:       true,
		EBSOptimized:     false,
	}

	terraformConfig := &terraform.TerraformConfig{
		ResourceID:        "aws_instance.test",
		InstanceID:        "i-1234567890abcdef0",
		InstanceType:      "t3.micro",
		AMI:               "ami-0abcdef1234567890",
		SecurityGroupRefs: []string{"sg-12345678"},
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "dev",
		},
		SubnetID:         subnetID,
		VPCID:            vpcID,
		AvailabilityZone: availabilityZone,
		KeyName:          keyName,
		Monitoring:       &[]bool{true}[0],
		EBSOptimized:     &[]bool{false}[0],
		PublicIP:         publicIP,
		PrivateIP:        privateIP,
	}

	result, err := detector.DetectDrift(awsInstance, terraformConfig)
	if err != nil {
		t.Fatalf("DetectDrift() error = %v", err)
	}

	if result.IsDrifted {
		t.Error("Expected no drift for identical resources")
		for _, diff := range result.DriftDetails {
			t.Logf("Difference: %s - AWS: %v, Terraform: %v", diff.Attribute, diff.ActualValue, diff.ExpectedValue)
		}
	}

	if result.Severity != interfaces.SeverityNone {
		t.Errorf("Expected SeverityNone, got %v", result.Severity)
	}

	if len(result.DriftDetails) != 0 {
		t.Errorf("Expected 0 differences, got %d", len(result.DriftDetails))
	}
}

func TestDetectDrift_WithDifferences(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	subnetID2 := "subnet-12345678"
	vpcID2 := "vpc-12345678"
	availabilityZone2 := "us-west-2a"
	imageID2 := "ami-0abcdef1234567890"

	awsInstance := &aws.EC2Instance{
		InstanceID:   "i-1234567890abcdef0",
		InstanceType: "t3.small", // Different from Terraform
		ImageID:      &imageID2,
		State:        "running",
		SecurityGroups: []aws.SecurityGroup{
			{GroupID: "sg-12345678", GroupName: "test-sg1"},
			{GroupID: "sg-87654321", GroupName: "test-sg2"}, // Extra security group
		},
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "prod",   // Different value
			"Owner":       "team-a", // Extra tag
		},
		SubnetID:         &subnetID2,
		VPCID:            &vpcID2,
		AvailabilityZone: &availabilityZone2,
		Monitoring:       false, // Different from Terraform
	}

	terraformConfig := &terraform.TerraformConfig{
		ResourceID:        "aws_instance.test",
		InstanceID:        "i-1234567890abcdef0",
		InstanceType:      "t3.micro", // Different from AWS
		AMI:               "ami-0abcdef1234567890",
		SecurityGroupRefs: []string{"sg-12345678"}, // Missing security group
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "dev", // Different value
			// Missing Owner tag
		},
		Monitoring: &[]bool{true}[0], // Different from AWS (false)
	}

	result, err := detector.DetectDrift(awsInstance, terraformConfig)
	if err != nil {
		t.Fatalf("DetectDrift() error = %v", err)
	}

	if !result.IsDrifted {
		t.Error("Expected drift to be detected")
	}

	if result.Severity == interfaces.SeverityNone {
		t.Errorf("Expected non-zero severity, got %v. Drift details: %d", result.Severity, len(result.DriftDetails))
	}

	if len(result.DriftDetails) == 0 {
		t.Error("Expected differences to be found")
	}

	// Check for specific differences
	foundInstanceType := false
	foundSecurityGroups := false
	foundTags := false
	foundMonitoring := false

	for _, diff := range result.DriftDetails {
		switch diff.Attribute {
		case "instance_type":
			foundInstanceType = true
			if diff.Severity != interfaces.SeverityCritical {
				t.Errorf("Expected instance_type to have critical severity, got %v", diff.Severity)
			}
		case "security_groups":
			foundSecurityGroups = true
			if diff.Severity != interfaces.SeverityCritical {
				t.Errorf("Expected security_groups to have critical severity, got %v", diff.Severity)
			}
		case "tags":
			foundTags = true
			if diff.Severity != interfaces.SeverityMedium {
				t.Errorf("Expected tags to have medium severity, got %v", diff.Severity)
			}
		case "monitoring":
			foundMonitoring = true
			if diff.Severity != interfaces.SeverityHigh {
				t.Errorf("Expected monitoring to have high severity, got %v", diff.Severity)
			}
		}
	}

	if !foundInstanceType {
		t.Error("Expected instance_type difference to be detected")
	}
	if !foundSecurityGroups {
		t.Error("Expected security_groups difference to be detected")
	}
	if !foundTags {
		t.Error("Expected tags difference to be detected")
	}
	if !foundMonitoring {
		t.Error("Expected monitoring difference to be detected")
	}
}

func TestDetectDrift_IgnoredAttributes(t *testing.T) {
	config := DefaultDetectionConfig()
	config.IgnoredAttributes = append(config.IgnoredAttributes, "instance_type", "ebs_optimized", "monitoring")
	detector := NewDriftDetector(config)

	imageID3 := "ami-0abcdef1234567890"

	awsInstance := &aws.EC2Instance{
		InstanceID:   "i-1234567890abcdef0",
		InstanceType: "t3.small", // Different but should be ignored
		ImageID:      &imageID3,
		State:        "running",
	}

	terraformConfig := &terraform.TerraformConfig{
		ResourceID:   "aws_instance.test",
		InstanceID:   "i-1234567890abcdef0",
		InstanceType: "t3.micro", // Different but should be ignored
		AMI:          "ami-0abcdef1234567890",
	}

	result, err := detector.DetectDrift(awsInstance, terraformConfig)
	if err != nil {
		t.Fatalf("DetectDrift() error = %v", err)
	}

	// Should not detect drift because instance_type is ignored
	if result.IsDrifted {
		t.Error("Expected no drift when instance_type is ignored")
	}

	// Verify instance_type difference is not in results
	for _, diff := range result.DriftDetails {
		if diff.Attribute == "instance_type" {
			t.Error("instance_type should be ignored but was found in differences")
		}
	}
}

func TestDetectDriftBatch(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	// Create test resource pairs
	resourcePairs := []ResourcePair{
		{
			Index: 0,
			AWSResource: &aws.EC2Instance{
				InstanceID:   "i-1111111111111111",
				InstanceType: "t3.micro",
				ImageID:      &[]string{"ami-0abcdef1234567890"}[0],
			},
			TerraformConfig: &terraform.TerraformConfig{
				ResourceID:   "aws_instance.test1",
				InstanceID:   "i-1111111111111111",
				InstanceType: "t3.micro",
				AMI:          "ami-0abcdef1234567890",
				Monitoring:   &[]bool{false}[0],
				EBSOptimized: &[]bool{false}[0],
			},
		},
		{
			Index: 1,
			AWSResource: &aws.EC2Instance{
				InstanceID:   "i-2222222222222222",
				InstanceType: "t3.small", // Different
				ImageID:      &[]string{"ami-0abcdef1234567890"}[0],
			},
			TerraformConfig: &terraform.TerraformConfig{
				ResourceID:   "aws_instance.test2",
				InstanceID:   "i-2222222222222222",
				InstanceType: "t3.micro", // Different
				AMI:          "ami-0abcdef1234567890",
				Monitoring:   &[]bool{false}[0],
				EBSOptimized: &[]bool{false}[0],
			},
		},
	}

	results, err := detector.DetectDriftBatch(resourcePairs)
	if err != nil {
		t.Fatalf("DetectDriftBatch() error = %v", err)
	}

	if len(results) != len(resourcePairs) {
		t.Errorf("Expected %d results, got %d", len(resourcePairs), len(results))
	}

	// First resource should have no drift
	if results[0].IsDrifted {
		t.Error("First resource should have no drift")
	}

	// Second resource should have drift
	if !results[1].IsDrifted {
		t.Error("Second resource should have drift")
	}
}

func TestDetectDriftBatch_WithErrors(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	// Create resource pairs with one invalid pair
	resourcePairs := []ResourcePair{
		{
			Index: 0,
			AWSResource: &aws.EC2Instance{
				InstanceID: "i-1111111111111111",
			},
			TerraformConfig: &terraform.TerraformConfig{
				ResourceID: "aws_instance.test1",
			},
		},
		{
			Index:       1,
			AWSResource: nil, // This will cause an error
			TerraformConfig: &terraform.TerraformConfig{
				ResourceID: "aws_instance.test2",
			},
		},
	}

	results, err := detector.DetectDriftBatch(resourcePairs)
	if err == nil {
		t.Error("Expected error due to nil AWS resource")
	}

	// Should still return results array with valid results
	if len(results) != len(resourcePairs) {
		t.Errorf("Expected %d results, got %d", len(resourcePairs), len(results))
	}

	// First result should be valid
	if results[0] == nil {
		t.Error("First result should not be nil")
	}

	// Second result should be nil due to error
	if results[1] != nil {
		t.Error("Second result should be nil due to error")
	}
}

func TestUpdateConfig(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	newConfig := DefaultDetectionConfig()
	newConfig.MaxConcurrency = 20
	newConfig.StrictMode = true

	detector.UpdateConfig(newConfig)

	updatedConfig := detector.GetConfig()
	if updatedConfig.MaxConcurrency != 20 {
		t.Errorf("Expected MaxConcurrency 20, got %d", updatedConfig.MaxConcurrency)
	}

	if !updatedConfig.StrictMode {
		t.Error("Expected StrictMode to be true")
	}
}

func TestResourceToMap_EC2Instance(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	instance := &aws.EC2Instance{
		InstanceID:       "i-1234567890abcdef0",
		InstanceType:     "t3.micro",
		ImageID:          &[]string{"ami-0abcdef1234567890"}[0],
		State:            "running",
		PublicIPAddress:  &[]string{"203.0.113.12"}[0],
		PrivateIPAddress: &[]string{"10.0.1.55"}[0],
		Tags: map[string]string{
			"Name": "test-instance",
		},
	}

	m, err := detector.resourceToMap(instance)
	if err != nil {
		t.Fatalf("resourceToMap() error = %v", err)
	}

	if m["instance_id"] != "i-1234567890abcdef0" {
		t.Errorf("Expected instance_id 'i-1234567890abcdef0', got %v", m["instance_id"])
	}

	if m["instance_type"] != "t3.micro" {
		t.Errorf("Expected instance_type 't3.micro', got %v", m["instance_type"])
	}

	if m["ami"] != "ami-0abcdef1234567890" {
		t.Errorf("Expected ami 'ami-0abcdef1234567890', got %v", m["ami"])
	}

	tags, ok := m["tags"].(map[string]string)
	if !ok {
		t.Error("Expected tags to be map[string]string")
	} else if tags["Name"] != "test-instance" {
		t.Errorf("Expected tag Name 'test-instance', got %v", tags["Name"])
	}
}

func TestResourceToMap_TerraformConfig(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	config := &terraform.TerraformConfig{
		ResourceID:   "aws_instance.test",
		InstanceID:   "i-1234567890abcdef0",
		InstanceType: "t3.micro",
		AMI:          "ami-0abcdef1234567890",
		Tags: map[string]string{
			"Name": "test-instance",
		},
		SecurityGroupRefs: []string{"sg-12345678"},
	}

	m, err := detector.resourceToMap(config)
	if err != nil {
		t.Fatalf("resourceToMap() error = %v", err)
	}

	// resource_id is not included in the map for drift comparison
	if _, exists := m["resource_id"]; exists {
		t.Error("resource_id should not be included in drift comparison")
	}

	if m["instance_id"] != "i-1234567890abcdef0" {
		t.Errorf("Expected instance_id 'i-1234567890abcdef0', got %v", m["instance_id"])
	}

	securityGroups, ok := m["security_groups"].([]string)
	if !ok {
		t.Error("Expected security_groups to be []string")
	} else if len(securityGroups) != 1 || securityGroups[0] != "sg-12345678" {
		t.Errorf("Expected security_groups ['sg-12345678'], got %v", securityGroups)
	}
}

// TestToSnakeCase removed - toSnakeCase method is now internal

func TestExtractResourceID(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	tests := []struct {
		name     string
		resource interface{}
		expected string
	}{
		{
			"EC2Instance",
			&aws.EC2Instance{InstanceID: "i-1234567890abcdef0"},
			"i-1234567890abcdef0",
		},
		{
			"TerraformConfig",
			&terraform.TerraformConfig{ResourceID: "aws_instance.test"},
			"aws_instance.test",
		},
		{
			"EC2InstanceConfig",
			&terraform.EC2InstanceConfig{InstanceType: "t3.micro"},
			"",
		},
		{
			"Unknown",
			"some string",
			"unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractResourceID(tt.resource)
			if result != tt.expected {
				t.Errorf("extractResourceID() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestExtractResourceType(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	tests := []struct {
		name     string
		resource interface{}
		expected string
	}{
		{
			"EC2Instance",
			&aws.EC2Instance{},
			"aws_instance",
		},
		{
			"TerraformConfig",
			&terraform.TerraformConfig{},
			"terraform_config",
		},
		{
			"EC2InstanceConfig",
			&terraform.EC2InstanceConfig{},
			"ec2_instance_config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.extractResourceType(tt.resource)
			if result != tt.expected {
				t.Errorf("extractResourceType() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestDetermineSeverity(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	tests := []struct {
		attrName string
		expected DriftSeverity
	}{
		{"security_groups", SeverityCritical},
		{"instance_type", SeverityCritical},
		{"ami", SeverityCritical},
		{"vpc_id", SeverityCritical},
		{"key_name", SeverityHigh},
		{"monitoring", SeverityHigh},
		{"ebs_optimized", SeverityHigh},
		{"tags", SeverityMedium},
		{"availability_zone", SeverityMedium},
		{"public_ip", SeverityLow},
		{"private_ip", SeverityLow},
		{"unknown_attribute", SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.attrName, func(t *testing.T) {
			result := detector.determineSeverity(tt.attrName, "value1", "value2")
			if result != tt.expected {
				t.Errorf("determineSeverity(%s) = %v, want %v", tt.attrName, result, tt.expected)
			}
		})
	}
}

func TestShouldIgnoreAttribute(t *testing.T) {
	config := DefaultDetectionConfig()
	config.IgnoredAttributes = []string{"launch_time", "state_transition_reason", "custom_ignored"}
	detector := NewDriftDetector(config)

	tests := []struct {
		attrName string
		expected bool
	}{
		{"launch_time", true},
		{"state_transition_reason", true},
		{"custom_ignored", true},
		{"instance_id", false},
		{"instance_type", false},
		{"unknown_attribute", false},
	}

	for _, tt := range tests {
		t.Run(tt.attrName, func(t *testing.T) {
			result := detector.shouldIgnoreAttribute(tt.attrName)
			if result != tt.expected {
				t.Errorf("shouldIgnoreAttribute(%s) = %v, want %v", tt.attrName, result, tt.expected)
			}
		})
	}
}

func TestGetAttributeConfig(t *testing.T) {
	config := DefaultDetectionConfig()
	config.AttributeConfigs["custom_attr"] = AttributeConfig{
		ComparisonType: FuzzyMatch,
		CaseSensitive:  false,
	}
	detector := NewDriftDetector(config)

	// Test existing attribute config
	attrConfig := detector.getAttributeConfig("custom_attr")
	if attrConfig.ComparisonType != FuzzyMatch {
		t.Errorf("Expected FuzzyMatch, got %v", attrConfig.ComparisonType)
	}
	if attrConfig.CaseSensitive {
		t.Error("Expected CaseSensitive to be false")
	}

	// Test non-existing attribute config (should return default)
	defaultConfig := detector.getAttributeConfig("non_existing_attr")
	if defaultConfig.ComparisonType != config.DefaultConfig.ComparisonType {
		t.Errorf("Expected default comparison type %v, got %v", config.DefaultConfig.ComparisonType, defaultConfig.ComparisonType)
	}
}

func TestGetAllAttributeNames(t *testing.T) {
	detector := NewDriftDetector(DefaultDetectionConfig())

	awsMap := map[string]interface{}{
		"instance_id":   "i-123",
		"instance_type": "t3.micro",
		"ami":           "ami-123",
		"aws_only_attr": "value",
	}

	terraformMap := map[string]interface{}{
		"instance_id":         "i-123",
		"instance_type":       "t3.micro",
		"resource_id":         "aws_instance.test",
		"terraform_only_attr": "value",
	}

	attributes := detector.getAllAttributeNames(awsMap, terraformMap)

	expectedAttrs := map[string]bool{
		"instance_id":         true,
		"instance_type":       true,
		"ami":                 true,
		"aws_only_attr":       true,
		"resource_id":         true,
		"terraform_only_attr": true,
	}

	if len(attributes) != len(expectedAttrs) {
		t.Errorf("Expected %d attributes, got %d", len(expectedAttrs), len(attributes))
	}

	for _, attr := range attributes {
		if !expectedAttrs[attr] {
			t.Errorf("Unexpected attribute: %s", attr)
		}
	}
}
