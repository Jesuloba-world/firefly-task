package report

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"firefly-task/drift"
)

// FilterCriteria defines criteria for filtering drift results
type FilterCriteria struct {
	// Severity filtering
	MinSeverity    drift.DriftSeverity
	MaxSeverity    drift.DriftSeverity
	SeverityLevels []drift.DriftSeverity

	// Resource filtering
	ResourceIDs     []string
	ResourceTypes   []string
	InstanceIDs     []string
	ResourcePattern *regexp.Regexp

	// Attribute filtering
	AttributeNames    []string
	AttributePattern  *regexp.Regexp
	ExcludeAttributes []string

	// Time filtering
	After  *time.Time
	Before *time.Time

	// Drift status filtering
	OnlyWithDrift    bool
	OnlyWithoutDrift bool

	// Value filtering
	ExpectedValuePattern *regexp.Regexp
	ActualValuePattern   *regexp.Regexp

	// Limit and pagination
	Limit  int
	Offset int

	// Sorting
	SortBy    FilterSortBy
	SortOrder FilterSortOrder
}

// FilterSortBy defines sorting options
type FilterSortBy string

const (
	SortByResourceID      FilterSortBy = "resource_id"
	SortBySeverity        FilterSortBy = "severity"
	SortByTimestamp       FilterSortBy = "timestamp"
	SortByDifferenceCount FilterSortBy = "difference_count"
	SortByResourceType    FilterSortBy = "resource_type"
)

// String returns the string representation of FilterSortBy
func (fsb FilterSortBy) String() string {
	switch fsb {
	case SortByResourceID:
		return "resource_id"
	case SortBySeverity:
		return "severity"
	case SortByTimestamp:
		return "timestamp"
	case SortByDifferenceCount:
		return "difference_count"
	case SortByResourceType:
		return "resource_type"
	default:
		return "unknown"
	}
}

// FilterSortOrder defines sort order
type FilterSortOrder string

const (
	SortOrderAsc  FilterSortOrder = "asc"
	SortOrderDesc FilterSortOrder = "desc"
)

// DriftStatus defines drift status filtering options
type DriftStatus int

const (
	DriftStatusAny DriftStatus = iota
	DriftStatusDrifted
	DriftStatusNoDrift
)

// String returns the string representation of DriftStatus
func (ds DriftStatus) String() string {
	switch ds {
	case DriftStatusAny:
		return "any"
	case DriftStatusDrifted:
		return "drifted"
	case DriftStatusNoDrift:
		return "no_drift"
	default:
		return "unknown"
	}
}

// NewFilterCriteria creates a new FilterCriteria with default values
func NewFilterCriteria() *FilterCriteria {
	return &FilterCriteria{
		MinSeverity: drift.SeverityLow,
		MaxSeverity: drift.SeverityCritical,
		SortBy:      SortByResourceID,
		SortOrder:   SortOrderAsc,
	}
}

// WithMinSeverity sets the minimum severity level
func (fc *FilterCriteria) WithMinSeverity(severity drift.DriftSeverity) *FilterCriteria {
	fc.MinSeverity = severity
	return fc
}

// WithResourcePattern sets the resource pattern
func (fc *FilterCriteria) WithResourcePattern(pattern string) *FilterCriteria {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// A regex that will never match
		compiled = regexp.MustCompile(`\b\B`)
	}
	fc.ResourcePattern = compiled
	return fc
}

// WithAttributePattern sets the attribute pattern
func (fc *FilterCriteria) WithAttributePattern(pattern string) *FilterCriteria {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// A regex that will never match
		compiled = regexp.MustCompile(`\b\B`)
	}
	fc.AttributePattern = compiled
	return fc
}

// WithSince sets the after time filter
func (fc *FilterCriteria) WithSince(since time.Time) *FilterCriteria {
	fc.After = &since
	return fc
}

// WithUntil sets the before time filter
func (fc *FilterCriteria) WithUntil(until time.Time) *FilterCriteria {
	fc.Before = &until
	return fc
}

