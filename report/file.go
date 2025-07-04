package report

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"firefly-task/drift"
)

// FileWriter handles writing reports to files with various formats
type FileWriter struct {
	config *ReportConfig
}

// NewFileWriter creates a new FileWriter
func NewFileWriter(config *ReportConfig) *FileWriter {
	return &FileWriter{
		config: config,
	}
}

// WriteReport writes a report to file with the specified format
func (fw *FileWriter) WriteReport(results map[string]*drift.DriftResult, filePath string, format ReportFormat) error {
	if results == nil {
		return NewReportError(ErrorTypeInvalidInput, "results cannot be nil")
	}

	if filePath == "" {
		return NewReportError(ErrorTypeInvalidInput, "file path cannot be empty")
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to create directory", err)
	}

	// Generate content based on format
	var content []byte
	var err error

	switch format {
	case FormatJSON:
		generator := NewStandardReportGenerator()
		content, err = generator.GenerateJSONReport(results)
	case FormatYAML:
		generator := NewStandardReportGenerator()
		content, err = generator.GenerateYAMLReport(results)
	case FormatTable:
		generator := NewConsoleReportGenerator()
		tableReport, tableErr := generator.GenerateTableReport(results)
		if tableErr != nil {
			err = tableErr
		} else {
			content = []byte(tableReport)
		}
	case FormatConsole:
		generator := NewConsoleReportGenerator()
		consoleReport, consoleErr := generator.GenerateConsoleReport(results)
		if consoleErr != nil {
			err = consoleErr
		} else {
			content = []byte(consoleReport)
		}
	case FormatCI:
		generator := NewCIReportGenerator()
		content, err = generator.GenerateJSONReport(results)
	default:
		return NewReportError(ErrorTypeUnsupportedFormat, fmt.Sprintf("unsupported format: %s", format))
	}

	if err != nil {
		return WrapReportError(ErrorTypeGenerationFailed, "failed to generate report content", err)
	}

	// Add metadata if configured
	if fw.config != nil && fw.config.IncludeTimestamp {
		content = fw.addTimestampMetadata(content, format)
	}

	// Write to file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to write file", err)
	}

	return nil
}

// WriteMultipleFormats writes the same report in multiple formats
func (fw *FileWriter) WriteMultipleFormats(results map[string]*drift.DriftResult, baseFilePath string, formats []ReportFormat) error {
	if len(formats) == 0 {
		return NewReportError(ErrorTypeInvalidInput, "no formats specified")
	}

	var errors []string
	for _, format := range formats {
		filePath := fw.getFilePathForFormat(baseFilePath, format)
		if err := fw.WriteReport(results, filePath, format); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", format, err))
		}
	}

	if len(errors) > 0 {
		return NewReportError(ErrorTypeFileOperation, fmt.Sprintf("failed to write some formats: %s", strings.Join(errors, "; ")))
	}

	return nil
}

// getFilePathForFormat generates appropriate file path for each format
func (fw *FileWriter) getFilePathForFormat(baseFilePath string, format ReportFormat) string {
	ext := filepath.Ext(baseFilePath)
	base := strings.TrimSuffix(baseFilePath, ext)

	switch format {
	case FormatJSON:
		return base + ".json"
	case FormatYAML:
		return base + ".yaml"
	case FormatTable, FormatConsole:
		return base + ".txt"
	case FormatCI:
		return base + ".ci.json"
	default:
		return baseFilePath
	}
}

// addTimestampMetadata adds timestamp information to the content
func (fw *FileWriter) addTimestampMetadata(content []byte, format ReportFormat) []byte {
	timestamp := time.Now().Format(time.RFC3339)

	switch format {
	case FormatJSON:
		// JSON doesn't support comments, so we skip adding timestamp metadata
		// to maintain valid JSON format
		return content
	case FormatYAML:
		metadata := fmt.Sprintf("# Generated at: %s\n", timestamp)
		return append([]byte(metadata), content...)
	case FormatTable, FormatConsole:
		metadata := fmt.Sprintf("Generated at: %s\n\n", timestamp)
		return append([]byte(metadata), content...)
	default:
		return content
	}
}

// ArchiveWriter handles creating archived reports
type ArchiveWriter struct {
	fileWriter *FileWriter
	archiveDir string
}

// NewArchiveWriter creates a new ArchiveWriter
func NewArchiveWriter(config *ReportConfig, archiveDir string) *ArchiveWriter {
	return &ArchiveWriter{
		fileWriter: NewFileWriter(config),
		archiveDir: archiveDir,
	}
}

