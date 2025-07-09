package report

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"firefly-task/pkg/interfaces"
)

// CICDPlatform represents different CI/CD platforms
type CICDPlatform string

const (
	PlatformGitHubActions CICDPlatform = "github-actions"
	PlatformGitLab        CICDPlatform = "gitlab"
	PlatformJenkins       CICDPlatform = "jenkins"
	PlatformAzureDevOps   CICDPlatform = "azure-devops"
	PlatformCircleCI      CICDPlatform = "circleci"
	PlatformGeneric       CICDPlatform = "generic"
)

// Artifact represents a CI/CD artifact file
type Artifact struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size"`
}

// CIReportGenerator implements the ReportGenerator interface for CI/CD pipelines
type CIReportGenerator struct {
	config    *ReportConfig
	Platform  CICDPlatform
	workspace string
	OutputDir string
}

// String returns the string representation of CICDPlatform
func (p CICDPlatform) String() string {
	switch p {
	case PlatformGitHubActions:
		return "github-actions"
	case PlatformGitLab:
		return "gitlab"
	case PlatformJenkins:
		return "jenkins"
	case PlatformAzureDevOps:
		return "azure-devops"
	case PlatformCircleCI:
		return "circleci"
	case PlatformGeneric:
		return "generic"
	default:
		return "unknown"
	}
}

// NewCIReportGenerator creates a new CIReportGenerator
func NewCIReportGenerator() *CIReportGenerator {
	return &CIReportGenerator{
		config:    NewReportConfig(),
		Platform:  DetectPlatform(),
		workspace: "",
		OutputDir: os.TempDir(),
	}
}

// NewCIReportGeneratorWithConfig creates a new CI/CD integration with custom config
func NewCIReportGeneratorWithConfig(config *ReportConfig, platform CICDPlatform, workspace string) *CIReportGenerator {
	return &CIReportGenerator{
		config:    config,
		Platform:  platform,
		workspace: workspace,
		OutputDir: workspace,
	}
}

// DetectPlatform automatically detects the CI/CD platform from environment
func DetectPlatform() CICDPlatform {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		return PlatformGitHubActions
	}
	if os.Getenv("GITLAB_CI") == "true" {
		return PlatformGitLab
	}
	if os.Getenv("JENKINS_URL") != "" || os.Getenv("JENKINS_HOME") != "" {
		return PlatformJenkins
	}
	if os.Getenv("AZURE_HTTP_USER_AGENT") != "" || os.Getenv("TF_BUILD") != "" {
		return PlatformAzureDevOps
	}
	if os.Getenv("CIRCLECI") == "true" {
		return PlatformCircleCI
	}
	return PlatformGeneric
}

// DetectCICDPlatform is an alias for DetectPlatform for backward compatibility
func DetectCICDPlatform() CICDPlatform {
	return DetectPlatform()
}

// WithConfig applies configuration to the generator
func (crg *CIReportGenerator) WithConfig(config *ReportConfig) ReportGenerator {
	crg.config = config
	return crg
}

// GenerateReport generates a CI-friendly report
func (crg *CIReportGenerator) GenerateReport(results map[string]*interfaces.DriftResult, config ReportConfig) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	switch config.Format {
	case FormatJSON:
		return crg.GenerateJSONReport(results)
	case FormatCI:
		interfaceResults := make(map[string]interfaces.DriftResult)
		for k, v := range results {
			interfaceResults[k] = *v
		}
		return crg.generateJUnitXMLReport(interfaceResults)
	default:
		// Default to JSON for CI
		return crg.GenerateJSONReport(results)
	}
}

// GenerateJSONReport generates a JSON report optimized for CI/CD
func (crg *CIReportGenerator) GenerateJSONReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	// Create CI-optimized report structure
	ciReport := crg.buildCIReport(results)

	jsonData, err := json.MarshalIndent(ciReport, "", "  ")
	if err != nil {
		return nil, WrapError(ErrorTypeMarshaling, "failed to marshal CI JSON report", err)
	}

	return jsonData, nil
}

// GenerateYAMLReport generates YAML (delegates to standard implementation)
func (crg *CIReportGenerator) GenerateYAMLReport(results map[string]*interfaces.DriftResult) ([]byte, error) {
	return nil, fmt.Errorf("YAML report generation is not supported in CICD mode with the new interface types")
}

