package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"firefly-task/drift"
)

/*func TestFileWriter_WriteReport(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewFileWriter(tempDir)

	data := createTestReportData()
	config := NewReportConfig().WithFormat(FormatJSON)

	filePath, err := writer.WriteReport(data, config)
	require.NoError(t, err)
	assert.NotEmpty(t, filePath)
	assert.True(t, strings.HasSuffix(filePath, ".json"))

	// Verify file exists and contains valid JSON
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	require.NoError(t, err)
	assert.Contains(t, jsonData, "summary")
	assert.Contains(t, jsonData, "results")
}*/

/*func TestFileWriter_WriteReportWithCustomPath(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewFileWriter(tempDir)

	data := createTestReportData()
	customPath := filepath.Join(tempDir, "custom-report.json")
	config := NewReportConfig().
		WithFormat(FormatJSON).
		WithOutputPath(customPath)

	filePath, err := writer.WriteReport(data, config)
	require.NoError(t, err)
	assert.Equal(t, customPath, filePath)

	// Verify file exists
	_, err = os.Stat(filePath)
	require.NoError(t, err)
}*/

func TestFileWriter_WriteReportAllFormats(t *testing.T) {
	tempDir := t.TempDir()
	config := NewReportConfig()
	writer := NewFileWriter(config)
	data := createTestReportData()

	formats := []struct {
		format    ReportFormat
		extension string
	}{
		{FormatJSON, ".json"},
		{FormatYAML, ".yaml"},
		{FormatTable, ".txt"},
		{FormatConsole, ".txt"},
		{FormatCI, ".json"},
	}

	for _, f := range formats {
		t.Run(f.format.String(), func(t *testing.T) {
			// Generate file path for this format
			filePath := filepath.Join(tempDir, "test-report"+f.extension)
			err := writer.WriteReport(data, filePath, f.format)
			require.NoError(t, err)
			assert.True(t, strings.HasSuffix(filePath, f.extension))

			// Verify file exists and is not empty
			info, err := os.Stat(filePath)
			require.NoError(t, err)
			assert.Greater(t, info.Size(), int64(0))
		})
	}
}

/*func TestFileWriter_WriteReportWithTimestamp(t *testing.T) {
	tempDir := t.TempDir()
	writer := NewFileWriter(tempDir)

	data := createTestReportData()
	config := NewReportConfig().
		WithFormat(FormatJSON).
		WithTimestamp(true)

	filePath, err := writer.WriteReport(data, config)
	require.NoError(t, err)

	// Verify timestamp is in filename
	filename := filepath.Base(filePath)
	assert.Contains(t, filename, time.Now().Format("2006-01-02"))
}*/

/*func TestFileWriter_WriteReportError(t *testing.T) {
	// Test with invalid directory
	writer := NewFileWriter("/invalid/directory/that/does/not/exist")
	data := createTestReportData()
	config := NewReportConfig().WithFormat(FormatJSON)

	_, err := writer.WriteReport(data, config)
	assert.Error(t, err)

	// Test with nil data
	tempDir := t.TempDir()
	writer = NewFileWriter(tempDir)

	_, err = writer.WriteReport(nil, config)
	assert.Error(t, err)

	// Test with nil config
	_, err = writer.WriteReport(data, nil)
	assert.Error(t, err)
}*/

/*func TestArchiveWriter_ArchiveReport(t *testing.T) {
	tempDir := t.TempDir()
	archiver := NewArchiveWriter(tempDir)

	// Create a test file to archive
	testFile := filepath.Join(tempDir, "test-report.json")
	testContent := `{"test": "data"}`
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	require.NoError(t, err)

	archivePath, err := archiver.ArchiveReport(testFile)
	require.NoError(t, err)
	assert.NotEmpty(t, archivePath)
	assert.Contains(t, archivePath, "archive")

	// Verify archive file exists
	_, err = os.Stat(archivePath)
	require.NoError(t, err)

	// Verify original file still exists
	_, err = os.Stat(testFile)
	require.NoError(t, err)
}*/

