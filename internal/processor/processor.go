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
	DateTimeOverride string
	RegexPattern     string
	PatternFormat    string
	UpdateModified   bool
	OverwriteExif    bool
	OverrideOriginal bool
	OutputDir        string
	InputDir         string
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

	// Extract date from filename or use override
	dateStr := p.config.DateTimeOverride
	if dateStr == "" {
		var err error
		dateStr, err = p.extractDateFromFilename(filepath.Base(filePath))
		if err != nil {
			result.Error = err
			return result
		}
	}

	// Parse the date
	parsedDate, err := parseISODate(dateStr)
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
	if err := updateExifData(outputPath, parsedDate, p.config); err != nil {
		// Attempt cleanup on failure
		if outputPath != filePath {
			os.Remove(outputPath)
		}
		result.Error = fmt.Errorf("failed to update EXIF data: %v", err)
		return result
	}

	// Update file modification time if requested
	if p.config.UpdateModified {
		if err := os.Chtimes(outputPath, parsedDate, parsedDate); err != nil {
			result.Error = fmt.Errorf("failed to update modification time: %v", err)
			return result
		}
	}

	result.OutputFile = outputPath
	result.Success = true
	return result
}

// extractDateFromFilename extracts date using regex or pattern
func (p *Processor) extractDateFromFilename(filename string) (string, error) {
	// Remove extension for pattern matching
	nameWithoutExt := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Use custom regex if provided
	if p.config.RegexPattern != "" {
		re, err := regexp.Compile(p.config.RegexPattern)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %v", err)
		}

		matches := re.FindStringSubmatch(filename)
		if len(matches) < 2 {
			return "", fmt.Errorf("regex pattern did not match filename: %s", filename)
		}

		// Try to find named group "date"
		for i, name := range re.SubexpNames() {
			if name == "date" && i < len(matches) {
				return convertDateFormat(matches[i])
			}
		}

		return "", fmt.Errorf("regex pattern does not contain named group 'date'")
	}

	// Use custom pattern if provided
	if p.config.PatternFormat != "" {
		return extractDateFromPattern(nameWithoutExt, p.config.PatternFormat)
	}

	// Default WhatsApp pattern: IMG-YYYYMMDD-WA
	return extractDateFromWhatsAppPattern(nameWithoutExt)
}

// extractDateFromWhatsAppPattern extracts date from default WhatsApp pattern
func extractDateFromWhatsAppPattern(filename string) (string, error) {
	// Pattern: IMG-YYYYMMDD-WA
	re := regexp.MustCompile(`IMG-(\d{8})-WA`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return "", fmt.Errorf("filename does not match WhatsApp pattern (IMG-YYYYMMDD-WA): %s", filename)
	}

	return convertDateFormat(matches[1])
}

// extractDateFromPattern extracts date from custom pattern format
func extractDateFromPattern(filename, pattern string) (string, error) {
	// Convert pattern to regex
	regexPattern := strings.ReplaceAll(regexp.QuoteMeta(pattern), `\{date\}`, `(\d{8})`)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return "", fmt.Errorf("invalid pattern format: %v", err)
	}

	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return "", fmt.Errorf("filename does not match pattern '%s': %s", pattern, filename)
	}

	return convertDateFormat(matches[1])
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

// parseISODate parses an ISO date string (YYYY-MM-DD) to time.Time
func parseISODate(dateStr string) (time.Time, error) {
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

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetImageVideoFiles returns all image and video files in a directory
func GetImageVideoFiles(dirPath string) ([]string, error) {
	var files []string
	supportedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true,
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".flv": true, ".m4v": true,
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