// GenerateTableReport generates a simple table for CI logs
func (crg *CIReportGenerator) generateSummary(results map[string]interfaces.DriftResult) (string, error) {
	if results == nil {
		return "", NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	var builder strings.Builder

	// Simple header for CI logs
	builder.WriteString("DRIFT DETECTION SUMMARY\n")
	builder.WriteString(strings.Repeat("=", 50) + "\n")

	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0

	for _, result := range results {
		if result.IsDrifted {
				resourcesWithDrift++
				totalDifferences += len(result.DriftDetails)
			}
		}

	builder.WriteString(fmt.Sprintf("Total Resources: %d\n", totalResources))
	builder.WriteString(fmt.Sprintf("Resources with Drift: %d\n", resourcesWithDrift))
	builder.WriteString(fmt.Sprintf("Total Differences: %d\n", totalDifferences))

	if resourcesWithDrift > 0 {
		builder.WriteString("\nRESOURCES WITH DRIFT:\n")
		for resourceID, result := range results {
			if result.IsDrifted {
				builder.WriteString(fmt.Sprintf("- %s (%d differences, severity: %s)\n",
					resourceID, len(result.DriftDetails), string(result.Severity)))
			}
		}
	} else {
		builder.WriteString("\n✓ No drift detected\n")
	}

	return builder.String(), nil
}

// generateJUnitXMLReport generates JUnit XML format for CI systems
func (crg *CIReportGenerator) generateJUnitXMLReport(results map[string]interfaces.DriftResult) ([]byte, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	var testCases []JUnitTestCase
	failures := 0

	for resourceID, result := range results {
		testCase := JUnitTestCase{
			Name:      fmt.Sprintf("drift-check-%s", resourceID),
			ClassName: "drift.detection",
			Time:      0.001,
		}

		if result.IsDrifted {
			failures++
			testCase.Failure = &JUnitFailure{
				Message: fmt.Sprintf("Drift detected in %s", resourceID),
				Type:    "DriftDetected",
				Content: fmt.Sprintf("Resource %s has %d differences with %s severity", resourceID, len(result.DriftDetails), string(result.Severity)),
			}
		}

		testCases = append(testCases, testCase)
	}

	testSuite := JUnitTestSuite{
		Name:      "Terraform Drift Detection",
		Tests:     len(results),
		Failures:  failures,
		Time:      0.001,
		TestCases: testCases,
	}

	return xml.MarshalIndent(testSuite, "", "  ")
}


func (crg *CIReportGenerator) GenerateTableReport(results map[string]*interfaces.DriftResult) (string, error) {
	interfaceResults := make(map[string]interfaces.DriftResult)
	for k, v := range results {
		interfaceResults[k] = *v
	}
	return crg.generateSummary(interfaceResults)
}

func (crg *CIReportGenerator) GenerateConsoleReport(results map[string]*interfaces.DriftResult) (string, error) {
	return crg.GenerateTableReport(results)
}

// WriteToFile writes the report to a file
func (crg *CIReportGenerator) WriteToFile(content []byte, filePath string) error {
	return os.WriteFile(filePath, content, 0644)
}

// GenerateCIReport generates a CI/CD-optimized report
func (crg *CIReportGenerator) GenerateCIReport(results map[string]interfaces.DriftResult) (*CIReport, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	interfaceResults := make(map[string]*interfaces.DriftResult)
	for k, v := range results {
		// Create a new variable to hold the value and get its address
		vc := v
		interfaceResults[k] = &vc
	}
	report := crg.buildCIReport(interfaceResults)

	// Add metadata that was previously in this function
	report.Metadata.Platform = string(crg.Platform)
	report.Metadata.Workspace = crg.workspace
	report.Metadata.Timestamp = time.Now()
	report.Metadata.JobID = crg.getJobID()
	report.Metadata.BuildNumber = crg.getBuildNumber()
	report.Metadata.Branch = crg.getBranch()
	report.Metadata.CommitSHA = crg.getCommitSHA()

	return report, nil
}

// CI-specific structures

// CIReport represents a CI-optimized report structure
type CIReport struct {
	Version   string                             `json:"version"`
	Type      string                             `json:"type"`
	Timestamp string                             `json:"timestamp"`
	Status    string                             `json:"status"`
	ExitCode  int                                `json:"exit_code"`
	Summary   CISummary                          `json:"summary"`
	Results   map[string]*interfaces.DriftResult `json:"results"`
	Actions   []CIAction                         `json:"actions"`
	Metadata  CIMetadata                         `json:"metadata"`
}

// CISummary contains CI-relevant summary information
type CISummary struct {
	TotalResources     int            `json:"total_resources"`
	ResourcesWithDrift int            `json:"resources_with_drift"`
	DriftedResources   int            `json:"drifted_resources"`
	CleanResources     int            `json:"clean_resources"`
	TotalDifferences   int            `json:"total_differences"`
	SeverityCounts     map[string]int `json:"severity_counts"`
	HighestSeverity    string         `json:"highest_severity"`
	Passed             bool           `json:"passed"`
}

// CIAction represents an action that should be taken
type CIAction struct {
	Type        string `json:"type"`
	ResourceID  string `json:"resource_id"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Command     string `json:"command,omitempty"`
}

// CIMetadata contains CI pipeline metadata
type CIMetadata struct {
	Generator     string            `json:"generator"`
	GeneratedAt   string            `json:"generated_at"`
	GeneratorType string            `json:"generator_type"`
	Version       string            `json:"version"`
	PipelineInfo  map[string]string `json:"pipeline_info,omitempty"`
	// Additional fields for CICD integration
	Platform    string    `json:"platform,omitempty"`
	Workspace   string    `json:"workspace,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	JobID       string    `json:"job_id,omitempty"`
	BuildNumber string    `json:"build_number,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
}

// JUnit XML structures for test result integration
type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      float64         `xml:"time,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	ClassName string        `xml:"classname,attr"`
	Time      float64       `xml:"time,attr"`
	Failure   *JUnitFailure `xml:"failure,omitempty"`
}

type JUnitFailure struct {
	XMLName xml.Name `xml:"failure"`
	Message string   `xml:"message,attr"`
	Type    string   `xml:"type,attr"`
	Content string   `xml:",chardata"`
}

// buildCIReport creates a CI-optimized report
func (crg *CIReportGenerator) buildCIReport(results map[string]*interfaces.DriftResult) *CIReport {
	summary := crg.buildCISummary(results)
	actions := crg.generateCIActions(results)

	status := "success"
	exitCode := 0
	if summary.ResourcesWithDrift > 0 {
		status = "failure"
		exitCode = 1
	}

	timestamp := time.Now().Format(time.RFC3339)

	return &CIReport{
		Version:   "1.0",
		Type:      "drift-detection",
		Timestamp: timestamp,
		Status:    status,
		ExitCode:  exitCode,
		Summary:   summary,
		Results:   results,
		Actions:   actions,
		Metadata: CIMetadata{
			Generator:     "firefly-task",
			GeneratedAt:   timestamp,
			GeneratorType: "ci",
			Version:       "1.0.0",
			Platform:      string(crg.Platform),
			Workspace:     crg.workspace,
			JobID:         crg.getJobID(),
			BuildNumber:   crg.getBuildNumber(),
			Branch:        crg.getBranch(),
			CommitSHA:     crg.getCommitSHA(),
		},
	}
}

// buildCISummary creates a CI-focused summary
func (crg *CIReportGenerator) buildCISummary(results map[string]*interfaces.DriftResult) CISummary {
	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0
	severityCounts := make(map[string]int)
	highestSeverity := interfaces.SeverityNone

	for _, result := range results {
		if result.IsDrifted {
			resourcesWithDrift++
			totalDifferences += len(result.DriftDetails)
			if result.Severity > highestSeverity {
				highestSeverity = result.Severity
			}
		}
		severityCounts[strings.ToLower(string(result.Severity))]++
	}

	cleanResources := totalResources - resourcesWithDrift

	highestSeverityStr := "NONE"
	if highestSeverity != interfaces.SeverityNone {
		highestSeverityStr = strings.ToUpper(string(highestSeverity))
	}

	return CISummary{
		TotalResources:     totalResources,
		ResourcesWithDrift: resourcesWithDrift,
		DriftedResources:   resourcesWithDrift,
		CleanResources:     cleanResources,
		TotalDifferences:   totalDifferences,
		SeverityCounts:     severityCounts,
		HighestSeverity:    highestSeverityStr,
		Passed:             resourcesWithDrift == 0,
	}
}

// generateCIActions creates actionable items from drift results
func (crg *CIReportGenerator) generateCIActions(results map[string]*interfaces.DriftResult) []CIAction {
	var actions []CIAction

	for resourceID, result := range results {
		if !result.IsDrifted {
			continue
		}

		for _, diff := range result.DriftDetails {
			action := CIAction{
				Type:        "drift-detected",
				ResourceID:  resourceID,
				Description: fmt.Sprintf("Drift detected in %s: %s", diff.Attribute, diff.Description),
				Priority:    strings.ToLower(string(diff.Severity)),
			}

			// Add command suggestions based on drift type
			switch diff.DriftType {
			case "added":
				action.Command = fmt.Sprintf("terraform import %s", resourceID)
			case "removed":
				action.Command = fmt.Sprintf("terraform apply -target=%s", resourceID)
			case "modified":
				action.Command = fmt.Sprintf("terraform plan -target=%s", resourceID)
			default:
				action.Command = fmt.Sprintf("terraform plan -target=%s", resourceID)
			}

			actions = append(actions, action)
		}
	}

	// Sort actions by priority (high -> medium -> low)
	sort.Slice(actions, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2, "none": 3}
		return priorityOrder[actions[i].Priority] < priorityOrder[actions[j].Priority]
	})

	return actions
}