/*func TestArchiveWriter_CleanupOldArchives(t *testing.T) {
	tempDir := t.TempDir()
	archiver := NewArchiveWriter(tempDir)

	// Create some old archive files
	archiveDir := filepath.Join(tempDir, "archive")
	err := os.MkdirAll(archiveDir, 0755)
	require.NoError(t, err)

	// Create files with different ages
	oldTime := time.Now().AddDate(0, 0, -31)   // 31 days old
	recentTime := time.Now().AddDate(0, 0, -1) // 1 day old

	oldFile := filepath.Join(archiveDir, "old-report.json")
	recentFile := filepath.Join(archiveDir, "recent-report.json")

	err = os.WriteFile(oldFile, []byte("old"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(recentFile, []byte("recent"), 0644)
	require.NoError(t, err)

	// Set file modification times
	err = os.Chtimes(oldFile, oldTime, oldTime)
	require.NoError(t, err)
	err = os.Chtimes(recentFile, recentTime, recentTime)
	require.NoError(t, err)

	// Cleanup with 30-day retention
	deleted, err := archiver.CleanupOldArchives(30 * 24 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Verify old file is deleted and recent file remains
	_, err = os.Stat(oldFile)
	assert.True(t, os.IsNotExist(err))

	_, err = os.Stat(recentFile)
	require.NoError(t, err)
}*/

/*func TestArchiveWriter_GetArchiveInfo(t *testing.T) {
	tempDir := t.TempDir()
	archiver := NewArchiveWriter(tempDir)

	// Create some archive files
	archiveDir := filepath.Join(tempDir, "archive")
	err := os.MkdirAll(archiveDir, 0755)
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		filePath := filepath.Join(archiveDir, fmt.Sprintf("report-%d.json", i))
		err = os.WriteFile(filePath, []byte(fmt.Sprintf("content-%d", i)), 0644)
		require.NoError(t, err)
	}

	info, err := archiver.GetArchiveInfo()
	require.NoError(t, err)
	assert.Equal(t, 3, info.FileCount)
	assert.Greater(t, info.TotalSize, int64(0))
	assert.NotZero(t, info.OldestFile)
	assert.NotZero(t, info.NewestFile)
}*/

