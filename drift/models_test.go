package drift

import (
	"testing"
	"time"
)

func TestComparisonType_String(t *testing.T) {
	tests := []struct {
		name     string
		ct       ComparisonType
		expected string
	}{
		{"ExactMatch", ExactMatch, "exact"},
		{"FuzzyMatch", FuzzyMatch, "fuzzy"},
		{"NumericTolerance", NumericTolerance, "numeric_tolerance"},
		{"ArrayUnordered", ArrayUnordered, "array_unordered"},
		{"ArrayOrdered", ArrayOrdered, "array_ordered"},
		{"MapComparison", MapComparison, "map"},
		{"Unknown", ComparisonType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ct.String(); got != tt.expected {
				t.Errorf("ComparisonType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDriftSeverity_String(t *testing.T) {
	tests := []struct {
		name     string
		ds       DriftSeverity
		expected string
	}{
		{"SeverityLow", SeverityLow, "low"},
		{"SeverityMedium", SeverityMedium, "medium"},
		{"SeverityHigh", SeverityHigh, "high"},
		{"SeverityCritical", SeverityCritical, "critical"},
		{"Unknown", DriftSeverity(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.String(); got != tt.expected {
				t.Errorf("DriftSeverity.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAttributeConfig_String(t *testing.T) {
	config := &AttributeConfig{
		AttributeName:  "instance_type",
		ComparisonType: ExactMatch,
		Required:       true,
	}

	expected := "AttributeConfig{Name: instance_type, Type: exact, Required: true}"
	if got := config.String(); got != expected {
		t.Errorf("AttributeConfig.String() = %v, want %v", got, expected)
	}
}

func TestAttributeDifference_String(t *testing.T) {
	diff := &AttributeDifference{
		AttributeName:  "instance_type",
		ActualValue:    "t3.micro",
		ExpectedValue:  "t3.small",
		DifferenceType: "value_mismatch",
	}

	expected := "instance_type: expected 't3.small', got 't3.micro' (value_mismatch)"
	if got := diff.String(); got != expected {
		t.Errorf("AttributeDifference.String() = %v, want %v", got, expected)
	}
}

func TestNewDriftResult(t *testing.T) {
	resourceID := "aws_instance.web"
	instanceID := "i-1234567890abcdef0"

	result := NewDriftResult(resourceID, instanceID)

	if result.ResourceID != resourceID {
		t.Errorf("NewDriftResult() ResourceID = %v, want %v", result.ResourceID, resourceID)
	}
	if result.InstanceID != instanceID {
		t.Errorf("NewDriftResult() InstanceID = %v, want %v", result.InstanceID, instanceID)
	}
	if result.HasDrift {
		t.Errorf("NewDriftResult() HasDrift = %v, want false", result.HasDrift)
	}
	if result.OverallSeverity != SeverityLow {
		t.Errorf("NewDriftResult() OverallSeverity = %v, want %v", result.OverallSeverity, SeverityLow)
	}
	if len(result.DriftedAttributes) != 0 {
		t.Errorf("NewDriftResult() DriftedAttributes length = %v, want 0", len(result.DriftedAttributes))
	}
	if len(result.Differences) != 0 {
		t.Errorf("NewDriftResult() Differences length = %v, want 0", len(result.Differences))
	}
	if len(result.CheckedAttributes) != 0 {
		t.Errorf("NewDriftResult() CheckedAttributes length = %v, want 0", len(result.CheckedAttributes))
	}
	if result.Metadata == nil {
		t.Errorf("NewDriftResult() Metadata should not be nil")
	}
}

func TestDriftResult_AddDifference(t *testing.T) {
	result := NewDriftResult("aws_instance.web", "i-1234567890abcdef0")

	diff1 := AttributeDifference{
		AttributeName:  "instance_type",
		ActualValue:    "t3.micro",
		ExpectedValue:  "t3.small",
		DifferenceType: "value_mismatch",
		Severity:       SeverityMedium,
	}

	diff2 := AttributeDifference{
		AttributeName:  "ami",
		ActualValue:    "ami-12345",
		ExpectedValue:  "ami-67890",
		DifferenceType: "value_mismatch",
		Severity:       SeverityHigh,
	}

	result.AddDifference(diff1)
	if !result.HasDrift {
		t.Errorf("AddDifference() should set HasDrift to true")
	}
	if len(result.Differences) != 1 {
		t.Errorf("AddDifference() Differences length = %v, want 1", len(result.Differences))
	}
	if len(result.DriftedAttributes) != 1 {
		t.Errorf("AddDifference() DriftedAttributes length = %v, want 1", len(result.DriftedAttributes))
	}
	if result.OverallSeverity != SeverityMedium {
		t.Errorf("AddDifference() OverallSeverity = %v, want %v", result.OverallSeverity, SeverityMedium)
	}

	result.AddDifference(diff2)
	if len(result.Differences) != 2 {
		t.Errorf("AddDifference() Differences length = %v, want 2", len(result.Differences))
	}
	if result.OverallSeverity != SeverityHigh {
		t.Errorf("AddDifference() OverallSeverity = %v, want %v", result.OverallSeverity, SeverityHigh)
	}
}

// Removed tests for deleted methods GetDifferencesBySeverity, GetHighestSeverityDifferences, HasAttribute, and HasDriftInAttribute

func TestDriftResult_String(t *testing.T) {
	// Test with no drift
	result1 := NewDriftResult("aws_instance.web", "i-1234567890abcdef0")
	result1.CheckedAttributes = []string{"instance_type", "ami"}

	expected1 := "DriftResult{ResourceID: aws_instance.web, HasDrift: false, CheckedAttributes: 2}"
	if got := result1.String(); got != expected1 {
		t.Errorf("DriftResult.String() = %v, want %v", got, expected1)
	}

	// Test with drift
	result2 := NewDriftResult("aws_instance.web", "i-1234567890abcdef0")
	diff := AttributeDifference{AttributeName: "instance_type", Severity: SeverityHigh}
	result2.AddDifference(diff)

	expected2 := "DriftResult{ResourceID: aws_instance.web, HasDrift: true, Differences: 1, Severity: high}"
	if got := result2.String(); got != expected2 {
		t.Errorf("DriftResult.String() = %v, want %v", got, expected2)
	}
}

func TestDriftResult_Summary(t *testing.T) {
	// Test with no drift
	result1 := NewDriftResult("aws_instance.web", "i-1234567890abcdef0")
	result1.CheckedAttributes = []string{"instance_type", "ami"}

	expected1 := "No drift detected for resource aws_instance.web (checked 2 attributes)"
	if got := result1.Summary(); got != expected1 {
		t.Errorf("DriftResult.Summary() = %v, want %v", got, expected1)
	}

	// Test with drift
	result2 := NewDriftResult("aws_instance.web", "i-1234567890abcdef0")
	diff1 := AttributeDifference{AttributeName: "instance_type", Severity: SeverityHigh}
	diff2 := AttributeDifference{AttributeName: "ami", Severity: SeverityMedium}
	result2.AddDifference(diff1)
	result2.AddDifference(diff2)

	expected2 := "Drift detected for resource aws_instance.web: 1 high, 1 medium differences across 2 attributes"
	if got := result2.Summary(); got != expected2 {
		t.Errorf("DriftResult.Summary() = %v, want %v", got, expected2)
	}
}

func TestNewAttributeConfig(t *testing.T) {
	config := NewAttributeConfig("instance_type", ExactMatch)

	if config.AttributeName != "instance_type" {
		t.Errorf("NewAttributeConfig() AttributeName = %v, want instance_type", config.AttributeName)
	}
	if config.ComparisonType != ExactMatch {
		t.Errorf("NewAttributeConfig() ComparisonType = %v, want %v", config.ComparisonType, ExactMatch)
	}
	if !config.CaseSensitive {
		t.Errorf("NewAttributeConfig() CaseSensitive = %v, want true", config.CaseSensitive)
	}
	if config.Required {
		t.Errorf("NewAttributeConfig() Required = %v, want false", config.Required)
	}
}

func TestAttributeConfig_FluentInterface(t *testing.T) {
	tolerance := 0.1
	config := NewAttributeConfig("cpu_utilization", NumericTolerance).
		WithTolerance(tolerance).
		WithDescription("CPU utilization percentage").
		WithRequired(true).
		WithCaseSensitive(false)

	if config.Tolerance == nil || *config.Tolerance != tolerance {
		t.Errorf("WithTolerance() failed, got %v, want %v", config.Tolerance, tolerance)
	}
	if config.Description != "CPU utilization percentage" {
		t.Errorf("WithDescription() failed, got %v, want 'CPU utilization percentage'", config.Description)
	}
	if !config.Required {
		t.Errorf("WithRequired() failed, got %v, want true", config.Required)
	}
	if config.CaseSensitive {
		t.Errorf("WithCaseSensitive() failed, got %v, want false", config.CaseSensitive)
	}
}

// Removed test for deleted GetDifferenceCount method

func TestDriftResult_Timestamp(t *testing.T) {
	before := time.Now()
	result := NewDriftResult("aws_instance.web", "i-1234567890abcdef0")
	after := time.Now()

	if result.Timestamp.Before(before) || result.Timestamp.After(after) {
		t.Errorf("NewDriftResult() timestamp %v should be between %v and %v", result.Timestamp, before, after)
	}
}