// WriteArtifacts writes CI/CD artifacts (reports, logs, etc.)
func (crg *CIReportGenerator) WriteArtifacts(results map[string]*interfaces.DriftResult) ([]Artifact, error) {
	if results == nil {
		return nil, NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}
	artifactDir := crg.OutputDir
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to create artifact directory", err)
	}

	var artifacts []Artifact

	// Write JSON artifact
	jsonArtifact, err := crg.WriteJSONArtifact(results)
	if err != nil {
		return nil, err
	}
	artifacts = append(artifacts, *jsonArtifact)

	// Write JUnit XML artifact
	junitArtifact, err := crg.WriteJUnitXMLArtifact(results)
	if err != nil {
		return nil, err
	}
	artifacts = append(artifacts, *junitArtifact)

	// Write summary artifact
	summaryArtifact, err := crg.WriteSummaryArtifact(results)
	if err != nil {
		return nil, err
	}
	artifacts = append(artifacts, *summaryArtifact)

	// Write platform-specific artifacts
		interfaceResults := make(map[string]interfaces.DriftResult)
	for k, v := range results {
		interfaceResults[k] = *v
	}
	platformArtifacts, err := crg.writePlatformSpecificArtifacts(interfaceResults, artifactDir)
	if err != nil {
		return artifacts, err
	}
	artifacts = append(artifacts, platformArtifacts...)

	return artifacts, nil
}

