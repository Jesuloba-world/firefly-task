package drift

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"firefly-task/pkg/interfaces"
)

// ConcreteDriftDetector implements the interfaces.DriftDetector interface
type ConcreteDriftDetector struct {
	detector *DriftDetector
	logger   *logrus.Logger
}

// ConcreteDriftComparator implements the interfaces.DriftComparator interface
type ConcreteDriftComparator struct {
	logger *logrus.Logger
}

// ConcreteDriftAnalyzer implements the interfaces.DriftAnalyzer interface
type ConcreteDriftAnalyzer struct {
	logger *logrus.Logger
}

// NewConcreteDriftDetector creates a new concrete drift detector
func NewConcreteDriftDetector(logger *logrus.Logger) interfaces.DriftDetector {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	// Create a default detection config
	config := DetectionConfig{}
	detector := NewDriftDetector(config)
	return &ConcreteDriftDetector{
		detector: detector,
		logger:   logger,
	}
}

// NewConcreteDriftComparator creates a new concrete drift comparator
func NewConcreteDriftComparator(logger *logrus.Logger) interfaces.DriftComparator {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteDriftComparator{
		logger: logger,
	}
}

// NewConcreteDriftAnalyzer creates a new concrete drift analyzer
func NewConcreteDriftAnalyzer(logger *logrus.Logger) interfaces.DriftAnalyzer {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.InfoLevel)
	}

	return &ConcreteDriftAnalyzer{
		logger: logger,
	}
}

// AnalyzeDriftSeverity determines the severity level of detected drift
func (a *ConcreteDriftAnalyzer) AnalyzeDriftSeverity(driftResult *interfaces.DriftResult) interfaces.SeverityLevel {
	a.logger.Debugf("Analyzing drift severity for resource %s", driftResult.ResourceID)
	// Placeholder implementation
	if len(driftResult.DriftDetails) > 0 {
		return interfaces.SeverityHigh
	}
	return interfaces.SeverityNone
}

// DriftDetector implementation methods

// DetectDrift compares actual AWS resources with expected Terraform configuration
func (d *ConcreteDriftDetector) DetectDrift(actual *interfaces.EC2Instance, expected *interfaces.TerraformConfig, attributesToCheck []string) (*interfaces.DriftResult, error) {
	d.logger.Debugf("ConcreteDriftDetector: Detecting drift for single resource %s", actual.InstanceID)
	return d.detector.DetectDrift(actual, expected)
}

// DetectMultipleDrift performs drift detection on multiple resources
func (d *ConcreteDriftDetector) DetectMultipleDrift(actualResources map[string]*interfaces.EC2Instance, expectedConfigs map[string]*interfaces.TerraformConfig, attributesToCheck []string) (map[string]*interfaces.DriftResult, error) {
	d.logger.Debugf("ConcreteDriftDetector: Detecting drift for %d resources", len(actualResources))
	results := make(map[string]*interfaces.DriftResult)
	for id, actual := range actualResources {
		if expected, ok := expectedConfigs[id]; ok {
			result, err := d.DetectDrift(actual, expected, attributesToCheck)
			if err != nil {
				d.logger.Errorf("Error detecting drift for %s: %v", id, err)
				continue
			}
			results[id] = result
		}
	}
	return results, nil
}

// ValidateConfiguration validates that the Terraform configuration is valid
func (d *ConcreteDriftDetector) ValidateConfiguration(config *interfaces.TerraformConfig) error {
	d.logger.Debug("ConcreteDriftDetector: Validating configuration")
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}
	return nil
}

// DriftComparator implementation methods

// CompareAttribute compares a single attribute between actual and expected values
func (c *ConcreteDriftComparator) CompareAttribute(attributeName string, actual interface{}, expected interface{}) (*interfaces.DriftDetail, error) {
	c.logger.Debugf("Comparing attribute %s", attributeName)
	// Placeholder implementation
	return nil, nil
}

// CompareAttributes compares multiple attributes at once
func (c *ConcreteDriftComparator) CompareAttributes(actual map[string]interface{}, expected map[string]interface{}, attributesToCheck []string) ([]*interfaces.DriftDetail, error) {
	c.logger.Debug("ConcreteDriftComparator: Comparing attributes")
	// Placeholder implementation
	return nil, nil
}

// GetSupportedAttributes returns the list of attributes this comparator can handle
func (c *ConcreteDriftComparator) GetSupportedAttributes() []string {
	c.logger.Debug("ConcreteDriftComparator: Getting supported attributes")
	// Placeholder implementation
	return []string{"instance_type", "tags"}
}

// DriftAnalyzer implementation methods

// GroupDriftByType groups drift results by resource type
func (a *ConcreteDriftAnalyzer) GroupDriftByType(results map[string]*interfaces.DriftResult) map[string][]*interfaces.DriftResult {
	a.logger.Debug("ConcreteDriftAnalyzer: Grouping drift by type")
	// Placeholder implementation
	return nil
}

// FilterDriftByAttribute filters drift results by specific attributes
func (a *ConcreteDriftAnalyzer) FilterDriftByAttribute(results map[string]*interfaces.DriftResult, attributes []string) map[string]*interfaces.DriftResult {
	a.logger.Debug("ConcreteDriftAnalyzer: Filtering drift by attribute")
	// Placeholder implementation
	return nil
}

// CalculateDriftStatistics calculates statistics about the drift results
func (a *ConcreteDriftAnalyzer) CalculateDriftStatistics(results map[string]*interfaces.DriftResult) *interfaces.DriftStatistics {
	a.logger.Debug("ConcreteDriftAnalyzer: Calculating drift statistics")
	// Placeholder implementation
	return &interfaces.DriftStatistics{TotalResources: len(results)}
}

// DriftStatistics represents statistics about drift results
type DriftStatistics struct {
	TotalResources     int            `json:"total_resources"`
	ResourcesWithDrift int            `json:"resources_with_drift"`
	ResourcesNoDrift   int            `json:"resources_no_drift"`
	SeverityBreakdown  map[string]int `json:"severity_breakdown"`
}