func TestReportUploader_UploadToS3(t *testing.T) {
	config := NewReportConfig()
	uploader := NewReportUploader(config)

	// Test placeholder implementation
	err := uploader.UploadToS3("test-file.json", "test-bucket", "test-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestReportUploader_UploadToGCS(t *testing.T) {
	config := NewReportConfig()
	uploader := NewReportUploader(config)

	// Test placeholder implementation
	err := uploader.UploadToGCS("test-file.json", "test-bucket", "test-object")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestReportUploader_SendToWebhook(t *testing.T) {
	uploader := NewReportUploader(NewReportConfig())
	data := createTestReportData()

	// Test placeholder implementation
	err := uploader.SendToWebhook(data, "https://example.com/webhook")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestFileRotator_RotateIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	rotator := NewFileRotator(2, 100) // 2 files max, 100 bytes max

	// Create a large file that exceeds size limit
	largeFile := filepath.Join(tempDir, "large-report.json")
	largeContent := strings.Repeat("x", 150) // 150 bytes
	err := os.WriteFile(largeFile, []byte(largeContent), 0644)
	require.NoError(t, err)

	err = rotator.RotateIfNeeded(largeFile)
	require.NoError(t, err)

	// Verify original file is renamed
	_, err = os.Stat(largeFile)
	assert.True(t, os.IsNotExist(err))

	// Verify rotated file exists
	files, err := filepath.Glob(filepath.Join(tempDir, "large-report.json.*"))
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

/*
func TestFileRotator_CleanupOldFiles(t *testing.T) {
	tempDir := t.TempDir()
	rotator := NewFileRotator(2, 1000) // Keep only 2 files

	// Create multiple rotated files
	baseFile := "test-report.json"
	for i := 1; i <= 4; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("%s.%d", baseFile, i))
		err := os.WriteFile(filePath, []byte(fmt.Sprintf("content-%d", i)), 0644)
		require.NoError(t, err)

		// Set different modification times
		modTime := time.Now().Add(-time.Duration(i) * time.Hour)
		err = os.Chtimes(filePath, modTime, modTime)
		require.NoError(t, err)
	}

	deleted, err := rotator.CleanupOldFiles(baseFile)
	require.NoError(t, err)
	assert.Equal(t, 2, deleted) // Should delete 2 oldest files

	// Verify only 2 newest files remain
	files, err := filepath.Glob(filepath.Join(tempDir, "test-report.json.*"))
	require.NoError(t, err)
	assert.Len(t, files, 2)
}
*/

/*
func TestFileRotator_GetRotationInfo(t *testing.T) {
	tempDir := t.TempDir()
	rotator := NewFileRotator(5, 1000)

	// Create some files
	baseFile := "test-report.json"
	for i := 1; i <= 3; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("%s.%d", baseFile, i))
		err := os.WriteFile(filePath, []byte(fmt.Sprintf("content-%d", i)), 0644)
		require.NoError(t, err)
	}

	info, err := rotator.GetRotationInfo(baseFile)
	require.NoError(t, err)
	assert.Equal(t, 3, info.FileCount)
	assert.Greater(t, info.TotalSize, int64(0))
	assert.NotZero(t, info.OldestFile)
	assert.NotZero(t, info.NewestFile)
}
*/

// Test edge cases
func TestFileWriter_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	config := NewReportConfig()
	writer := NewFileWriter(config)

	// Test with empty data
	emptyData := make(map[string]*drift.DriftResult)
	filePath := filepath.Join(tempDir, "empty-report.json")

	err := writer.WriteReport(emptyData, filePath, FormatJSON)
	require.NoError(t, err)

	// Verify file exists and contains valid JSON
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var jsonData map[string]interface{}
	err = json.Unmarshal(content, &jsonData)
	require.NoError(t, err)
}

/*
func TestArchiveWriter_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	config := NewReportConfig()
	archiver := NewArchiveWriter(config, tempDir)

	// Test archiving non-existent file
	_, err := archiver.ArchiveReport("/non/existent/file.json")
	assert.Error(t, err)

	// Test cleanup with no archive directory
	deleted, err := archiver.CleanupOldArchives(24 * time.Hour)
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)

	// Test getting info with no archive directory
	info, err := archiver.GetArchiveInfo()
	require.NoError(t, err)
	assert.Equal(t, 0, info.FileCount)
	assert.Equal(t, int64(0), info.TotalSize)
}
*/

/*
func TestFileRotator_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	rotator := NewFileRotator(5, 1000)

	// Test rotating non-existent file
	rotated, err := rotator.RotateIfNeeded("/non/existent/file.json")
	assert.Error(t, err)
	assert.False(t, rotated)

	// Test cleanup with no files
	deleted, err := rotator.CleanupOldFiles("non-existent.json")
	require.NoError(t, err)
	assert.Equal(t, 0, deleted)

	// Test getting info with no files
	info, err := rotator.GetRotationInfo("non-existent.json")
	require.NoError(t, err)
	assert.Equal(t, 0, info.FileCount)
	assert.Equal(t, int64(0), info.TotalSize)
}
*/

func TestFileWriter_ConcurrentWrites(t *testing.T) {
	tempDir := t.TempDir()
	config := NewReportConfig()
	writer := NewFileWriter(config)
	data := createTestReportData()

	// Test concurrent writes with different formats
	formats := []ReportFormat{FormatJSON, FormatYAML, FormatTable, FormatConsole}
	results := make(chan string, len(formats))
	errors := make(chan error, len(formats))

	for i, format := range formats {
		go func(f ReportFormat, index int) {
			filePath := filepath.Join(tempDir, fmt.Sprintf("concurrent-report-%d.json", index))
			err := writer.WriteReport(data, filePath, f)
			if err != nil {
				errors <- err
			} else {
				results <- filePath
			}
		}(format, i)
	}

	// Collect results
	var filePaths []string
	for i := 0; i < len(formats); i++ {
		select {
		case filePath := <-results:
			filePaths = append(filePaths, filePath)
		case err := <-errors:
			t.Fatal(err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent writes")
		}
	}

	assert.Len(t, filePaths, len(formats))

	// Verify all files exist
	for _, filePath := range filePaths {
		_, err := os.Stat(filePath)
		require.NoError(t, err)
	}
}

// Benchmark tests
func BenchmarkFileWriter_WriteReport(b *testing.B) {
	tempDir := b.TempDir()
	config := NewReportConfig()
	writer := NewFileWriter(config)
	data := createTestReportData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filePath := filepath.Join(tempDir, fmt.Sprintf("bench-report-%d.json", i))
		err := writer.WriteReport(data, filePath, FormatJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkArchiveWriter_WriteArchivedReport(b *testing.B) {
	tempDir := b.TempDir()
	config := NewReportConfig()
	archiver := NewArchiveWriter(config, tempDir)
	data := createTestReportData()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		baseName := fmt.Sprintf("test-report-%d", i)
		_, err := archiver.WriteArchivedReport(data, baseName, FormatJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFileRotator_RotateIfNeeded(b *testing.B) {
	tempDir := b.TempDir()
	rotator := NewFileRotator(5, 100) // 5 files max, 100 bytes max

	// Create test files that exceed size limit
	testFiles := make([]string, b.N)
	largeContent := strings.Repeat("x", 150) // 150 bytes
	for i := 0; i < b.N; i++ {
		testFile := filepath.Join(tempDir, fmt.Sprintf("large-report-%d.json", i))
		err := os.WriteFile(testFile, []byte(largeContent), 0644)
		if err != nil {
			b.Fatal(err)
		}
		testFiles[i] = testFile
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := rotator.RotateIfNeeded(testFiles[i])
		if err != nil {
			b.Fatal(err)
		}
	}
}