// WriteJSONArtifact writes a JSON artifact and returns artifact info
func (crg *CIReportGenerator) WriteJSONArtifact(results map[string]*interfaces.DriftResult) (*Artifact, error) {
	// Convert to interface results
	interfaceResults := make(map[string]interfaces.DriftResult)
	for k, v := range results {
		interfaceResults[k] = *v
	}
	// Generate CI report
	ciReport, err := crg.GenerateCIReport(interfaceResults)
	if err != nil {
		return nil, err
	}

	// Write CI report as JSON
	filePath := filepath.Join(crg.OutputDir, "drift-report.ci.json")
	if err = crg.writeJSONFile(ciReport, filePath); err != nil {
		return nil, err
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to get file info", err)
	}

	return &Artifact{
		Path: filePath,
		Type: "json",
		Size: fileInfo.Size(),
	}, nil
}

// WriteJUnitXMLArtifact writes a JUnit XML artifact and returns artifact info
func (crg *CIReportGenerator) WriteJUnitXMLArtifact(results map[string]*interfaces.DriftResult) (*Artifact, error) {
	filePath := filepath.Join(crg.OutputDir, "drift-report.junit.xml")
	// Convert to interface results
	interfaceResults := make(map[string]interfaces.DriftResult)
	for k, v := range results {
		interfaceResults[k] = *v
	}
	if err := crg.writeJUnitXML(interfaceResults, filePath); err != nil {
		return nil, err
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to get file info", err)
	}

	return &Artifact{
		Path: filePath,
		Type: "junit-xml",
		Size: fileInfo.Size(),
	}, nil
}

// WriteSummaryArtifact writes a summary artifact and returns artifact info
func (crg *CIReportGenerator) WriteSummaryArtifact(results map[string]*interfaces.DriftResult) (*Artifact, error) {
	filePath := filepath.Join(crg.OutputDir, "drift-summary.md")
	summary, err := crg.generateMarkdownSummary(results)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filePath, []byte(summary), 0644); err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to write summary file", err)
	}

	// Get file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to get file info", err)
	}

	return &Artifact{
		Path: filePath,
		Type: "summary",
		Size: fileInfo.Size(),
	}, nil
}

// SetExitCode sets appropriate exit code based on drift results
func (crg *CIReportGenerator) SetExitCode(results map[string]*interfaces.DriftResult) int {
	if results == nil {
		return 1 // Error
	}

	hasCritical := false
	hasHigh := false
	hasDrift := false

	for _, result := range results {
		if result.IsDrifted {
			hasDrift = true
			switch result.Severity {
			case interfaces.SeverityCritical:
				hasCritical = true
			case interfaces.SeverityHigh:
				hasHigh = true
			}
		}
	}

	if hasCritical {
		return 2 // Critical drift
	}
	if hasHigh {
		return 1 // High severity drift
	}
	if hasDrift {
		return 0 // Drift detected but not critical
	}
	return 0 // No drift
}

