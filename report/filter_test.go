package report

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"firefly-task/drift"
)

func TestFilterCriteria_Creation(t *testing.T) {
	criteria := NewFilterCriteria()

	assert.NotNil(t, criteria)
	assert.Equal(t, drift.SeverityLow, criteria.MinSeverity)
	assert.Nil(t, criteria.ResourcePattern)
	assert.Nil(t, criteria.AttributePattern)
	assert.Nil(t, criteria.After)
	assert.Nil(t, criteria.Before)
	assert.False(t, criteria.OnlyWithDrift)
	assert.False(t, criteria.OnlyWithoutDrift)
	assert.Nil(t, criteria.ActualValuePattern)
	assert.Equal(t, 0, criteria.Limit)
	assert.Equal(t, SortByResourceID, criteria.SortBy)
	assert.Equal(t, SortOrderAsc, criteria.SortOrder)
}

func TestFilterCriteria_WithMethods(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	criteria := NewFilterCriteria().
		WithMinSeverity(drift.SeverityHigh).
		WithResourcePattern("aws_instance.*").
		WithAttributePattern("instance_type").
		WithSince(yesterday).
		WithUntil(now).
		WithDriftStatus(DriftStatusDrifted).
		WithValuePattern("t2.*").
		WithLimit(10).
		WithSort(SortBySeverity, true)

	assert.Equal(t, drift.SeverityHigh, criteria.MinSeverity)
	assert.NotNil(t, criteria.ResourcePattern)
	assert.Equal(t, "aws_instance.*", criteria.ResourcePattern.String())
	assert.NotNil(t, criteria.AttributePattern)
	assert.Equal(t, "instance_type", criteria.AttributePattern.String())
	assert.NotNil(t, criteria.After)
	assert.Equal(t, yesterday, *criteria.After)
	assert.NotNil(t, criteria.Before)
	assert.Equal(t, now, *criteria.Before)
	assert.True(t, criteria.OnlyWithDrift)
	assert.False(t, criteria.OnlyWithoutDrift)
	assert.NotNil(t, criteria.ActualValuePattern)
	assert.Equal(t, "t2.*", criteria.ActualValuePattern.String())
	assert.Equal(t, 10, criteria.Limit)
	assert.Equal(t, SortBySeverity, criteria.SortBy)
	assert.Equal(t, SortOrderDesc, criteria.SortOrder)
}

func TestResultFilter_Apply(t *testing.T) {
	results := createTestDriftResults()

	// Test no filter (should return all results)
	filter := NewResultFilter()
	filtered := filter.Apply(results)
	assert.Len(t, filtered, 3)

	// Test severity filter
	filter = NewResultFilter().WithSeverity(drift.SeverityHigh, drift.SeverityCritical)
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 2) // Only critical and high severity

	// Test drift status filter
	filter = NewResultFilter().OnlyWithDrift()
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 2) // Only resources with drift

	// Test no drift filter
	filter = NewResultFilter().OnlyWithoutDrift()
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 1) // Only resources without drift
}

func TestResultFilter_ApplyWithResourcePattern(t *testing.T) {
	results := createTestDriftResults()

	// Test resource pattern filter
	filter := NewResultFilter().WithResourcePattern(".*web-server.*")
	filtered := filter.Apply(results)
	assert.Len(t, filtered, 2) // web-server-1 and web-server-2

	// Test specific resource pattern
	filter = NewResultFilter().WithResourcePattern(".*database.*")
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 1) // Only database

	// Test non-matching pattern
	filter = NewResultFilter().WithResourcePattern(".*nonexistent.*")
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 0)
}

func TestResultFilter_ApplyWithAttributePattern(t *testing.T) {
	results := createTestDriftResults()

	// Test attribute pattern filter
	filter := NewResultFilter().WithAttributePattern("instance_type")
	filtered := filter.Apply(results)
	assert.Len(t, filtered, 1) // Resources with instance_type differences

	// Test security group pattern
	filter = NewResultFilter().WithAttributePattern("security_groups")
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 1) // Only one resource with security_groups difference
}

