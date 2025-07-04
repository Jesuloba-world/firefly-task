package drift

import (
	"fmt"
	"reflect"
	"sync"
	"time"
	"unicode"

	"firefly-task/aws"
	"firefly-task/terraform"
)

// DetectionConfig holds configuration for drift detection
type DetectionConfig struct {
	// AttributeConfigs maps attribute names to their comparison configurations
	AttributeConfigs map[string]AttributeConfig

	// DefaultConfig is used for attributes not explicitly configured
	DefaultConfig AttributeConfig

	// IgnoredAttributes lists attributes to skip during comparison
	IgnoredAttributes []string

	// StrictMode determines if unknown attributes should cause errors
	StrictMode bool

	// MaxConcurrency limits the number of concurrent drift detections
	MaxConcurrency int

	// Timeout for individual drift detection operations
	Timeout time.Duration
}

// DefaultDetectionConfig returns a sensible default configuration
func DefaultDetectionConfig() DetectionConfig {
	return DetectionConfig{
		AttributeConfigs: map[string]AttributeConfig{
			"instance_id":                          {ComparisonType: ExactMatch, CaseSensitive: true},
			"instance_type":                        {ComparisonType: ExactMatch, CaseSensitive: true},
			"ami":                                  {ComparisonType: ExactMatch, CaseSensitive: true},
			"state":                                {ComparisonType: ExactMatch, CaseSensitive: false},
			"public_ip":                            {ComparisonType: ExactMatch, CaseSensitive: true},
			"private_ip":                           {ComparisonType: ExactMatch, CaseSensitive: true},
			"public_dns":                           {ComparisonType: ExactMatch, CaseSensitive: false},
			"private_dns":                          {ComparisonType: ExactMatch, CaseSensitive: false},
			"security_groups":                      {ComparisonType: ArrayUnordered},
			"tags":                                 {ComparisonType: MapComparison},
			"subnet_id":                            {ComparisonType: ExactMatch, CaseSensitive: true},
			"vpc_id":                               {ComparisonType: ExactMatch, CaseSensitive: true},
			"availability_zone":                    {ComparisonType: ExactMatch, CaseSensitive: true},
			"key_name":                             {ComparisonType: ExactMatch, CaseSensitive: true},
			"monitoring":                           {ComparisonType: ExactMatch},
			"ebs_optimized":                        {ComparisonType: ExactMatch},
			"source_dest_check":                    {ComparisonType: ExactMatch},
			"disable_api_termination":              {ComparisonType: ExactMatch},
			"instance_initiated_shutdown_behavior": {ComparisonType: ExactMatch, CaseSensitive: false},
			"placement_group":                      {ComparisonType: ExactMatch, CaseSensitive: true},
			"tenancy":                              {ComparisonType: ExactMatch, CaseSensitive: false},
			"host_id":                              {ComparisonType: ExactMatch, CaseSensitive: true},
			"cpu_core_count":                       {ComparisonType: ExactMatch},
			"cpu_threads_per_core":                 {ComparisonType: ExactMatch},
			"root_device_name":                     {ComparisonType: ExactMatch, CaseSensitive: true},
			"root_device_type":                     {ComparisonType: ExactMatch, CaseSensitive: false},
			"block_device_mappings":                {ComparisonType: ArrayUnordered},
		},
		DefaultConfig: AttributeConfig{
			ComparisonType: ExactMatch,
			CaseSensitive:  true,
		},
		IgnoredAttributes: []string{
			"launch_time",              // AWS-managed, changes on every start
			"state_transition_reason",  // AWS-managed
			"state_reason",             // AWS-managed
			"network_interfaces",       // Complex nested structure, handled separately
			"security_groups_detailed", // Redundant with security_groups
		},
		StrictMode:     false,
		MaxConcurrency: 10,
		Timeout:        30 * time.Second,
	}
}

// DriftDetector handles drift detection operations
type DriftDetector struct {
	config DetectionConfig
	mu     sync.RWMutex
}

// NewDriftDetector creates a new drift detector with the given configuration
func NewDriftDetector(config DetectionConfig) *DriftDetector {
	return &DriftDetector{
		config: config,
	}
}