// SetEnvironmentVariables sets CI/CD environment variables with results
func (crg *CIReportGenerator) SetEnvironmentVariables(results map[string]*interfaces.DriftResult) error {
	summary := crg.buildCISummary(results)

	// Determine max severity
	maxSeverity := "low"
	if summary.SeverityCounts["critical"] > 0 {
		maxSeverity = "critical"
	} else if summary.SeverityCounts["high"] > 0 {
		maxSeverity = "high"
	} else if summary.SeverityCounts["medium"] > 0 {
		maxSeverity = "medium"
	}

	// Set common environment variables
	envVars := map[string]string{
		"DRIFT_TOTAL_RESOURCES":      strconv.Itoa(summary.TotalResources),
		"DRIFT_RESOURCES_WITH_DRIFT": strconv.Itoa(summary.ResourcesWithDrift),
		"DRIFT_TOTAL_DIFFERENCES":    strconv.Itoa(summary.TotalDifferences),
		"DRIFT_MAX_SEVERITY":         maxSeverity,
		"DRIFT_CRITICAL_COUNT":       strconv.Itoa(summary.SeverityCounts["critical"]),
		"DRIFT_HIGH_COUNT":           strconv.Itoa(summary.SeverityCounts["high"]),
		"DRIFT_MEDIUM_COUNT":         strconv.Itoa(summary.SeverityCounts["medium"]),
		"DRIFT_LOW_COUNT":            strconv.Itoa(summary.SeverityCounts["low"]),
		"DRIFT_HAS_DRIFT":            strconv.FormatBool(summary.ResourcesWithDrift > 0),
		"DRIFT_TIMESTAMP":            time.Now().Format(time.RFC3339),
	}

	// Set platform-specific environment variables
	switch crg.Platform {
	case PlatformGitHubActions:
		return crg.setGitHubActionsEnv(envVars, results)
	case PlatformGitLab:
		return crg.setGitLabEnv(envVars, results)
	case PlatformJenkins:
		return crg.setJenkinsEnv(envVars, results)
	default:
		return crg.setGenericEnv(envVars)
	}
}

// ArtifactInfo contains summary information about generated artifacts
type ArtifactInfo struct {
	FileCount int           `json:"file_count"`
	TotalSize int64         `json:"total_size"`
	Files     []os.FileInfo `json:"-"` // Excluded from JSON serialization
}

// GetArtifactInfo returns summary information about the generated artifacts
func (crg *CIReportGenerator) GetArtifactInfo() (*ArtifactInfo, error) {
	files, err := os.ReadDir(crg.OutputDir)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to read artifact directory", err)
	}

	var totalSize int64
	var fileInfos []os.FileInfo
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				return nil, WrapReportError(ErrorTypeFileOperation, "failed to get file info", err)
			}
			fileInfos = append(fileInfos, info)
			totalSize += info.Size()
		}
	}

	return &ArtifactInfo{
		FileCount: len(fileInfos),
		TotalSize: totalSize,
		Files:     fileInfos,
	}, nil
}

// Helper methods for CI/CD metadata

func (crg *CIReportGenerator) getJobID() string {
	switch crg.Platform {
	case PlatformGitHubActions:
		return os.Getenv("GITHUB_RUN_ID")
	case PlatformGitLab:
		return os.Getenv("CI_JOB_ID")
	case PlatformJenkins:
		return os.Getenv("BUILD_ID")
	case PlatformAzureDevOps:
		return os.Getenv("BUILD_BUILDID")
	case PlatformCircleCI:
		return os.Getenv("CIRCLE_BUILD_NUM")
	default:
		return "unknown"
	}
}

func (crg *CIReportGenerator) getBuildNumber() string {
	switch crg.Platform {
	case PlatformGitHubActions:
		return os.Getenv("GITHUB_RUN_NUMBER")
	case PlatformGitLab:
		return os.Getenv("CI_PIPELINE_ID")
	case PlatformJenkins:
		return os.Getenv("BUILD_NUMBER")
	case PlatformAzureDevOps:
		return os.Getenv("BUILD_BUILDNUMBER")
	case PlatformCircleCI:
		return os.Getenv("CIRCLE_BUILD_NUM")
	default:
		return "unknown"
	}
}

func (crg *CIReportGenerator) getBranch() string {
	switch crg.Platform {
	case PlatformGitHubActions:
		return strings.TrimPrefix(os.Getenv("GITHUB_REF"), "refs/heads/")
	case PlatformGitLab:
		return os.Getenv("CI_COMMIT_REF_NAME")
	case PlatformJenkins:
		return os.Getenv("GIT_BRANCH")
	case PlatformAzureDevOps:
		return os.Getenv("BUILD_SOURCEBRANCHNAME")
	case PlatformCircleCI:
		return os.Getenv("CIRCLE_BRANCH")
	default:
		return "unknown"
	}
}

