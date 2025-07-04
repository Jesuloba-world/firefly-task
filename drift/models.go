package drift

import (
	"fmt"
	"strings"
	"time"
)

// ComparisonType defines the type of comparison to perform
type ComparisonType int

const (
	// ExactMatch requires values to be identical
	ExactMatch ComparisonType = iota
	// FuzzyMatch allows for case-insensitive string comparison
	FuzzyMatch
	// NumericTolerance allows numeric comparison with tolerance
	NumericTolerance
	// ArrayUnordered compares arrays ignoring order
	ArrayUnordered
	// ArrayOrdered compares arrays considering order
	ArrayOrdered
	// MapComparison compares maps key by key
	MapComparison
	// NestedObject compares nested objects recursively
	NestedObject
)

// String returns the string representation of ComparisonType
func (ct ComparisonType) String() string {
	switch ct {
	case ExactMatch:
		return "exact"
	case FuzzyMatch:
		return "fuzzy"
	case NumericTolerance:
		return "numeric_tolerance"
	case ArrayUnordered:
		return "array_unordered"
	case ArrayOrdered:
		return "array_ordered"
	case MapComparison:
		return "map"
	case NestedObject:
		return "nested_object"
	default:
		return "unknown"
	}
}

// AttributeConfig defines how to compare a specific attribute
type AttributeConfig struct {
	// AttributeName is the name of the attribute to compare
	AttributeName string `json:"attribute_name"`

	// ComparisonType specifies how to compare the attribute
	ComparisonType ComparisonType `json:"comparison_type"`

	// Tolerance is used for numeric comparisons (optional)
	Tolerance *float64 `json:"tolerance,omitempty"`

	// CaseSensitive indicates if string comparisons should be case sensitive
	CaseSensitive bool `json:"case_sensitive"`

	// Required indicates if the attribute must be present in both configurations
	Required bool `json:"required"`

	// Description provides a human-readable description of what this attribute represents
	Description string `json:"description,omitempty"`
}

// String returns a string representation of the AttributeConfig
func (ac *AttributeConfig) String() string {
	return fmt.Sprintf("AttributeConfig{Name: %s, Type: %s, Required: %t}",
		ac.AttributeName, ac.ComparisonType.String(), ac.Required)
}

// AttributeDifference represents a single difference found between actual and expected values
type AttributeDifference struct {
	// AttributeName is the name of the attribute that differs
	AttributeName string `json:"attribute_name"`

	// ActualValue is the value found in the actual configuration
	ActualValue interface{} `json:"actual_value"`

	// ExpectedValue is the value from the expected configuration
	ExpectedValue interface{} `json:"expected_value"`

	// DifferenceType describes the type of difference
	DifferenceType string `json:"difference_type"`

	// Description provides a human-readable description of the difference
	Description string `json:"description"`

	// Severity indicates the importance of this difference
	Severity DriftSeverity `json:"severity"`
}

// String returns a string representation of the AttributeDifference
func (ad *AttributeDifference) String() string {
	return fmt.Sprintf("%s: expected '%v', got '%v' (%s)",
		ad.AttributeName, ad.ExpectedValue, ad.ActualValue, ad.DifferenceType)
}

// DriftSeverity indicates the severity level of a drift
type DriftSeverity int

const (
	// SeverityNone indicates no drift detected
	SeverityNone DriftSeverity = iota
	// SeverityLow indicates minor differences that may not require immediate action
	SeverityLow
	// SeverityMedium indicates differences that should be reviewed
	SeverityMedium
	// SeverityHigh indicates critical differences that require immediate attention
	SeverityHigh
	// SeverityCritical indicates differences that may cause system failures
	SeverityCritical
)