// WithDriftStatus sets the drift status filter
func (fc *FilterCriteria) WithDriftStatus(status DriftStatus) *FilterCriteria {
	switch status {
	case DriftStatusDrifted:
		fc.OnlyWithDrift = true
		fc.OnlyWithoutDrift = false
	case DriftStatusNoDrift:
		fc.OnlyWithoutDrift = true
		fc.OnlyWithDrift = false
	default:
		fc.OnlyWithDrift = false
		fc.OnlyWithoutDrift = false
	}
	return fc
}

// WithValuePattern sets the value pattern filter
func (fc *FilterCriteria) WithValuePattern(pattern string) *FilterCriteria {
	if compiled, err := regexp.Compile(pattern); err == nil {
		fc.ActualValuePattern = compiled
	}
	return fc
}

// WithLimit sets the limit
func (fc *FilterCriteria) WithLimit(limit int) *FilterCriteria {
	fc.Limit = limit
	return fc
}

// WithSort sets the sort criteria
func (fc *FilterCriteria) WithSort(sortBy FilterSortBy, desc bool) *FilterCriteria {
	fc.SortBy = sortBy
	if desc {
		fc.SortOrder = SortOrderDesc
	} else {
		fc.SortOrder = SortOrderAsc
	}
	return fc
}

// ResultFilter provides filtering capabilities for drift results
type ResultFilter struct {
	criteria *FilterCriteria
}

// NewResultFilter creates a new result filter
func NewResultFilter() *ResultFilter {
	return &ResultFilter{
		criteria: &FilterCriteria{
			MinSeverity: drift.SeverityNone,
			MaxSeverity: drift.SeverityCritical,
			SortBy:      SortByResourceID,
			SortOrder:   SortOrderAsc,
		},
	}
}

// WithSeverity sets severity filtering
func (rf *ResultFilter) WithSeverity(min, max drift.DriftSeverity) *ResultFilter {
	rf.criteria.MinSeverity = min
	rf.criteria.MaxSeverity = max
	return rf
}

// WithSeverityLevels sets specific severity levels to include
func (rf *ResultFilter) WithSeverityLevels(levels ...drift.DriftSeverity) *ResultFilter {
	rf.criteria.SeverityLevels = levels
	return rf
}

// WithResourceIDs filters by specific resource IDs
func (rf *ResultFilter) WithResourceIDs(resourceIDs ...string) *ResultFilter {
	rf.criteria.ResourceIDs = resourceIDs
	return rf
}

// WithResourceTypes filters by resource types
func (rf *ResultFilter) WithResourceTypes(resourceTypes ...string) *ResultFilter {
	rf.criteria.ResourceTypes = resourceTypes
	return rf
}

// WithResourcePattern filters by resource ID pattern
func (rf *ResultFilter) WithResourcePattern(pattern string) *ResultFilter {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// A regex that will never match
		compiled = regexp.MustCompile(`\b\B`)
	}
	rf.criteria.ResourcePattern = compiled
	return rf
}

// WithAttributeNames filters by specific attribute names
func (rf *ResultFilter) WithAttributeNames(attributeNames ...string) *ResultFilter {
	rf.criteria.AttributeNames = attributeNames
	return rf
}

// WithAttributePattern filters by attribute name pattern
func (rf *ResultFilter) WithAttributePattern(pattern string) *ResultFilter {
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		// A regex that will never match
		compiled = regexp.MustCompile(`\b\B`)
	}
	rf.criteria.AttributePattern = compiled
	return rf
}

// ExcludeAttributes excludes specific attributes from results
func (rf *ResultFilter) ExcludeAttributes(attributeNames ...string) *ResultFilter {
	rf.criteria.ExcludeAttributes = attributeNames
	return rf
}

// WithTimeRange filters by time range
func (rf *ResultFilter) WithTimeRange(after, before *time.Time) *ResultFilter {
	rf.criteria.After = after
	rf.criteria.Before = before
	return rf
}

// OnlyWithDrift includes only resources with drift
func (rf *ResultFilter) OnlyWithDrift() *ResultFilter {
	rf.criteria.OnlyWithDrift = true
	rf.criteria.OnlyWithoutDrift = false
	return rf
}