func (crg *CIReportGenerator) getCommitSHA() string {
	switch crg.Platform {
	case PlatformGitHubActions:
		return os.Getenv("GITHUB_SHA")
	case PlatformGitLab:
		return os.Getenv("CI_COMMIT_SHA")
	case PlatformJenkins:
		return os.Getenv("GIT_COMMIT")
	case PlatformAzureDevOps:
		return os.Getenv("BUILD_SOURCEVERSION")
	case PlatformCircleCI:
		return os.Getenv("CIRCLE_SHA1")
	default:
		return "unknown"
	}
}

// SetPlatformSpecificVariables sets CI/CD environment variables with results
func (crg *CIReportGenerator) SetPlatformSpecificVariables(results map[string]*interfaces.DriftResult) error {
	if results == nil {
		return NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}
	summary := crg.buildCISummary(results)

	// Set common environment variables
	envVars := map[string]string{
		"DRIFT_TOTAL_RESOURCES":      strconv.Itoa(summary.TotalResources),
		"DRIFT_RESOURCES_WITH_DRIFT": strconv.Itoa(summary.ResourcesWithDrift),
		"DRIFT_TOTAL_DIFFERENCES":    strconv.Itoa(summary.TotalDifferences),
		"DRIFT_CRITICAL_COUNT":       strconv.Itoa(summary.SeverityCounts["critical"]),
		"DRIFT_HIGH_COUNT":           strconv.Itoa(summary.SeverityCounts["high"]),
		"DRIFT_MEDIUM_COUNT":         strconv.Itoa(summary.SeverityCounts["medium"]),
		"DRIFT_LOW_COUNT":            strconv.Itoa(summary.SeverityCounts["low"]),
		"DRIFT_HAS_DRIFT":            strconv.FormatBool(summary.ResourcesWithDrift > 0),
		"DRIFT_TIMESTAMP":            time.Now().Format(time.RFC3339),
	}

	// Set platform-specific environment variables
	switch crg.Platform {
	case PlatformGitHubActions:
		return crg.setGitHubActionsEnv(envVars, results)
	case PlatformGitLab:
		return crg.setGitLabEnv(envVars, results)
	case PlatformJenkins:
		return crg.setJenkinsEnv(envVars, results)
	default:
		return crg.setGenericEnv(envVars)
	}
}

// Platform-specific environment variable setters

func (crg *CIReportGenerator) setGitHubActionsEnv(envVars map[string]string, results map[string]*interfaces.DriftResult) error {
	// GitHub Actions uses GITHUB_ENV file
	githubEnvFile := os.Getenv("GITHUB_ENV")
	if githubEnvFile != "" {
		file, err := os.OpenFile(githubEnvFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to open GITHUB_ENV file", err)
		}
		defer file.Close()

		for key, value := range envVars {
			if _, err := file.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
				return WrapReportError(ErrorTypeFileOperation, "failed to write to GITHUB_ENV", err)
			}
		}
	}

	// Set GitHub Actions outputs
	githubOutputFile := os.Getenv("GITHUB_OUTPUT")
	if githubOutputFile != "" {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(githubOutputFile), 0755); err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to create parent directory for GITHUB_OUTPUT file", err)
		}
		outputFile, err := os.OpenFile(githubOutputFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to open GITHUB_OUTPUT file", err)
		}
		defer outputFile.Close()
		for key, value := range envVars {
			outputFile.WriteString(fmt.Sprintf("%s=%s\n", strings.ToLower(key), value))
		}
	}

	// Set GitHub Actions step summary
	githubSummaryFile := os.Getenv("GITHUB_STEP_SUMMARY")
	if githubSummaryFile != "" {
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(githubSummaryFile), 0755); err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to create parent directory for GITHUB_STEP_SUMMARY file", err)
		}
		summaryFile, err := os.OpenFile(githubSummaryFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to open GITHUB_STEP_SUMMARY file", err)
		}
		defer summaryFile.Close()
		summary, err := crg.generateMarkdownSummary(results)
		if err != nil {
			return err
		}
		summaryFile.WriteString(summary)
	}

	return nil
}

