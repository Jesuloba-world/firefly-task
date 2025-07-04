package aws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEC2Instance_ToJSON(t *testing.T) {
	launchTime := time.Now()
	instance := &EC2Instance{
		InstanceID:       "i-1234567890abcdef0",
		InstanceType:     "t3.micro",
		State:            InstanceStateRunning,
		PublicIPAddress:  aws.String("203.0.113.12"),
		PrivateIPAddress: aws.String("10.0.1.55"),
		SecurityGroups: []SecurityGroup{
			{GroupID: "sg-12345678", GroupName: "default"},
		},
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "test",
		},
		SubnetID:         aws.String("subnet-12345678"),
		VPCID:            aws.String("vpc-12345678"),
		AvailabilityZone: aws.String("us-east-1a"),
		LaunchTime:       &launchTime,
		ImageID:          aws.String("ami-12345678"),
		KeyName:          aws.String("my-key"),
		Monitoring:       true,
		EBSOptimized:     false,
	}

	jsonStr, err := instance.ToJSON()
	require.NoError(t, err)
	assert.NotEmpty(t, jsonStr)

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &parsed)
	require.NoError(t, err)

	// Check some key fields
	assert.Equal(t, "i-1234567890abcdef0", parsed["instance_id"])
	assert.Equal(t, "t3.micro", parsed["instance_type"])
	assert.Equal(t, InstanceStateRunning, parsed["state"])
}

func TestFromJSON(t *testing.T) {
	jsonStr := `{
		"instance_id": "i-1234567890abcdef0",
		"instance_type": "t3.micro",
		"state": "running",
		"public_ip_address": "203.0.113.12",
		"private_ip_address": "10.0.1.55",
		"security_groups": [
			{"group_id": "sg-12345678", "group_name": "default"}
		],
		"tags": {
			"Name": "test-instance",
			"Environment": "test"
		},
		"subnet_id": "subnet-12345678",
		"vpc_id": "vpc-12345678",
		"availability_zone": "us-east-1a",
		"monitoring": true,
		"ebs_optimized": false
	}`

	instance, err := FromJSON(jsonStr)
	require.NoError(t, err)
	require.NotNil(t, instance)

	assert.Equal(t, "i-1234567890abcdef0", instance.InstanceID)
	assert.Equal(t, "t3.micro", instance.InstanceType)
	assert.Equal(t, InstanceStateRunning, instance.State)
	assert.Equal(t, "203.0.113.12", *instance.PublicIPAddress)
	assert.Equal(t, "10.0.1.55", *instance.PrivateIPAddress)
	assert.Len(t, instance.SecurityGroups, 1)
	assert.Equal(t, "sg-12345678", instance.SecurityGroups[0].GroupID)
	assert.Equal(t, "default", instance.SecurityGroups[0].GroupName)
	assert.Equal(t, "test-instance", instance.Tags["Name"])
	assert.Equal(t, "test", instance.Tags["Environment"])
	assert.True(t, instance.Monitoring)
	assert.False(t, instance.EBSOptimized)
}

func TestEC2Instance_GetTag(t *testing.T) {
	instance := &EC2Instance{
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "production",
		},
	}

	assert.Equal(t, "test-instance", instance.GetTag("Name"))
	assert.Equal(t, "production", instance.GetTag("Environment"))
	assert.Equal(t, "", instance.GetTag("NonExistent"))

	// Test with nil tags
	instance.Tags = nil
	assert.Equal(t, "", instance.GetTag("Name"))
}

func TestEC2Instance_HasTag(t *testing.T) {
	instance := &EC2Instance{
		Tags: map[string]string{
			"Name":        "test-instance",
			"Environment": "production",
		},
	}

	assert.True(t, instance.HasTag("Name"))
	assert.True(t, instance.HasTag("Environment"))
	assert.False(t, instance.HasTag("NonExistent"))

	// Test with nil tags
	instance.Tags = nil
	assert.False(t, instance.HasTag("Name"))
}