// WriteArchivedReport writes a report with timestamp-based archiving
func (aw *ArchiveWriter) WriteArchivedReport(results map[string]*drift.DriftResult, baseName string, format ReportFormat) (string, error) {
	timestamp := time.Now().Format("20060102-150405")
	fileName := fmt.Sprintf("%s-%s", baseName, timestamp)
	filePath := filepath.Join(aw.archiveDir, aw.fileWriter.getFilePathForFormat(fileName, format))

	if err := aw.fileWriter.WriteReport(results, filePath, format); err != nil {
		return "", err
	}

	return filePath, nil
}

// CleanupOldReports removes old archived reports based on retention policy
func (aw *ArchiveWriter) CleanupOldReports(baseName string, maxAge time.Duration) error {
	if aw.archiveDir == "" {
		return NewReportError(ErrorTypeInvalidInput, "archive directory not set")
	}

	entries, err := os.ReadDir(aw.archiveDir)
	if err != nil {
		return WrapReportError(ErrorTypeFileOperation, "failed to read archive directory", err)
	}

	cutoff := time.Now().Add(-maxAge)
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Check if file matches our naming pattern
		if !strings.HasPrefix(entry.Name(), baseName+"-") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(aw.archiveDir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", entry.Name(), err))
			}
		}
	}

	if len(errors) > 0 {
		return NewReportError(ErrorTypeFileOperation, fmt.Sprintf("failed to cleanup some files: %s", strings.Join(errors, "; ")))
	}

	return nil
}

// ReportUploader handles uploading reports to external systems
type ReportUploader struct {
	config *ReportConfig
}

// NewReportUploader creates a new ReportUploader
func NewReportUploader(config *ReportConfig) *ReportUploader {
	return &ReportUploader{
		config: config,
	}
}

// UploadToS3 uploads a report to AWS S3 (placeholder implementation)
func (ru *ReportUploader) UploadToS3(filePath, bucket, key string) error {
	// This would integrate with AWS SDK
	// For now, return a placeholder implementation
	return NewReportError(ErrorTypeNotImplemented, "S3 upload not implemented yet")
}

// UploadToGCS uploads a report to Google Cloud Storage (placeholder implementation)
func (ru *ReportUploader) UploadToGCS(filePath, bucket, object string) error {
	// This would integrate with Google Cloud SDK
	return NewReportError(ErrorTypeNotImplemented, "GCS upload not implemented yet")
}

// SendToWebhook sends report data to a webhook endpoint (placeholder implementation)
func (ru *ReportUploader) SendToWebhook(results map[string]*drift.DriftResult, webhookURL string) error {
	// This would make HTTP POST request to webhook
	return NewReportError(ErrorTypeNotImplemented, "webhook integration not implemented yet")
}

// FileRotator handles log rotation-style file management
type FileRotator struct {
	maxFiles int
	maxSize  int64 // in bytes
}

// NewFileRotator creates a new FileRotator
func NewFileRotator(maxFiles int, maxSize int64) *FileRotator {
	return &FileRotator{
		maxFiles: maxFiles,
		maxSize:  maxSize,
	}
}

// RotateIfNeeded rotates files if they exceed size or count limits
func (fr *FileRotator) RotateIfNeeded(filePath string) error {
	// Check file size
	if info, err := os.Stat(filePath); err == nil {
		if info.Size() > fr.maxSize {
			return fr.rotateFile(filePath)
		}
	}

	// Check file count
	return fr.cleanupOldFiles(filePath)
}

// rotateFile moves current file to .1, .2, etc.
func (fr *FileRotator) rotateFile(filePath string) error {
	// Move existing numbered files up
	for i := fr.maxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", filePath, i)
		newPath := fmt.Sprintf("%s.%d", filePath, i+1)

		if _, err := os.Stat(oldPath); err == nil {
			if i == fr.maxFiles-1 {
				// Remove the oldest file
				os.Remove(oldPath)
			} else {
				os.Rename(oldPath, newPath)
			}
		}
	}

	// Move current file to .1
	if _, err := os.Stat(filePath); err == nil {
		return os.Rename(filePath, filePath+".1")
	}

	return nil
}

// cleanupOldFiles removes excess numbered files
func (fr *FileRotator) cleanupOldFiles(filePath string) error {
	for i := fr.maxFiles; i <= fr.maxFiles+10; i++ {
		oldPath := fmt.Sprintf("%s.%d", filePath, i)
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
		}
	}
	return nil
}