func (crg *CIReportGenerator) setGitLabEnv(envVars map[string]string, results map[string]*interfaces.DriftResult) error {
	// GitLab CI uses dotenv artifacts
	dotenvFile := filepath.Join(crg.workspace, "drift.env")
	file, err := os.Create(dotenvFile)
	if err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to create dotenv file", err)
	}
	defer file.Close()

	for key, value := range envVars {
		if _, err := file.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to write to dotenv file", err)
		}
	}

	return nil
}

func (crg *CIReportGenerator) setJenkinsEnv(envVars map[string]string, results map[string]*interfaces.DriftResult) error {
	// Jenkins can use properties file
	propsFile := filepath.Join(crg.workspace, "drift.properties")
	file, err := os.Create(propsFile)
	if err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to create properties file", err)
	}
	defer file.Close()

	for key, value := range envVars {
		if _, err := file.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
			return WrapReportError(ErrorTypeFileOperation, "failed to write to properties file", err)
		}
	}

	return nil
}

func (crg *CIReportGenerator) setGenericEnv(envVars map[string]string) error {
	// For generic platforms, just set environment variables
	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			return WrapReportError(ErrorTypeFileOperation, fmt.Sprintf("failed to set environment variable %s", key), err)
		}
	}
	return nil
}

// Helper methods for file writing

func (crg *CIReportGenerator) writeJSONFile(data interface{}, filePath string) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return WrapReportError(ErrorTypeGenerationFailed, "failed to marshal JSON", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to write JSON file", err)
	}

	return nil
}

func (crg *CIReportGenerator) writeJUnitXML(results map[string]interfaces.DriftResult, filePath string) error {
	// Generate JUnit XML content
	xmlContent, err := crg.generateJUnitXMLReport(results)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filePath, xmlContent, 0644); err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to write JUnit XML file", err)
	}

	return nil
}

func (crg *CIReportGenerator) writeSummaryFile(results map[string]*interfaces.DriftResult, filePath string) error {
	summary := crg.buildCISummary(results)
	content := fmt.Sprintf(`Drift Detection Summary
======================

Total Resources: %d
Resources with Drift: %d
Total Differences: %d

Severity Breakdown:
- Critical: %d
- High: %d
- Medium: %d
- Low: %d

Generated: %s
Platform: %s
`,
		summary.TotalResources,
		summary.ResourcesWithDrift,
		summary.TotalDifferences,
		summary.SeverityCounts["critical"],
		summary.SeverityCounts["high"],
		summary.SeverityCounts["medium"],
		summary.SeverityCounts["low"],
		time.Now().Format(time.RFC3339),
		crg.Platform,
	)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to write summary file", err)
	}

	return nil
}

func (crg *CIReportGenerator) writePlatformSpecificArtifacts(results map[string]interfaces.DriftResult, artifactDir string) ([]Artifact, error) {
	switch crg.Platform {
	case PlatformGitHubActions:
		return crg.writeGitHubActionsArtifacts(results, artifactDir)
	case PlatformGitLab:
		return crg.writeGitLabArtifacts(results, artifactDir)
	case PlatformJenkins:
		return crg.writeJenkinsArtifacts(results, artifactDir)
	default:
		return nil, nil // No platform-specific artifacts
	}
}

func (crg *CIReportGenerator) writeGitHubActionsArtifacts(results map[string]interfaces.DriftResult, artifactDir string) ([]Artifact, error) {
	// Write GitHub Actions job summary
	summaryFile := filepath.Join(artifactDir, "github-summary.md")
	// Convert to pointer results
	pointerResults := make(map[string]*interfaces.DriftResult)
	for k, v := range results {
		vc := v
		pointerResults[k] = &vc
	}
	summary, err := crg.generateMarkdownSummary(pointerResults)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(summaryFile, []byte(summary), 0644)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to write GitHub summary", err)
	}
	info, err := os.Stat(summaryFile)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to stat GitHub summary", err)
	}
	return []Artifact{{
		Path: summaryFile,
		Type: "github-summary-md",
		Size: info.Size(),
	}}, nil
}

func (crg *CIReportGenerator) writeGitLabArtifacts(results map[string]interfaces.DriftResult, artifactDir string) ([]Artifact, error) {
	// Write GitLab merge request note
	noteFile := filepath.Join(artifactDir, "gitlab-note.md")
	// Convert to pointer results
	pointerResults := make(map[string]*interfaces.DriftResult)
	for k, v := range results {
		vc := v
		pointerResults[k] = &vc
	}
	note, err := crg.generateMarkdownSummary(pointerResults)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(noteFile, []byte(note), 0644)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to write GitLab note", err)
	}
	info, err := os.Stat(noteFile)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to stat GitLab note", err)
	}
	return []Artifact{{
		Path: noteFile,
		Type: "gitlab-note-md",
		Size: info.Size(),
	}}, nil
}