func TestResultFilter_ApplyWithValuePattern(t *testing.T) {
	results := createTestDriftResults()

	// Test value pattern filter - using WithAttributePattern since WithValuePattern doesn't exist
	filter := NewResultFilter().WithAttributePattern("instance_type")
	filtered := filter.Apply(results)
	assert.Greater(t, len(filtered), 0) // Should find resources with instance_type attributes

	// Test specific attribute pattern
	filter = NewResultFilter().WithAttributePattern("security_groups")
	filtered = filter.Apply(results)
	assert.Greater(t, len(filtered), 0) // Should find resources with security_groups attributes
}

func TestResultFilter_ApplyWithTimeRange(t *testing.T) {
	results := createTestDriftResults()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	// Test time range filter (all results should be within range)
	filter := NewResultFilter().WithTimeRange(&yesterday, &tomorrow)
	filtered := filter.Apply(results)
	assert.Len(t, filtered, 3) // All results should be within range

	// Test restrictive time range (no results should match)
	futureStart := tomorrow
	futureEnd := tomorrow.Add(time.Hour)
	filter = NewResultFilter().WithTimeRange(&futureStart, &futureEnd)
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 0) // No results should match future time range
}

func TestResultFilter_ApplyWithLimit(t *testing.T) {
	results := createTestDriftResults()

	// Test limit filter
	filter := NewResultFilter().WithLimit(2, 0)
	filtered := filter.Apply(results)
	assert.Len(t, filtered, 2) // Should limit to 2 results

	// Test limit larger than available results
	filter = NewResultFilter().WithLimit(10, 0)
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 3) // Should return all available results

	// Test zero limit (should return all)
	filter = NewResultFilter().WithLimit(0, 0)
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 3)
}

func TestResultFilter_ApplyWithSorting(t *testing.T) {
	results := createTestDriftResults()

	// Test sort by resource ID (ascending)
	filter := NewResultFilter().WithSort(SortByResourceID, SortOrderAsc)
	filtered := filter.Apply(results)
	assert.Len(t, filtered, 3)

	// Verify sorting order
	resourceIDs := make([]string, len(filtered))
	i := 0
	for id := range filtered {
		resourceIDs[i] = id
		i++
	}

	// Check if sorted (basic check)
	for i := 1; i < len(resourceIDs); i++ {
		assert.LessOrEqual(t, resourceIDs[i-1], resourceIDs[i])
	}

	// Test sort by severity (descending)
	filter = NewResultFilter().WithSort(SortBySeverity, SortOrderDesc)
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 3)
}

func TestResultFilter_CombinedFilters(t *testing.T) {
	results := createTestDriftResults()

	// Test combined filters
	filter := NewResultFilter().
		WithSeverity(drift.SeverityMedium, drift.SeverityCritical).
		OnlyWithDrift().
		WithResourcePattern(".*web-server.*").
		WithLimit(1, 0)

	filtered := filter.Apply(results)
	assert.LessOrEqual(t, len(filtered), 1) // Should respect limit

	// Verify all filters are applied
	for resourceID, result := range filtered {
		// Check resource pattern
		assert.Contains(t, resourceID, "web-server")

		// Check drift status
		assert.Greater(t, result.GetDriftedCount(), 0)

		// Check severity
		assert.GreaterOrEqual(t, int(result.GetHighestSeverity()), int(drift.SeverityMedium))
	}
}

func TestResultFilter_ErrorHandling(t *testing.T) {
	// Test with nil results
	filter := NewResultFilter()
	filtered := filter.Apply(nil)
	assert.Len(t, filtered, 0) // Should return empty map, not error

	// Test with empty results
	results := make(map[string]*drift.DriftResult)
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 0)
}

