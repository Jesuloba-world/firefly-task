package report

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"firefly-task/drift"
)

// RecommendationEngine generates actionable recommendations based on drift results
type RecommendationEngine struct {
	config *RecommendationConfig
}

// RecommendationConfig configures the recommendation engine
type RecommendationConfig struct {
	// Include settings
	IncludeActions     bool
	IncludePriority    bool
	IncludeTimeEstimate bool

	// Grouping settings
	GroupByResourceType bool
	GroupBySeverity     bool

	// Filtering
	MinSeverity        drift.DriftSeverity
	MaxRecommendations int
}

// RecommendationType defines types of recommendations
type RecommendationType int

const (
	RecommendationTypeUpdate RecommendationType = iota
	RecommendationTypeReview
	RecommendationTypeInvestigate
	RecommendationTypeIgnore
	RecommendationTypeAlert
)

// String returns the string representation of RecommendationType
func (rt RecommendationType) String() string {
	switch rt {
	case RecommendationTypeUpdate:
		return "update"
	case RecommendationTypeReview:
		return "review"
	case RecommendationTypeInvestigate:
		return "investigate"
	case RecommendationTypeIgnore:
		return "ignore"
	case RecommendationTypeAlert:
		return "alert"
	default:
		return "unknown"
	}
}

// RecommendationPriority defines priority levels
type RecommendationPriority int

const (
	PriorityCritical RecommendationPriority = iota
	PriorityHigh
	PriorityMedium
	PriorityLow
)

// String returns the string representation of RecommendationPriority
func (rp RecommendationPriority) String() string {
	switch rp {
	case PriorityCritical:
		return "critical"
	case PriorityHigh:
		return "high"
	case PriorityMedium:
		return "medium"
	case PriorityLow:
		return "low"
	default:
		return "unknown"
	}
}

// ActionableRecommendation represents a specific recommendation
type ActionableRecommendation struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Type        RecommendationType     `json:"type"`
	Priority    RecommendationPriority `json:"priority"`
	Severity    drift.DriftSeverity    `json:"severity"`

	// Affected resources
	AffectedResources []string `json:"affected_resources"`
	ResourceCount     int      `json:"resource_count"`

	// Actions
	Commands         []string `json:"commands,omitempty"`
	ManualSteps      []string `json:"manual_steps,omitempty"`
	DocumentationURL string   `json:"documentation_url,omitempty"`

	// Metadata
	EstimatedTime string    `json:"estimated_time,omitempty"`
	RiskLevel     string    `json:"risk_level,omitempty"`
	Category      string    `json:"category"`
	Tags          []string  `json:"tags,omitempty"`
	CreatedAt     time.Time `json:"created_at"`

	// Dependencies
	DependsOn []string `json:"depends_on,omitempty"`
	Blocks    []string `json:"blocks,omitempty"`
}