// DetectDrift compares an AWS resource with its Terraform configuration
func (d *DriftDetector) DetectDrift(awsResource interface{}, terraformConfig interface{}) (*DriftResult, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if awsResource == nil || terraformConfig == nil {
		return nil, fmt.Errorf("both AWS resource and Terraform configuration must be provided")
	}

	// Convert resources to comparable maps
	awsMap, err := d.resourceToMap(awsResource)
	if err != nil {
		return nil, fmt.Errorf("failed to convert AWS resource: %w", err)
	}

	terraformMap, err := d.resourceToMap(terraformConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Terraform configuration: %w", err)
	}

	// Perform drift detection
	result := &DriftResult{
		ResourceID:    d.extractResourceID(awsResource),
		ResourceType:  d.extractResourceType(awsResource),
		DetectionTime: time.Now(),
		Differences:   []AttributeDifference{},
	}

	// Get all unique attribute names
	attributeNames := d.getAllAttributeNames(awsMap, terraformMap)

	// Compare each attribute
	for _, attrName := range attributeNames {
		if d.shouldIgnoreAttribute(attrName) {
			continue
		}

		awsValue, awsExists := awsMap[attrName]
		terraformValue, terraformExists := terraformMap[attrName]

		// Handle missing attributes
		if !awsExists && !terraformExists {
			continue
		}

		if !awsExists {
			result.Differences = append(result.Differences, AttributeDifference{
				AttributeName: attrName,
				ActualValue:   nil,
				ExpectedValue: terraformValue,
				Severity:      SeverityMedium,
				Description:   fmt.Sprintf("Attribute '%s' missing in AWS resource but present in Terraform configuration", attrName),
			})
			continue
		}

		if !terraformExists {
			result.Differences = append(result.Differences, AttributeDifference{
				AttributeName: attrName,
				ActualValue:   awsValue,
				ExpectedValue: nil,
				Severity:      SeverityLow,
				Description:   fmt.Sprintf("Attribute '%s' present in AWS resource but missing in Terraform configuration", attrName),
			})
			continue
		}

		// Compare attribute values
		config := d.getAttributeConfig(attrName)
		isEqual, description := CompareValues(awsValue, terraformValue, config)

		if !isEqual {
			severity := d.determineSeverity(attrName, awsValue, terraformValue)
			result.Differences = append(result.Differences, AttributeDifference{
				AttributeName: attrName,
				ActualValue:   awsValue,
				ExpectedValue: terraformValue,
				Severity:      severity,
				Description:   description,
			})
		}
	}

	// Determine overall drift status
	result.HasDrift = len(result.Differences) > 0
	if result.HasDrift {
		result.OverallSeverity = result.GetHighestSeverity()
	} else {
		result.OverallSeverity = SeverityNone
	}

	return result, nil
}

func (d *DriftDetector) toSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// DetectDriftBatch performs drift detection on multiple resource pairs concurrently
func (d *DriftDetector) DetectDriftBatch(resourcePairs []ResourcePair) ([]*DriftResult, error) {
	d.mu.RLock()
	maxConcurrency := d.config.MaxConcurrency
	d.mu.RUnlock()

	if maxConcurrency <= 0 {
		maxConcurrency = 1
	}

	// Create channels for work distribution
	workChan := make(chan ResourcePair, len(resourcePairs))
	resultChan := make(chan BatchResult, len(resourcePairs))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pair := range workChan {
				result, err := d.DetectDrift(pair.AWSResource, pair.TerraformConfig)
				resultChan <- BatchResult{
					Index:  pair.Index,
					Result: result,
					Error:  err,
				}
			}
		}()
	}

	// Send work to workers
	go func() {
		defer close(workChan)
		for _, pair := range resourcePairs {
			workChan <- pair
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results
	results := make([]*DriftResult, len(resourcePairs))
	var errors []error

	for batchResult := range resultChan {
		if batchResult.Error != nil {
			errors = append(errors, fmt.Errorf("index %d: %w", batchResult.Index, batchResult.Error))
			continue
		}
		results[batchResult.Index] = batchResult.Result
	}

	if len(errors) > 0 {
		return results, fmt.Errorf("batch processing errors: %v", errors)
	}

	return results, nil
}

// UpdateConfig updates the detector's configuration
func (d *DriftDetector) UpdateConfig(config DetectionConfig) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.config = config
}

// GetConfig returns a copy of the current configuration
func (d *DriftDetector) GetConfig() DetectionConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.config
}

// Helper types for batch processing
type ResourcePair struct {
	Index           int
	AWSResource     interface{}
	TerraformConfig interface{}
}

type BatchResult struct {
	Index  int
	Result *DriftResult
	Error  error
}

// Helper methods

func (d *DriftDetector) resourceToMap(resource interface{}) (map[string]interface{}, error) {
	switch r := resource.(type) {
	case *aws.EC2Instance:
		return d.ec2InstanceToMap(r), nil
	case *terraform.TerraformConfig:
		return d.terraformConfigToMap(r), nil
	case *terraform.EC2InstanceConfig:
		return d.ec2InstanceConfigToMap(r), nil
	default:
		// Use reflection as fallback
		return d.reflectToMap(resource)
	}
}

func (d *DriftDetector) ec2InstanceToMap(instance *aws.EC2Instance) map[string]interface{} {
	m := map[string]interface{}{
		"instance_id":   instance.InstanceID,
		"instance_type": instance.InstanceType,
		"tags":          instance.Tags,
		"monitoring":    instance.Monitoring,
		"ebs_optimized": instance.EBSOptimized,
	}

	// Handle pointer fields
	if instance.ImageID != nil {
		m["ami"] = *instance.ImageID
	}

	// Handle security groups - extract just the group IDs
	if len(instance.SecurityGroups) > 0 {
		groupIDs := make([]string, len(instance.SecurityGroups))
		for i, sg := range instance.SecurityGroups {
			groupIDs[i] = sg.GroupID
		}
		m["security_groups"] = groupIDs
	}

	return m
}

