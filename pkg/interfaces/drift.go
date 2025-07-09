package interfaces

// DriftDetector defines the interface for drift detection operations
type DriftDetector interface {
	// DetectDrift compares actual AWS resources with expected Terraform configuration
	DetectDrift(actual *EC2Instance, expected *TerraformConfig, attributesToCheck []string) (*DriftResult, error)

	// DetectMultipleDrift performs drift detection on multiple resources
	DetectMultipleDrift(actualResources map[string]*EC2Instance, expectedConfigs map[string]*TerraformConfig, attributesToCheck []string) (map[string]*DriftResult, error)

	// ValidateConfiguration validates that the Terraform configuration is valid
	ValidateConfiguration(config *TerraformConfig) error
}

// DriftComparator defines the interface for comparing individual attributes
type DriftComparator interface {
	// CompareAttribute compares a single attribute between actual and expected values
	CompareAttribute(attributeName string, actual interface{}, expected interface{}) (*DriftDetail, error)
	
	// CompareAttributes compares multiple attributes at once
	CompareAttributes(actual map[string]interface{}, expected map[string]interface{}, attributesToCheck []string) ([]*DriftDetail, error)
	
	// GetSupportedAttributes returns the list of attributes this comparator can handle
	GetSupportedAttributes() []string
}

// DriftAnalyzer defines the interface for analyzing drift results
type DriftAnalyzer interface {
	// AnalyzeDriftSeverity determines the severity level of detected drift
	AnalyzeDriftSeverity(driftResult *DriftResult) SeverityLevel
	
	// GroupDriftByType groups drift results by resource type
	GroupDriftByType(results map[string]*DriftResult) map[string][]*DriftResult
	
	// FilterDriftByAttribute filters drift results by specific attributes
	FilterDriftByAttribute(results map[string]*DriftResult, attributes []string) map[string]*DriftResult
	
	// CalculateDriftStatistics calculates statistics about the drift results
	CalculateDriftStatistics(results map[string]*DriftResult) *DriftStatistics
}