func (crg *CIReportGenerator) writeJenkinsArtifacts(results map[string]interfaces.DriftResult, artifactDir string) ([]Artifact, error) {
	// Write Jenkins HTML report
	htmlFile := filepath.Join(artifactDir, "jenkins-report.html")
	// Convert to pointer results
	pointerResults := make(map[string]*interfaces.DriftResult)
	for k, v := range results {
		vc := v
		pointerResults[k] = &vc
	}
	html, err := crg.generateHTMLSummary(pointerResults)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(htmlFile, []byte(html), 0644)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to write Jenkins HTML report", err)
	}
	info, err := os.Stat(htmlFile)
	if err != nil {
		return nil, WrapReportError(ErrorTypeFileOperation, "failed to stat Jenkins HTML report", err)
	}
	return []Artifact{{
		Path: htmlFile,
		Type: "jenkins-html-report",
		Size: info.Size(),
	}}, nil
}

// Summary generation helpers

func (crg *CIReportGenerator) generateMarkdownSummary(results map[string]*interfaces.DriftResult) (string, error) {
	summary := crg.buildCISummary(results)

	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Terraform Drift Detection Summary\n\n## Summary\n- **Total Resources**: %d\n- **Resources with Drift**: %d\n- **Total Differences**: %d\n\n## Severity Breakdown\n- 🔴 **Critical**: %d\n- 🟠 **High**: %d\n- 🟡 **Medium**: %d\n- 🔵 **Low**: %d\n",
		summary.TotalResources,
		summary.ResourcesWithDrift,
		summary.TotalDifferences,
		summary.SeverityCounts["critical"],
		summary.SeverityCounts["high"],
		summary.SeverityCounts["medium"],
		summary.SeverityCounts["low"],
	))

	if summary.ResourcesWithDrift == 0 {
		md.WriteString("\n## ✅ Result\n\nNo drift detected! All resources are in sync.\n")
	} else {
		md.WriteString("\n## ⚠️ Action Required\n\nDrift detected in infrastructure. Review the detailed report and consider running `terraform plan` and `terraform apply`.\n")
	}

	return md.String(), nil
}

func (crg *CIReportGenerator) generateHTMLSummary(results map[string]*interfaces.DriftResult) (string, error) {
	summary := crg.buildCISummary(results)

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Drift Detection Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .summary { background: #f5f5f5; padding: 15px; border-radius: 5px; }
        .critical { color: #d32f2f; }
        .high { color: #f57c00; }
        .medium { color: #fbc02d; }
        .low { color: #1976d2; }
        .success { color: #388e3c; }
    </style>
</head>
<body>
    <h1>🔍 Drift Detection Report</h1>
    <div class="summary">
        <h2>Summary</h2>
        <p><strong>Total Resources:</strong> %d</p>
        <p><strong>Resources with Drift:</strong> %d</p>
        <p><strong>Total Differences:</strong> %d</p>
        
        <h3>Severity Breakdown</h3>
        <p class="critical">Critical: %d</p>
        <p class="high">High: %d</p>
        <p class="medium">Medium: %d</p>
        <p class="low">Low: %d</p>
    </div>
    
    <p><em>Generated: %s</em></p>
</body>
</html>`,
		summary.TotalResources,
		summary.ResourcesWithDrift,
		summary.TotalDifferences,
		summary.SeverityCounts["critical"],
		summary.SeverityCounts["high"],
		summary.SeverityCounts["medium"],
		summary.SeverityCounts["low"],
		time.Now().Format(time.RFC3339),
	), nil
}

// generateCISummary creates a CI-focused summary with detailed counts
func (crg *CIReportGenerator) generateCISummary(results map[string]*interfaces.DriftResult) CISummary {
	totalResources := len(results)
	resourcesWithDrift := 0
	totalDifferences := 0
	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, result := range results {
		if result.IsDrifted {
			resourcesWithDrift++
			totalDifferences += len(result.DriftDetails)
		}
		switch result.Severity {
		case interfaces.SeverityCritical:
			criticalCount++
		case interfaces.SeverityHigh:
			highCount++
		case interfaces.SeverityMedium:
			mediumCount++
		case interfaces.SeverityLow:
			lowCount++
		}
	}

	severityCounts := map[string]int{
		"critical": criticalCount,
		"high":     highCount,
		"medium":   mediumCount,
		"low":      lowCount,
	}

	return CISummary{
		TotalResources:     totalResources,
		ResourcesWithDrift: resourcesWithDrift,
		TotalDifferences:   totalDifferences,
		SeverityCounts:     severityCounts,
		Passed:             resourcesWithDrift == 0,
	}
}