// String returns the string representation of DriftSeverity
func (ds DriftSeverity) String() string {
	switch ds {
	case SeverityNone:
		return "none"
	case SeverityLow:
		return "low"
	case SeverityMedium:
		return "medium"
	case SeverityHigh:
		return "high"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// DriftResult contains the results of comparing actual vs expected configurations
type DriftResult struct {
	// ResourceID identifies the resource that was checked
	ResourceID string `json:"resource_id"`

	// InstanceID is the AWS instance ID (if applicable)
	InstanceID string `json:"instance_id,omitempty"`

	// ResourceType indicates the type of resource (e.g., "ec2_instance")
	ResourceType string `json:"resource_type"`

	// DetectionTime indicates when the drift detection was performed
	DetectionTime time.Time `json:"detection_time"`

	// HasDrift indicates whether any drift was detected
	HasDrift bool `json:"has_drift"`

	// DriftedAttributes contains the list of attributes that have drift
	DriftedAttributes []string `json:"drifted_attributes"`

	// Differences contains detailed information about each difference found
	Differences []AttributeDifference `json:"differences"`

	// CheckedAttributes lists all attributes that were checked
	CheckedAttributes []string `json:"checked_attributes"`

	// Timestamp indicates when the drift check was performed
	Timestamp time.Time `json:"timestamp"`

	// OverallSeverity indicates the highest severity level found
	OverallSeverity DriftSeverity `json:"overall_severity"`

	// ErrorMessage contains any error that occurred during drift detection
	ErrorMessage string `json:"error_message,omitempty"`

	// Metadata contains additional information about the drift check
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AddDifference adds a new difference to the drift result
func (dr *DriftResult) AddDifference(diff AttributeDifference) {
	dr.Differences = append(dr.Differences, diff)
	dr.DriftedAttributes = append(dr.DriftedAttributes, diff.AttributeName)
	dr.HasDrift = true

	// Update overall severity if this difference is more severe
	if diff.Severity > dr.OverallSeverity {
		dr.OverallSeverity = diff.Severity
	}
}

// GetDriftedCount returns the number of drifted attributes
func (dr *DriftResult) GetDriftedCount() int {
	return len(dr.DriftedAttributes)
}

// GetHighestSeverity returns the overall severity of the drift result
func (dr *DriftResult) GetHighestSeverity() DriftSeverity {
	if len(dr.Differences) == 0 {
		return SeverityNone
	}

	highestSeverity := SeverityNone
	for _, diff := range dr.Differences {
		if diff.Severity > highestSeverity {
			highestSeverity = diff.Severity
		}
	}
	return highestSeverity
}

// String returns a string representation of the DriftResult
func (dr *DriftResult) String() string {
	if !dr.HasDrift {
		return fmt.Sprintf("DriftResult{ResourceID: %s, HasDrift: false, CheckedAttributes: %d}",
			dr.ResourceID, len(dr.CheckedAttributes))
	}

	return fmt.Sprintf("DriftResult{ResourceID: %s, HasDrift: true, Differences: %d, Severity: %s}",
		dr.ResourceID, len(dr.Differences), dr.OverallSeverity.String())
}

// Summary returns a human-readable summary of the drift result
func (dr *DriftResult) Summary() string {
	if !dr.HasDrift {
		return fmt.Sprintf("No drift detected for resource %s (checked %d attributes)",
			dr.ResourceID, len(dr.CheckedAttributes))
	}

	var severityCounts = make(map[DriftSeverity]int)
	for _, diff := range dr.Differences {
		severityCounts[diff.Severity]++
	}

	var parts []string
	for severity := SeverityCritical; severity >= SeverityLow; severity-- {
		if count := severityCounts[severity]; count > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", count, severity.String()))
		}
	}

	return fmt.Sprintf("Drift detected for resource %s: %s differences across %d attributes",
		dr.ResourceID, strings.Join(parts, ", "), len(dr.DriftedAttributes))
}

// NewDriftResult creates a new DriftResult with basic information
func NewDriftResult(resourceID, instanceID string) *DriftResult {
	return &DriftResult{
		ResourceID:        resourceID,
		InstanceID:        instanceID,
		HasDrift:          false,
		DriftedAttributes: make([]string, 0),
		Differences:       make([]AttributeDifference, 0),
		CheckedAttributes: make([]string, 0),
		Timestamp:         time.Now(),
		OverallSeverity:   SeverityLow,
		Metadata:          make(map[string]interface{}),
	}
}

// NewAttributeConfig creates a new AttributeConfig with default values
func NewAttributeConfig(name string, comparisonType ComparisonType) *AttributeConfig {
	return &AttributeConfig{
		AttributeName:  name,
		ComparisonType: comparisonType,
		CaseSensitive:  true,
		Required:       false,
	}
}

// WithTolerance sets the tolerance for numeric comparisons
func (ac *AttributeConfig) WithTolerance(tolerance float64) *AttributeConfig {
	ac.Tolerance = &tolerance
	return ac
}

// WithDescription sets the description for the attribute
func (ac *AttributeConfig) WithDescription(description string) *AttributeConfig {
	ac.Description = description
	return ac
}

// WithRequired sets whether the attribute is required
func (ac *AttributeConfig) WithRequired(required bool) *AttributeConfig {
	ac.Required = required
	return ac
}

// WithCaseSensitive sets whether string comparisons should be case sensitive
func (ac *AttributeConfig) WithCaseSensitive(caseSensitive bool) *AttributeConfig {
	ac.CaseSensitive = caseSensitive
	return ac
}
