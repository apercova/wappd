package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Config holds all processor configuration
type Config struct {
	UpdateModified   bool
	OverwriteExif    bool
	OverrideOriginal bool
	OutputDir        string
	InputDir         string
	Verbose          bool
	DryRun           bool
}

// ProcessResult holds the result of processing a single file
type ProcessResult struct {
	InputFile  string
	OutputFile string
	Success    bool
	Error      error
}

// Processor handles file processing
type Processor struct {
	config Config
}

// New creates a new Processor
func New(config Config) *Processor {
	return &Processor{config: config}
}

// ProcessFiles processes multiple files and returns results
func (p *Processor) ProcessFiles(filePaths []string) []ProcessResult {
	results := make([]ProcessResult, 0, len(filePaths))

	for _, filePath := range filePaths {
		result := p.ProcessFile(filePath)
		results = append(results, result)
	}

	return results
}

// ProcessFile processes a single file
func (p *Processor) ProcessFile(filePath string) ProcessResult {
	result := ProcessResult{InputFile: filePath}

	// Extract date from filename
	dateStr, err := ExtractDateFromFilename(filepath.Base(filePath))
	if err != nil {
		result.Error = err
		return result
	}

	// Parse the date
	parsedDateTime, err := parseISODateTime(dateStr)
	if err != nil {
		result.Error = fmt.Errorf("invalid date format: %v", err)
		return result
	}

	// Determine output path
	outputPath, err := p.determineOutputPath(filePath, p.config.OutputDir)
	if err != nil {
		result.Error = err
		return result
	}

	// In dry-run mode, skip all file operations
	if p.config.DryRun {
		result.OutputFile = outputPath
		result.Success = true
		return result
	}

	// If output dir differs from input, ensure it exists
	if p.config.OutputDir != "" {
		if err := os.MkdirAll(p.config.OutputDir, 0755); err != nil {
			result.Error = fmt.Errorf("failed to create output directory: %v", err)
			return result
		}
	}

	// Copy file to output location if different
	if outputPath != filePath {
		if err := copyFile(filePath, outputPath); err != nil {
			result.Error = fmt.Errorf("failed to copy file: %v", err)
			return result
		}
	}

	// Update EXIF data
	if err := updateExifData(outputPath, parsedDateTime, p.config); err != nil {
		// Attempt cleanup on failure
		if outputPath != filePath {
			os.Remove(outputPath)
		}
		result.Error = fmt.Errorf("failed to update EXIF data: %v", err)
		return result
	}

	// Update file modification time if requested
	if p.config.UpdateModified {
		if err := os.Chtimes(outputPath, parsedDateTime, parsedDateTime); err != nil {
			result.Error = fmt.Errorf("failed to update modification time: %v", err)
			return result
		}
	}

	result.OutputFile = outputPath
	result.Success = true
	return result
}

// ExtractDateFromFilename extracts date using default WhatsApp patterns
func ExtractDateFromFilename(filename string) (string, error) {
	// Remove extension for pattern matching
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Try default patterns
	patterns := []struct {
		regex     string
		dateGroup int
		timeGroup int
		timeFormat string
		converter func(string, string) string
	}{
		{`IMG-(\d{8})-WA`, 1, 0, "", func(d, t string) string { ds, _ := convertDateFormat(d); return ds }},
		{`VID-(\d{8})-WA`, 1, 0, "", func(d, t string) string { ds, _ := convertDateFormat(d); return ds }},
		{`WhatsApp Image (\d{4}-\d{2}-\d{2}) at (\d{1,2}\.\d{2}\.\d{2}) (AM|PM)`, 1, 2, "3.04.05 PM", func(d, t string) string { return convertDateTimeFormat(d, t) }},
		{`WhatsApp Video (\d{4}-\d{2}-\d{2}) at (\d{1,2}\.\d{2}\.\d{2}) (AM|PM)`, 1, 2, "3.04.05 PM", func(d, t string) string { return convertDateTimeFormat(d, t) }},
	}

	for _, pat := range patterns {
		re := regexp.MustCompile(pat.regex)
		matches := re.FindStringSubmatch(nameWithoutExt)
		if len(matches) > pat.dateGroup {
			dateStr := matches[pat.dateGroup]
			timeStr := ""
			if pat.timeGroup > 0 && len(matches) > pat.timeGroup {
				timeStr = matches[pat.timeGroup]
				if pat.timeGroup+1 < len(matches) {
					timeStr += " " + matches[pat.timeGroup+1]
				}
			}
			return pat.converter(dateStr, timeStr), nil
		}
	}

	return "", fmt.Errorf("no default pattern matched filename: %s", filename)
}

// convertDateFormat converts YYYYMMDD to YYYY-MM-DD
func convertDateFormat(dateStr string) (string, error) {
	if len(dateStr) != 8 {
		return "", fmt.Errorf("invalid date format, expected 8 digits: %s", dateStr)
	}

	year := dateStr[0:4]
	month := dateStr[4:6]
	day := dateStr[6:8]

	return fmt.Sprintf("%s-%s-%s", year, month, day), nil
}

// convertDateTimeFormat combines date and time strings into ISO datetime
func convertDateTimeFormat(dateStr, timeStr string) string {
	date, _ := time.Parse("2006-01-02", dateStr)
	tt, _ := time.Parse("3.04.05 PM", timeStr)
	combined := time.Date(date.Year(), date.Month(), date.Day(), tt.Hour(), tt.Minute(), tt.Second(), 0, time.UTC)
	return combined.Format("2006-01-02T15:04:05")
}

// parseISODateTime parses an ISO date or datetime string to time.Time
func parseISODateTime(dateStr string) (time.Time, error) {
	if strings.Contains(dateStr, "T") {
		return time.Parse("2006-01-02T15:04:05", dateStr)
	}
	return time.Parse("2006-01-02", dateStr)
}

// determineOutputPath determines the output file path based on configuration
func (p *Processor) determineOutputPath(inputPath, outputDir string) (string, error) {
	absInputDir, _ := filepath.Abs(p.config.InputDir)

	// If no output dir specified
	if outputDir == "" {
		if p.config.OverrideOriginal {
			return inputPath, nil
		}
		// Add suffix to original location
		return addSuffixToPath(inputPath), nil
	}

	// Output dir specified
	absOutputDir, _ := filepath.Abs(outputDir)

	// If output dir is same as input dir, add suffix
	if absOutputDir == absInputDir {
		return addSuffixToPath(inputPath), nil
	}

	// Use original filename in output directory
	filename := filepath.Base(inputPath)
	return filepath.Join(outputDir, filename), nil
}

// addSuffixToPath adds a "_modified" suffix before file extension
func addSuffixToPath(filePath string) string {
	ext := filepath.Ext(filePath)
	nameWithoutExt := strings.TrimSuffix(filePath, ext)
	return nameWithoutExt + "_modified" + ext
}

// copyFile copies a file from src to dst, preserving original file permissions
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Get original file permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	// Write file with original permissions
	return os.WriteFile(dst, data, info.Mode())
}

// GetImageVideoFiles returns all image and video files in a directory
func GetImageVideoFiles(dirPath string) ([]string, error) {
	var files []string
	supportedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true,
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".flv": true, ".m4v": true, ".3gp": true,
	}

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExts[ext] {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}