// RecommendationSummary provides an overview of recommendations
type RecommendationSummary struct {
	TotalRecommendations int                            `json:"total_recommendations"`
	ByPriority           map[RecommendationPriority]int `json:"by_priority"`
	ByType               map[RecommendationType]int     `json:"by_type"`
	BySeverity           map[drift.DriftSeverity]int    `json:"by_severity"`
	ByCategory           map[string]int                 `json:"by_category"`
	EstimatedTotalTime   string                         `json:"estimated_total_time"`
	HighestPriority      RecommendationPriority         `json:"highest_priority"`
	MostCommonCategory   string                         `json:"most_common_category"`
	Recommendations      []ActionableRecommendation     `json:"recommendations"`
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine() *RecommendationEngine {
	return &RecommendationEngine{
		config: NewRecommendationConfig(),
	}
}

// NewRecommendationConfig creates a new recommendation configuration with default values
func NewRecommendationConfig() *RecommendationConfig {
	return &RecommendationConfig{
		IncludeActions:      true,
		IncludePriority:     true,
		IncludeTimeEstimate: true,
		GroupBySeverity:     true,
		GroupByResourceType: true,
		MaxRecommendations:  10,
		MinSeverity:         drift.SeverityLow,
	}
}

// WithConfig applies configuration to the engine
func (re *RecommendationEngine) WithConfig(config *RecommendationConfig) *RecommendationEngine {
	re.config = config
	return re
}

// GetConfig returns the current recommendation configuration
func (re *RecommendationEngine) GetConfig() *RecommendationConfig {
	return re.config
}

// WithActions sets whether to include actions in recommendations
func (rc *RecommendationConfig) WithActions(include bool) *RecommendationConfig {
	rc.IncludeActions = include
	return rc
}

// WithPriority sets whether to include priority in recommendations
func (rc *RecommendationConfig) WithPriority(include bool) *RecommendationConfig {
	rc.IncludePriority = include
	return rc
}

// WithTimeEstimate sets whether to include time estimates
func (rc *RecommendationConfig) WithTimeEstimate(include bool) *RecommendationConfig {
	rc.IncludeTimeEstimate = include
	return rc
}

// WithGroupBySeverity sets whether to group by severity
func (rc *RecommendationConfig) WithGroupBySeverity(group bool) *RecommendationConfig {
	rc.GroupBySeverity = group
	return rc
}

// WithGroupByResourceType sets whether to group by resource type
func (rc *RecommendationConfig) WithGroupByResourceType(group bool) *RecommendationConfig {
	rc.GroupByResourceType = group
	return rc
}

// WithMaxRecommendations sets the maximum number of recommendations
func (rc *RecommendationConfig) WithMaxRecommendations(max int) *RecommendationConfig {
	rc.MaxRecommendations = max
	return rc
}

// WithMinSeverity sets the minimum severity for recommendations
func (rc *RecommendationConfig) WithMinSeverity(severity drift.DriftSeverity) *RecommendationConfig {
	rc.MinSeverity = severity
	return rc
}

// GenerateRecommendations generates actionable recommendations from drift results
func (re *RecommendationEngine) GenerateRecommendations(results map[string]*drift.DriftResult) (*RecommendationSummary, error) {
	log.Println("Starting recommendation generation")
	if results == nil {
		log.Println("Error: drift results are nil")
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}
	log.Printf("Received %d drift results", len(results))

	var recommendations []ActionableRecommendation

	// Analyze patterns
	patterns := re.AnalyzeDriftPatterns(results)
	log.Printf("Analyzed and found %d drift patterns", len(patterns))

	// Generate recommendations from patterns
	recommendations = re.generateFromPatterns(patterns, results)
	log.Printf("Generated %d initial recommendations", len(recommendations))

	// Filter recommendations by minimum severity
	originalCountBeforeFilter := len(recommendations)
	var filteredRecommendations []ActionableRecommendation
	for _, rec := range recommendations {
		if rec.Severity >= re.config.MinSeverity {
			filteredRecommendations = append(filteredRecommendations, rec)
		}
	}
	recommendations = filteredRecommendations
	log.Printf("Filtered recommendations from %d to %d (min severity: %s)", originalCountBeforeFilter, len(recommendations), re.config.MinSeverity)

	// Prioritize and sort recommendations
	re.prioritizeRecommendations(recommendations)
	log.Printf("Prioritized %d recommendations", len(recommendations))

	// Apply limits
	originalCountBeforeLimit := len(recommendations)
	if re.config.MaxRecommendations > 0 && len(recommendations) > re.config.MaxRecommendations {
		recommendations = recommendations[:re.config.MaxRecommendations]
		log.Printf("Limited recommendations from %d to %d (max: %d)", originalCountBeforeLimit, len(recommendations), re.config.MaxRecommendations)
	}

	// Generate summary
	summary := re.generateSummary(recommendations)
	log.Println("Generated summary")

	log.Println("Finished recommendation generation")
	return summary, nil
}

// DriftPattern represents a pattern in drift results
type DriftPattern struct {
	Type              string
	AffectedResources []string
	CommonAttributes  []string
	Severity          drift.DriftSeverity
	Frequency         int
	Category          string
}

// analyzeDriftPatterns identifies common patterns in drift results
func (re *RecommendationEngine) AnalyzeDriftPatterns(results map[string]*drift.DriftResult) []DriftPattern {
	var patterns []DriftPattern

	// Pattern 1: Resources with same attribute drifts
	attributeGroups := make(map[string][]string)
	for resourceID, result := range results {
		if !result.HasDrift {
			continue
		}
		for _, diff := range result.Differences {
			attributeGroups[diff.AttributeName] = append(attributeGroups[diff.AttributeName], resourceID)
		}
	}

	for attrName, resources := range attributeGroups {
		if len(resources) > 1 {
			patterns = append(patterns, DriftPattern{
				Type:              "common_attribute_drift",
				AffectedResources: resources,
				CommonAttributes:  []string{attrName},
				Frequency:         len(resources),
				Category:          re.categorizeAttribute(attrName),
			})
		}
	}

	// Pattern 2: Resources of same type with drift
	typeGroups := make(map[string][]string)
	for resourceID, result := range results {
		if !result.HasDrift {
			continue
		}
		resourceType := re.extractResourceType(resourceID)
		typeGroups[resourceType] = append(typeGroups[resourceType], resourceID)
	}

	for resourceType, resources := range typeGroups {
		if len(resources) > 1 {
			patterns = append(patterns, DriftPattern{
				Type:              "resource_type_drift",
				AffectedResources: resources,
				CommonAttributes:  []string{"resource_type"},
				Frequency:         len(resources),
				Category:          re.categorizeResourceType(resourceType),
			})
		}
	}

	// Pattern 3: High severity drifts
	var highSeverityResources []string
	for resourceID, result := range results {
		if result.HasDrift && (result.OverallSeverity == drift.SeverityCritical || result.OverallSeverity == drift.SeverityHigh) {
			highSeverityResources = append(highSeverityResources, resourceID)
		}
	}

	if len(highSeverityResources) > 0 {
		patterns = append(patterns, DriftPattern{
			Type:              "high_severity_drift",
			AffectedResources: highSeverityResources,
			CommonAttributes:  []string{"severity"},
			Severity:          drift.SeverityHigh,
			Frequency:         len(highSeverityResources),
			Category:          "critical",
		})
	}

	return patterns
}

// generateFromPatterns generates recommendations from identified patterns
func (re *RecommendationEngine) generateFromPatterns(patterns []DriftPattern, results map[string]*drift.DriftResult) []ActionableRecommendation {
	var recommendations []ActionableRecommendation

	for i, pattern := range patterns {
		switch pattern.Type {
		case "common_attribute_drift":
			recommendations = append(recommendations, re.generateAttributeDriftRecommendation(pattern, i))
		case "resource_type_drift":
			recommendations = append(recommendations, re.generateResourceTypeDriftRecommendation(pattern, i))
		case "high_severity_drift":
			recommendations = append(recommendations, re.generateHighSeverityRecommendation(pattern, i))
		}
	}

	// Generate individual resource recommendations
	for resourceID, result := range results {
		if result.HasDrift {
			recommendations = append(recommendations, re.generateResourceSpecificRecommendation(resourceID, result))
		}
	}

	return recommendations
}

// generateAttributeDriftRecommendation creates recommendation for common attribute drift
func (re *RecommendationEngine) generateAttributeDriftRecommendation(pattern DriftPattern, index int) ActionableRecommendation {
	attrName := pattern.CommonAttributes[0]

	return ActionableRecommendation{
		ID:                fmt.Sprintf("attr-drift-%d", index),
		Title:             fmt.Sprintf("Fix %s attribute drift across %d resources", attrName, len(pattern.AffectedResources)),
		Description:       fmt.Sprintf("Multiple resources have drift in the '%s' attribute. This suggests a systematic configuration issue.", attrName),
		Type:              RecommendationTypeReview,
		Priority:          re.determinePriorityFromFrequency(pattern.Frequency),
		Severity:          drift.SeverityMedium,
		AffectedResources: pattern.AffectedResources,
		ResourceCount:     len(pattern.AffectedResources),
		Commands: []string{
			"terraform plan",
			"terraform apply",
		},
		ManualSteps: []string{
			fmt.Sprintf("Review the '%s' attribute configuration in your Terraform files", attrName),
			"Check if this attribute should be consistent across resources",
			"Update the configuration to match the desired state",
			"Apply changes using terraform apply",
		},
		EstimatedTime: re.estimateTimeForResources(len(pattern.AffectedResources)),
		RiskLevel:     re.assessRiskLevel(pattern.Category, len(pattern.AffectedResources)),
		Category:      pattern.Category,
		Tags:          []string{"bulk-fix", "attribute-drift", attrName},
		CreatedAt:     time.Now(),
	}
}

// generateResourceTypeDriftRecommendation creates recommendation for resource type drift
func (re *RecommendationEngine) generateResourceTypeDriftRecommendation(pattern DriftPattern, index int) ActionableRecommendation {
	resourceType := re.extractResourceType(pattern.AffectedResources[0])

	return ActionableRecommendation{
		ID:                fmt.Sprintf("type-drift-%d", index),
		Title:             fmt.Sprintf("Address drift in %s resources", resourceType),
		Description:       fmt.Sprintf("%d %s resources have configuration drift. Consider reviewing the module or configuration template.", len(pattern.AffectedResources), resourceType),
		Type:              RecommendationTypeInvestigate,
		Priority:          re.determinePriorityFromFrequency(pattern.Frequency),
		Severity:          drift.SeverityMedium,
		AffectedResources: pattern.AffectedResources,
		ResourceCount:     len(pattern.AffectedResources),
		Commands: []string{
			fmt.Sprintf("terraform plan -target=%s", strings.Join(pattern.AffectedResources, " -target=")),
			"terraform apply",
		},
		ManualSteps: []string{
			fmt.Sprintf("Review the %s resource configuration", resourceType),
			"Check if there's a common module or template causing the drift",
			"Update the configuration to prevent future drift",
			"Consider using terraform modules for consistency",
		},
		EstimatedTime: re.estimateTimeForResources(len(pattern.AffectedResources)),
		RiskLevel:     re.assessRiskLevel(pattern.Category, len(pattern.AffectedResources)),
		Category:      pattern.Category,
		Tags:          []string{"resource-type", "investigation", resourceType},
		CreatedAt:     time.Now(),
	}
}

// generateHighSeverityRecommendation creates recommendation for high severity drift
func (re *RecommendationEngine) generateHighSeverityRecommendation(pattern DriftPattern, index int) ActionableRecommendation {
	return ActionableRecommendation{
		ID:                fmt.Sprintf("high-severity-%d", index),
		Title:             "URGENT: Address critical/high severity drift",
		Description:       fmt.Sprintf("%d resources have critical or high severity drift that requires immediate attention.", len(pattern.AffectedResources)),
		Type:              RecommendationTypeUpdate,
		Priority:          PriorityCritical,
		Severity:          drift.SeverityCritical,
		AffectedResources: pattern.AffectedResources,
		ResourceCount:     len(pattern.AffectedResources),
		Commands: []string{
			"terraform plan",
			"terraform apply",
		},
		ManualSteps: []string{
			"Immediately review the affected resources",
			"Assess the security and operational impact",
			"Create a backup or snapshot if necessary",
			"Apply fixes during the next maintenance window",
			"Monitor the resources after applying changes",
		},
		EstimatedTime: "1-2 hours",
		RiskLevel:     "high",
		Category:      "critical",
		Tags:          []string{"urgent", "high-severity", "immediate-action"},
		CreatedAt:     time.Now(),
	}
}

// generateResourceSpecificRecommendation creates recommendation for individual resource
func (re *RecommendationEngine) generateResourceSpecificRecommendation(resourceID string, result *drift.DriftResult) ActionableRecommendation {
	resourceType := re.extractResourceType(resourceID)

	return ActionableRecommendation{
		ID:                fmt.Sprintf("resource-%s", strings.ReplaceAll(resourceID, ".", "-")),
		Title:             fmt.Sprintf("Fix drift in %s", resourceID),
		Description:       fmt.Sprintf("Resource %s has %d configuration differences that need to be addressed.", resourceID, len(result.Differences)),
		Type:              re.determineRecommendationType(result.OverallSeverity),
		Priority:          re.determinePriorityFromSeverity(result.OverallSeverity),
		Severity:          result.OverallSeverity,
		AffectedResources: []string{resourceID},
		ResourceCount:     1,
		Commands: []string{
			fmt.Sprintf("terraform plan -target=%s", resourceID),
			fmt.Sprintf("terraform apply -target=%s", resourceID),
		},
		ManualSteps:   re.generateManualStepsForResource(resourceID, result),
		EstimatedTime: re.estimateTimeForSingleResource(result),
		RiskLevel:     re.assessRiskLevelForResource(resourceType, result.OverallSeverity),
		Category:      re.categorizeResourceType(resourceType),
		Tags:          []string{"single-resource", resourceType, result.OverallSeverity.String()},
		CreatedAt:     time.Now(),
	}
}

// generatePreventiveRecommendations creates recommendations to prevent future drift
func (re *RecommendationEngine) generatePreventiveRecommendations(results map[string]*drift.DriftResult) []ActionableRecommendation {
	var recommendations []ActionableRecommendation

	// Recommendation for regular drift detection
	recommendations = append(recommendations, ActionableRecommendation{
		ID:          "preventive-monitoring",
		Title:       "Implement regular drift detection",
		Description: "Set up automated drift detection to catch configuration changes early.",
		Type:        RecommendationTypeAlert,
		Priority:    PriorityMedium,
		Severity:    drift.SeverityLow,
		Commands: []string{
			"# Add to CI/CD pipeline:",
			"terraform plan -detailed-exitcode",
			"# Schedule regular drift checks",
		},
		ManualSteps: []string{
			"Set up a scheduled job to run drift detection daily",
			"Configure alerts for when drift is detected",
			"Document the drift detection process for your team",
			"Consider using Terraform Cloud or similar for automated checks",
		},
		EstimatedTime: "2-4 hours",
		RiskLevel:     "low",
		Category:      "automation",
		Tags:          []string{"preventive", "monitoring", "automation"},
		CreatedAt:     time.Now(),
	})

	// Recommendation for state locking
	recommendations = append(recommendations, ActionableRecommendation{
		ID:          "preventive-state-locking",
		Title:       "Implement Terraform state locking",
		Description: "Use remote state with locking to prevent concurrent modifications.",
		Type:        RecommendationTypeAlert,
		Priority:    PriorityHigh,
		Severity:    drift.SeverityMedium,
		ManualSteps: []string{
			"Configure remote state backend (S3, Azure Storage, etc.)",
			"Enable state locking (DynamoDB for S3, etc.)",
			"Update team procedures to use shared state",
			"Train team members on proper Terraform workflows",
		},
		EstimatedTime: "4-6 hours",
		RiskLevel:     "medium",
		Category:      "infrastructure",
		Tags:          []string{"preventive", "state-management", "collaboration"},
		CreatedAt:     time.Now(),
	})

	return recommendations
}

// Helper methods for recommendation generation

func (re *RecommendationEngine) extractResourceType(resourceID string) string {
	parts := strings.Split(resourceID, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return "unknown"
}

func (re *RecommendationEngine) categorizeAttribute(attrName string) string {
	lowerAttr := strings.ToLower(attrName)
	if strings.Contains(lowerAttr, "security") || strings.Contains(lowerAttr, "iam") || strings.Contains(lowerAttr, "policy") {
		return "security"
	}
	if strings.Contains(lowerAttr, "network") || strings.Contains(lowerAttr, "vpc") || strings.Contains(lowerAttr, "subnet") {
		return "networking"
	}
	if strings.Contains(lowerAttr, "storage") || strings.Contains(lowerAttr, "disk") || strings.Contains(lowerAttr, "volume") {
		return "storage"
	}
	return "configuration"
}

func (re *RecommendationEngine) categorizeResourceType(resourceType string) string {
	if strings.HasPrefix(resourceType, "aws_instance") || strings.HasPrefix(resourceType, "aws_ec2") {
		return "compute"
	}
	if strings.HasPrefix(resourceType, "aws_s3") || strings.Contains(resourceType, "storage") {
		return "storage"
	}
	if strings.Contains(resourceType, "vpc") || strings.Contains(resourceType, "subnet") || strings.Contains(resourceType, "security_group") {
		return "networking"
	}
	if strings.Contains(resourceType, "iam") || strings.Contains(resourceType, "policy") {
		return "security"
	}
	return "infrastructure"
}

func (re *RecommendationEngine) determinePriorityFromSeverity(severity drift.DriftSeverity) RecommendationPriority {
	switch severity {
	case drift.SeverityCritical:
		return PriorityCritical
	case drift.SeverityHigh:
		return PriorityHigh
	case drift.SeverityMedium:
		return PriorityMedium
	default:
		return PriorityLow
	}
}

func (re *RecommendationEngine) determinePriorityFromFrequency(frequency int) RecommendationPriority {
	if frequency >= 10 {
		return PriorityCritical
	}
	if frequency >= 5 {
		return PriorityHigh
	}
	if frequency >= 2 {
		return PriorityMedium
	}
	return PriorityLow
}

func (re *RecommendationEngine) determineRecommendationType(severity drift.DriftSeverity) RecommendationType {
	switch severity {
	case drift.SeverityCritical:
		return RecommendationTypeUpdate
	case drift.SeverityHigh:
		return RecommendationTypeReview
	case drift.SeverityMedium:
		return RecommendationTypeReview
	default:
		return RecommendationTypeIgnore
	}
}

func (re *RecommendationEngine) estimateTimeForResources(count int) string {
	if count == 1 {
		return "15-30 minutes"
	}
	if count <= 5 {
		return "1-2 hours"
	}
	if count <= 10 {
		return "2-4 hours"
	}
	return "4+ hours"
}

func (re *RecommendationEngine) estimateTimeForSingleResource(result *drift.DriftResult) string {
	diffCount := len(result.Differences)
	if diffCount == 1 {
		return "15 minutes"
	}
	if diffCount <= 3 {
		return "30 minutes"
	}
	return "1 hour"
}

func (re *RecommendationEngine) assessRiskLevel(category string, resourceCount int) string {
	if category == "security" || category == "critical" {
		return "high"
	}
	if resourceCount > 10 {
		return "medium"
	}
	return "low"
}

func (re *RecommendationEngine) assessRiskLevelForResource(resourceType string, severity drift.DriftSeverity) string {
	if severity == drift.SeverityCritical {
		return "high"
	}
	if strings.Contains(resourceType, "security") || strings.Contains(resourceType, "iam") {
		return "medium"
	}
	return "low"
}

func (re *RecommendationEngine) generateManualStepsForResource(resourceID string, result *drift.DriftResult) []string {
	steps := []string{
		fmt.Sprintf("Review the current configuration for %s", resourceID),
		"Compare with the desired state in your Terraform files",
	}

	for _, diff := range result.Differences {
		steps = append(steps, fmt.Sprintf("Update %s from '%v' to '%v'", diff.AttributeName, diff.ActualValue, diff.ExpectedValue))
	}

	steps = append(steps, "Apply the changes using terraform apply")
	steps = append(steps, "Verify the resource is in the expected state")

	return steps
}

// prioritizeRecommendations sorts recommendations by priority and other factors
func (re *RecommendationEngine) prioritizeRecommendations(recommendations []ActionableRecommendation) {
	sort.Slice(recommendations, func(i, j int) bool {
		// First by priority
		priorityOrder := map[RecommendationPriority]int{
			PriorityCritical: 0,
			PriorityHigh:     1,
			PriorityMedium:   2,
			PriorityLow:      3,
		}

		if priorityOrder[recommendations[i].Priority] != priorityOrder[recommendations[j].Priority] {
			return priorityOrder[recommendations[i].Priority] < priorityOrder[recommendations[j].Priority]
		}

		// Then by severity
		if recommendations[i].Severity != recommendations[j].Severity {
			return recommendations[i].Severity > recommendations[j].Severity
		}

		// Then by resource count (more resources = higher priority)
		return recommendations[i].ResourceCount > recommendations[j].ResourceCount
	})
}

// generateSummary creates a summary of all recommendations
func (re *RecommendationEngine) generateSummary(recommendations []ActionableRecommendation) *RecommendationSummary {
	summary := &RecommendationSummary{
		TotalRecommendations: len(recommendations),
		ByPriority:           make(map[RecommendationPriority]int),
		ByType:               make(map[RecommendationType]int),
		BySeverity:           make(map[drift.DriftSeverity]int),
		ByCategory:           make(map[string]int),
		Recommendations:      recommendations,
	}

	// Calculate statistics
	highestPriority := PriorityLow
	mostCommonCategory := ""
	maxCategoryCount := 0

	for _, rec := range recommendations {
		summary.ByPriority[rec.Priority]++
		summary.ByType[rec.Type]++
		summary.BySeverity[rec.Severity]++
		summary.ByCategory[rec.Category]++

		if priorityOrder := map[RecommendationPriority]int{
			PriorityCritical: 0, PriorityHigh: 1, PriorityMedium: 2, PriorityLow: 3,
		}; priorityOrder[rec.Priority] < priorityOrder[highestPriority] {
			highestPriority = rec.Priority
		}

		if count := summary.ByCategory[rec.Category]; count > maxCategoryCount {
			maxCategoryCount = count
			mostCommonCategory = rec.Category
		}
	}

	summary.HighestPriority = highestPriority
	summary.MostCommonCategory = mostCommonCategory
	summary.EstimatedTotalTime = re.calculateTotalEstimatedTime(recommendations)

	return summary
}

func (re *RecommendationEngine) calculateTotalEstimatedTime(recommendations []ActionableRecommendation) string {
	// Simple estimation based on number of recommendations
	count := len(recommendations)
	if count == 0 {
		return "0 minutes"
	}
	if count <= 5 {
		return "2-4 hours"
	}
	if count <= 10 {
		return "4-8 hours"
	}
	return "1-2 days"
}