// OnlyWithoutDrift includes only resources without drift
func (rf *ResultFilter) OnlyWithoutDrift() *ResultFilter {
	rf.criteria.OnlyWithoutDrift = true
	rf.criteria.OnlyWithDrift = false
	return rf
}

// WithLimit sets result limit and offset for pagination
func (rf *ResultFilter) WithLimit(limit, offset int) *ResultFilter {
	rf.criteria.Limit = limit
	rf.criteria.Offset = offset
	return rf
}

// WithSort sets sorting criteria
func (rf *ResultFilter) WithSort(sortBy FilterSortBy, order FilterSortOrder) *ResultFilter {
	rf.criteria.SortBy = sortBy
	rf.criteria.SortOrder = order
	return rf
}

// Apply applies the filter to drift results
func (rf *ResultFilter) Apply(results map[string]*drift.DriftResult) []*drift.DriftResult {
	if results == nil {
		return nil
	}

	// Convert to slice for easier processing
	var resultList []*drift.DriftResult
	for _, result := range results {
		if rf.matchesResourceCriteria(result) {
			filteredResult := rf.filterDifferences(result)
			if filteredResult != nil {
				resultList = append(resultList, filteredResult)
			}
		}
	}

	// Sort results
	rf.sortResults(resultList)

	// Apply pagination
	if rf.criteria.Limit > 0 || rf.criteria.Offset > 0 {
		resultList = rf.paginateResults(resultList)
	}

	return resultList
}