func TestEC2Instance_IsRunning(t *testing.T) {
	tests := []struct {
		state    string
		expected bool
	}{
		{InstanceStateRunning, true},
		{InstanceStateStopped, false},
		{InstanceStatePending, false},
		{InstanceStateStopping, false},
		{InstanceStateTerminated, false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			instance := &EC2Instance{State: tt.state}
			assert.Equal(t, tt.expected, instance.IsRunning())
		})
	}
}

func TestEC2Instance_IsStopped(t *testing.T) {
	tests := []struct {
		state    string
		expected bool
	}{
		{InstanceStateStopped, true},
		{InstanceStateRunning, false},
		{InstanceStatePending, false},
		{InstanceStateStopping, false},
		{InstanceStateTerminated, false},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			instance := &EC2Instance{State: tt.state}
			assert.Equal(t, tt.expected, instance.IsStopped())
		})
	}
}

func TestConvertFromAWSInstance(t *testing.T) {
	launchTime := time.Now()
	awsInstance := types.Instance{
		InstanceId:       aws.String("i-1234567890abcdef0"),
		InstanceType:     types.InstanceTypeT3Micro,
		State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
		PublicIpAddress:  aws.String("203.0.113.12"),
		PrivateIpAddress: aws.String("10.0.1.55"),
		PublicDnsName:    aws.String("ec2-203-0-113-12.compute-1.amazonaws.com"),
		PrivateDnsName:   aws.String("ip-10-0-1-55.ec2.internal"),
		SubnetId:         aws.String("subnet-12345678"),
		VpcId:            aws.String("vpc-12345678"),
		Placement:        &types.Placement{AvailabilityZone: aws.String("us-east-1a")},
		LaunchTime:       &launchTime,
		ImageId:          aws.String("ami-12345678"),
		KeyName:          aws.String("my-key"),
		Platform:         types.PlatformValuesWindows,
		Architecture:     types.ArchitectureValuesX8664,
		Monitoring:       &types.Monitoring{State: types.MonitoringStateEnabled},
		EbsOptimized:     aws.Bool(true),
		SecurityGroups: []types.GroupIdentifier{
			{
				GroupId:   aws.String("sg-12345678"),
				GroupName: aws.String("default"),
			},
		},
		Tags: []types.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-instance"),
			},
			{
				Key:   aws.String("Environment"),
				Value: aws.String("test"),
			},
		},
	}

	instance := convertFromAWSInstance(awsInstance)

	assert.Equal(t, "i-1234567890abcdef0", instance.InstanceID)
	assert.Equal(t, "t3.micro", instance.InstanceType)
	assert.Equal(t, InstanceStateRunning, instance.State)
	assert.Equal(t, "203.0.113.12", *instance.PublicIPAddress)
	assert.Equal(t, "10.0.1.55", *instance.PrivateIPAddress)
	assert.Equal(t, "ec2-203-0-113-12.compute-1.amazonaws.com", *instance.PublicDNSName)
	assert.Equal(t, "ip-10-0-1-55.ec2.internal", *instance.PrivateDNSName)
	assert.Equal(t, "subnet-12345678", *instance.SubnetID)
	assert.Equal(t, "vpc-12345678", *instance.VPCID)
	assert.Equal(t, "us-east-1a", *instance.AvailabilityZone)
	assert.Equal(t, launchTime, *instance.LaunchTime)
	assert.Equal(t, "ami-12345678", *instance.ImageID)
	assert.Equal(t, "my-key", *instance.KeyName)
	assert.Equal(t, "Windows", *instance.Platform)
	assert.Equal(t, "x86_64", *instance.Architecture)
	assert.True(t, instance.Monitoring)
	assert.True(t, instance.EBSOptimized)

	// Check security groups
	require.Len(t, instance.SecurityGroups, 1)
	assert.Equal(t, "sg-12345678", instance.SecurityGroups[0].GroupID)
	assert.Equal(t, "default", instance.SecurityGroups[0].GroupName)

	// Check tags
	assert.Equal(t, "test-instance", instance.Tags["Name"])
	assert.Equal(t, "test", instance.Tags["Environment"])
}