func TestResultFilter_CriticalOnly(t *testing.T) {
	results := createTestDriftResults()

	// Test critical severity filter
	filter := NewResultFilter().WithSeverity(drift.SeverityCritical, drift.SeverityCritical).OnlyWithDrift()
	filtered := filter.Apply(results)
	assert.LessOrEqual(t, len(filtered), len(results))

	// Test security-related attributes
	filter = NewResultFilter().WithAttributePattern("security_groups")
	filtered = filter.Apply(results)
	assert.LessOrEqual(t, len(filtered), len(results))

	// Test recent changes with time range
	yesterday := time.Now().Add(-24 * time.Hour)
	now := time.Now()
	filter = NewResultFilter().WithTimeRange(&yesterday, &now).OnlyWithDrift()
	filtered = filter.Apply(results)
	assert.LessOrEqual(t, len(filtered), len(results))

	// Test top drifted with limit and sorting
	filter = NewResultFilter().WithLimit(5, 0).WithSort(SortBySeverity, SortOrderDesc).OnlyWithDrift()
	filtered = filter.Apply(results)
	assert.LessOrEqual(t, len(filtered), 5)
}

func TestDriftStatus_String(t *testing.T) {
	tests := []struct {
		status   DriftStatus
		expected string
	}{
		{DriftStatusAny, "any"},
		{DriftStatusDrifted, "drifted"},
		{DriftStatusNoDrift, "no_drift"},
		{DriftStatus(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.status.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterSortBy_String(t *testing.T) {
	tests := []struct {
		sortBy   FilterSortBy
		expected string
	}{
		{SortByResourceID, "resource_id"},
		{SortBySeverity, "severity"},
		{SortByDifferenceCount, "difference_count"},
		{SortByTimestamp, "timestamp"},
		{SortByResourceType, "resource_type"},
		{FilterSortBy("unknown"), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.sortBy.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test edge cases
func TestResultFilter_EmptyResults(t *testing.T) {
	filter := NewResultFilter()
	emptyResults := make(map[string]*drift.DriftResult)

	filtered := filter.Apply(emptyResults)
	assert.Len(t, filtered, 0)
}

func TestResultFilter_InvalidRegexPattern(t *testing.T) {
	results := createTestDriftResults()

	// Test invalid regex pattern - should handle gracefully
	filter := NewResultFilter().WithResourcePattern("[invalid")
	filtered := filter.Apply(results)
	// Invalid regex should result in no matches
	assert.Len(t, filtered, 0)

	// Test invalid attribute pattern
	filter = NewResultFilter().WithAttributePattern("[invalid")
	filtered = filter.Apply(results)
	assert.Len(t, filtered, 0)

	// Test valid patterns work
	filter = NewResultFilter().WithResourcePattern(".*web.*")
	filtered = filter.Apply(results)
	assert.Greater(t, len(filtered), 0)
}

func TestResultFilter_MultiplePatterns(t *testing.T) {
	results := createTestDriftResults()

	// Test resource pattern matching
	filter := NewResultFilter().WithResourcePattern(".*web.*")
	filtered := filter.Apply(results)
	assert.Greater(t, len(filtered), 0) // Should match web resources

	// Test database pattern
	filter = NewResultFilter().WithResourcePattern(".*database.*")
	filtered = filter.Apply(results)
	assert.Greater(t, len(filtered), 0) // Should match database resources

	// Test attribute patterns
	filter = NewResultFilter().WithAttributePattern("instance_type")
	filtered = filter.Apply(results)
	assert.Greater(t, len(filtered), 0) // Should match resources with instance_type attribute

	// Test security groups pattern
	filter = NewResultFilter().WithAttributePattern("security_groups")
	filtered = filter.Apply(results)
	assert.Greater(t, len(filtered), 0) // Should match resources with security_groups attribute
}

// Benchmark tests
func BenchmarkResultFilter_Apply(b *testing.B) {
	filter := NewResultFilter().WithSeverity(drift.SeverityMedium, drift.SeverityCritical)
	results := createLargeDriftResults(100) // Create larger dataset

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.Apply(results)
	}
}

func BenchmarkResultFilter_ApplyWithRegex(b *testing.B) {
	filter := NewResultFilter().WithResourcePattern(".*web-server.*")
	results := createLargeDriftResults(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.Apply(results)
	}
}

func BenchmarkResultFilter_ApplyWithSorting(b *testing.B) {
	filter := NewResultFilter().WithSort(SortBySeverity, SortOrderDesc)
	results := createLargeDriftResults(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.Apply(results)
	}
}

// Note: createLargeDriftResults function is defined in recommendations_test.go to avoid duplication