// matchesResourceCriteria checks if a result matches resource-level criteria
func (rf *ResultFilter) matchesResourceCriteria(result *drift.DriftResult) bool {
	// Check drift status
	if rf.criteria.OnlyWithDrift && !result.HasDrift {
		return false
	}
	if rf.criteria.OnlyWithoutDrift && result.HasDrift {
		return false
	}

	// Check severity
	if !rf.matchesSeverity(result.OverallSeverity) {
		return false
	}

	// Check resource ID
	if len(rf.criteria.ResourceIDs) > 0 {
		found := false
		for _, id := range rf.criteria.ResourceIDs {
			if result.ResourceID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check resource pattern
	if rf.criteria.ResourcePattern != nil {
		if !rf.criteria.ResourcePattern.MatchString(result.ResourceID) {
			return false
		}
	}

	// Check instance IDs
	if len(rf.criteria.InstanceIDs) > 0 {
		found := false
		for _, id := range rf.criteria.InstanceIDs {
			if result.InstanceID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check time range
	if rf.criteria.After != nil && result.Timestamp.Before(*rf.criteria.After) {
		return false
	}
	if rf.criteria.Before != nil && result.Timestamp.After(*rf.criteria.Before) {
		return false
	}

	return true
}

// matchesSeverity checks if severity matches criteria
func (rf *ResultFilter) matchesSeverity(severity drift.DriftSeverity) bool {
	// Check specific severity levels
	if len(rf.criteria.SeverityLevels) > 0 {
		found := false
		for _, level := range rf.criteria.SeverityLevels {
			if severity == level {
				found = true
				break
			}
		}
		return found
	}

	// Check severity range
	return severity >= rf.criteria.MinSeverity && severity <= rf.criteria.MaxSeverity
}

// filterDifferences filters differences within a result
func (rf *ResultFilter) filterDifferences(result *drift.DriftResult) *drift.DriftResult {
	if result == nil {
		return nil
	}

	// Create a copy of the result
	filteredResult := &drift.DriftResult{
		ResourceID:      result.ResourceID,
		InstanceID:      result.InstanceID,
		Timestamp:       result.Timestamp,
		OverallSeverity: result.OverallSeverity,
		HasDrift:        result.HasDrift,
		Differences:     []drift.AttributeDifference{},
	}

	// Filter differences
	for _, diff := range result.Differences {
		if rf.matchesDifferenceCriteria(diff) {
			filteredResult.Differences = append(filteredResult.Differences, diff)
		}
	}

	// Check if attribute-related filters are applied and no differences match
	hasAttributeFilters := rf.criteria.AttributePattern != nil || 
		len(rf.criteria.AttributeNames) > 0 || 
		len(rf.criteria.ExcludeAttributes) > 0 ||
		rf.criteria.ExpectedValuePattern != nil ||
		rf.criteria.ActualValuePattern != nil
	
	if hasAttributeFilters && len(filteredResult.Differences) == 0 {
		return nil
	}

	// Update drift status based on filtered differences
	filteredResult.HasDrift = len(filteredResult.Differences) > 0
	if filteredResult.HasDrift {
		// Recalculate overall severity
		filteredResult.OverallSeverity = rf.calculateOverallSeverity(filteredResult.Differences)
	} else {
		filteredResult.OverallSeverity = drift.SeverityNone
	}

	// Check if filtered result still matches criteria
	if rf.criteria.OnlyWithDrift && !filteredResult.HasDrift {
		return nil
	}
	if rf.criteria.OnlyWithoutDrift && filteredResult.HasDrift {
		return nil
	}

	return filteredResult
}

// matchesDifferenceCriteria checks if a difference matches criteria
func (rf *ResultFilter) matchesDifferenceCriteria(diff drift.AttributeDifference) bool {
	// Check attribute names
	if len(rf.criteria.AttributeNames) > 0 {
		found := false
		for _, name := range rf.criteria.AttributeNames {
			if diff.AttributeName == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check excluded attributes
	for _, excluded := range rf.criteria.ExcludeAttributes {
		if diff.AttributeName == excluded {
			return false
		}
	}

	// Check attribute pattern
	if rf.criteria.AttributePattern != nil {
		if !rf.criteria.AttributePattern.MatchString(diff.AttributeName) {
			return false
		}
	}

	// Check value patterns
	if rf.criteria.ExpectedValuePattern != nil {
		if expectedStr, ok := diff.ExpectedValue.(string); ok {
			if !rf.criteria.ExpectedValuePattern.MatchString(strings.ToLower(expectedStr)) {
				return false
			}
		}
	}

	if rf.criteria.ActualValuePattern != nil {
		if actualStr, ok := diff.ActualValue.(string); ok {
			if !rf.criteria.ActualValuePattern.MatchString(strings.ToLower(actualStr)) {
				return false
			}
		}
	}

	return true
}

// calculateOverallSeverity calculates overall severity from differences
func (rf *ResultFilter) calculateOverallSeverity(differences []drift.AttributeDifference) drift.DriftSeverity {
	if len(differences) == 0 {
		return drift.SeverityNone
	}

	maxSeverity := drift.SeverityNone
	for _, diff := range differences {
		if diff.Severity > maxSeverity {
			maxSeverity = diff.Severity
		}
	}

	return maxSeverity
}

// sortResults sorts results based on criteria
func (rf *ResultFilter) sortResults(results []*drift.DriftResult) {
	sort.Slice(results, func(i, j int) bool {
		var less bool
		switch rf.criteria.SortBy {
		case SortByResourceID:
			less = results[i].ResourceID < results[j].ResourceID
		case SortBySeverity:
			less = results[i].OverallSeverity < results[j].OverallSeverity
		case SortByTimestamp:
			less = results[i].Timestamp.Before(results[j].Timestamp)
		case SortByDifferenceCount:
			less = len(results[i].Differences) < len(results[j].Differences)
		case SortByResourceType:
			// Extract resource type from resource ID (assuming format like "aws_instance.example")
			typeI := rf.extractResourceType(results[i].ResourceID)
			typeJ := rf.extractResourceType(results[j].ResourceID)
			less = typeI < typeJ
		default:
			less = results[i].ResourceID < results[j].ResourceID
		}

		if rf.criteria.SortOrder == SortOrderDesc {
			return !less
		}
		return less
	})
}

// extractResourceType extracts resource type from resource ID
func (rf *ResultFilter) extractResourceType(resourceID string) string {
	parts := strings.Split(resourceID, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return resourceID
}

// paginateResults applies pagination to results
func (rf *ResultFilter) paginateResults(results []*drift.DriftResult) []*drift.DriftResult {
	start := rf.criteria.Offset
	if start >= len(results) {
		return []*drift.DriftResult{}
	}

	end := len(results)
	if rf.criteria.Limit > 0 {
		end = start + rf.criteria.Limit
		if end > len(results) {
			end = len(results)
		}
	}

	return results[start:end]
}

// GetFilterSummary returns a summary of applied filters
func (rf *ResultFilter) GetFilterSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	if rf.criteria.MinSeverity != drift.SeverityNone || rf.criteria.MaxSeverity != drift.SeverityCritical {
		summary["severity_range"] = map[string]string{
			"min": rf.criteria.MinSeverity.String(),
			"max": rf.criteria.MaxSeverity.String(),
		}
	}

	if len(rf.criteria.SeverityLevels) > 0 {
		levels := make([]string, len(rf.criteria.SeverityLevels))
		for i, level := range rf.criteria.SeverityLevels {
			levels[i] = level.String()
		}
		summary["severity_levels"] = levels
	}

	if len(rf.criteria.ResourceIDs) > 0 {
		summary["resource_ids"] = rf.criteria.ResourceIDs
	}

	if len(rf.criteria.AttributeNames) > 0 {
		summary["attribute_names"] = rf.criteria.AttributeNames
	}

	if len(rf.criteria.ExcludeAttributes) > 0 {
		summary["excluded_attributes"] = rf.criteria.ExcludeAttributes
	}

	if rf.criteria.OnlyWithDrift {
		summary["drift_filter"] = "only_with_drift"
	} else if rf.criteria.OnlyWithoutDrift {
		summary["drift_filter"] = "only_without_drift"
	}

	if rf.criteria.Limit > 0 {
		summary["pagination"] = map[string]int{
			"limit":  rf.criteria.Limit,
			"offset": rf.criteria.Offset,
		}
	}

	summary["sort"] = map[string]string{
		"by":    string(rf.criteria.SortBy),
		"order": string(rf.criteria.SortOrder),
	}

	return summary
}

// PresetFilters provides common filter presets
type PresetFilters struct{}

// NewPresetFilters creates preset filter configurations
func NewPresetFilters() *PresetFilters {
	return &PresetFilters{}
}

// CriticalOnly returns filter for critical severity only
func (pf *PresetFilters) CriticalOnly() *ResultFilter {
	return NewResultFilter().WithSeverityLevels(drift.SeverityCritical)
}

// HighAndCritical returns filter for high and critical severity
func (pf *PresetFilters) HighAndCritical() *ResultFilter {
	return NewResultFilter().WithSeverityLevels(drift.SeverityCritical, drift.SeverityHigh)
}

// SecurityRelated returns filter for security-related attributes
func (pf *PresetFilters) SecurityRelated() *ResultFilter {
	return NewResultFilter().WithAttributePattern("(?i)(security|iam|policy|permission|role|key|secret|password|token)")
}

// NetworkRelated returns filter for network-related attributes
func (pf *PresetFilters) NetworkRelated() *ResultFilter {
	return NewResultFilter().WithAttributePattern("(?i)(vpc|subnet|security_group|route|gateway|ip|port|cidr)")
}

// RecentChanges returns filter for recent changes (last 24 hours)
func (pf *PresetFilters) RecentChanges() *ResultFilter {
	after := time.Now().Add(-24 * time.Hour)
	return NewResultFilter().WithTimeRange(&after, nil)
}

// EC2Instances returns filter for EC2 instance resources
func (pf *PresetFilters) EC2Instances() *ResultFilter {
	return NewResultFilter().WithResourcePattern("^aws_instance\\.")
}

// TopDrifted returns filter for resources with most differences (top 10)
func (pf *PresetFilters) TopDrifted() *ResultFilter {
	return NewResultFilter().
		OnlyWithDrift().
		WithSort(SortByDifferenceCount, SortOrderDesc).
		WithLimit(10, 0)
}