func (d *DriftDetector) terraformConfigToMap(config *terraform.TerraformConfig) map[string]interface{} {
	m := map[string]interface{}{
		"instance_id":   config.InstanceID,
		"instance_type": config.InstanceType,
		"ami":           config.AMI,
		"tags":          config.Tags,
	}

	// Handle security groups - prefer SecurityGroupRefs over SecurityGroups
	if len(config.SecurityGroupRefs) > 0 {
		m["security_groups"] = config.SecurityGroupRefs
	} else if len(config.SecurityGroups) > 0 {
		m["security_groups"] = config.SecurityGroups
	}

	// Add monitoring and EBS optimization if they have values
	if config.Monitoring != nil {
		m["monitoring"] = *config.Monitoring
	}
	if config.EBSOptimized != nil {
		m["ebs_optimized"] = *config.EBSOptimized
	}

	// Remove nil values
	for k, v := range m {
		if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
			delete(m, k)
		}
	}

	return m
}

func (d *DriftDetector) ec2InstanceConfigToMap(config *terraform.EC2InstanceConfig) map[string]interface{} {
	m := map[string]interface{}{
		"instance_type":          config.InstanceType,
		"ami":                    config.AMI,
		"tags":                   config.Tags,
		"vpc_security_group_ids": config.VPCSecurityGroups,
		"subnet_id":              config.SubnetID,
		"key_name":               config.KeyName,
		"user_data":              config.UserData,
		"resource_name":          config.ResourceName,
	}

	// Remove nil values
	for k, v := range m {
		if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
			delete(m, k)
		}
	}

	return m
}

func (d *DriftDetector) reflectToMap(resource interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	v := reflect.ValueOf(resource)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, fmt.Errorf("resource is nil")
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("resource must be a struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Convert field name to snake_case
		fieldName := d.toSnakeCase(field.Name)
		result[fieldName] = fieldValue.Interface()
	}

	return result, nil
}

func (d *DriftDetector) extractResourceID(resource interface{}) string {
	switch r := resource.(type) {
	case *aws.EC2Instance:
		return r.InstanceID
	case *terraform.TerraformConfig:
		return r.ResourceID
	case *terraform.EC2InstanceConfig:
		return "" // EC2InstanceConfig doesn't have a resource ID
	default:
		return "unknown"
	}
}

func (d *DriftDetector) extractResourceType(resource interface{}) string {
	switch resource.(type) {
	case *aws.EC2Instance:
		return "aws_instance"
	case *terraform.TerraformConfig:
		return "terraform_config"
	case *terraform.EC2InstanceConfig:
		return "ec2_instance_config"
	default:
		return reflect.TypeOf(resource).String()
	}
}

func (d *DriftDetector) getAllAttributeNames(awsMap, terraformMap map[string]interface{}) []string {
	attributeSet := make(map[string]bool)

	for name := range awsMap {
		attributeSet[name] = true
	}

	for name := range terraformMap {
		attributeSet[name] = true
	}

	attributes := make([]string, 0, len(attributeSet))
	for name := range attributeSet {
		attributes = append(attributes, name)
	}

	return attributes
}

func (d *DriftDetector) shouldIgnoreAttribute(attrName string) bool {
	for _, ignored := range d.config.IgnoredAttributes {
		if attrName == ignored {
			return true
		}
	}
	return false
}

func (d *DriftDetector) getAttributeConfig(attrName string) AttributeConfig {
	if config, exists := d.config.AttributeConfigs[attrName]; exists {
		return config
	}
	return d.config.DefaultConfig
}

func (d *DriftDetector) determineSeverity(attrName string, awsValue, terraformValue interface{}) DriftSeverity {
	// Critical attributes that affect security or functionality
	criticalAttrs := map[string]bool{
		"security_groups":         true,
		"instance_type":           true,
		"ami":                     true,
		"vpc_id":                  true,
		"subnet_id":               true,
		"disable_api_termination": true,
	}

	// High priority attributes
	highAttrs := map[string]bool{
		"key_name":                             true,
		"monitoring":                           true,
		"ebs_optimized":                        true,
		"source_dest_check":                    true,
		"instance_initiated_shutdown_behavior": true,
		"tenancy":                              true,
		"placement_group":                      true,
		"root_device_type":                     true,
		"block_device_mappings":                true,
	}

	// Medium priority attributes
	mediumAttrs := map[string]bool{
		"tags":                 true,
		"availability_zone":    true,
		"cpu_core_count":       true,
		"cpu_threads_per_core": true,
		"root_device_name":     true,
	}

	if criticalAttrs[attrName] {
		return SeverityCritical
	}
	if highAttrs[attrName] {
		return SeverityHigh
	}
	if mediumAttrs[attrName] {
		return SeverityMedium
	}

	return SeverityLow
}
