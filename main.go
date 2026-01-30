package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"github.com/apercova/wappd/internal/processor"
)

func main() {
	fmt.Println("wappd starting...")
	
	// Define command-line flags
	filePath := flag.String("f", "", "Path to a specific file to process")
	dirPath := flag.String("d", ".", "Input directory (default: current directory)")
	dateTime := flag.String("dt", "", "ISO format date (YYYY-MM-DD) to override extraction")
	regexPattern := flag.String("e", "", "Custom regex pattern with named group for date extraction")
	patternFormat := flag.String("p", "", "Custom pattern format with {date} placeholder")
	updateModified := flag.Bool("m", false, "Also update file's last modified date")
	overwriteExif := flag.Bool("ow", false, "Overwrite existing EXIF data")
	overrideOriginal := flag.Bool("o", false, "Override original files (don't add suffix)")
	outputDir := flag.String("out", "", "Output directory for processed files")

	flag.Parse()
	fmt.Println("FLAGS PARSED")

	if *filePath != "" && *dirPath != "." {
		log.Println("Warning: -f flag is set, -d flag will be ignored")
	}

	var inputPaths []string
	var err error

	fmt.Printf("filePath=%s, dirPath=%s\n", *filePath, *dirPath)

	if *filePath != "" {
		inputPaths = []string{*filePath}
	} else {
		fmt.Println("calling GetImageVideoFiles...")
		inputPaths, err = processor.GetImageVideoFiles(*dirPath)
		fmt.Printf("GetImageVideoFiles returned, error=%v, count=%d\n", err, len(inputPaths))
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}
	}

	fmt.Printf("Found %d files\n", len(inputPaths))
	if len(inputPaths) == 0 {
		fmt.Println("No image or video files found to process")
		return
	}

	for i, p := range inputPaths {
		dateStr, err := processor.ExtractDateFromFilename(filepath.Base(p), "", "")
		if err != nil {
			fmt.Printf("  %d: %s (date extraction failed: %v)\n", i, p, err)
		} else {
			fmt.Printf("  %d: %s (%s)\n", i, p, dateStr)
		}
	}

	config := processor.Config{
		DateTimeOverride:  *dateTime,
		RegexPattern:      *regexPattern,
		PatternFormat:     *patternFormat,
		UpdateModified:    *updateModified,
		OverwriteExif:     *overwriteExif,
		OverrideOriginal:  *overrideOriginal,
		OutputDir:         *outputDir,
		InputDir:          *dirPath,
	}

	fmt.Println("Creating processor...")
	proc := processor.New(config)
	fmt.Println("Processing files...")
	results := proc.ProcessFiles(inputPaths)
	fmt.Printf("ProcessFiles returned %d results\n", len(results))

	fmt.Printf("\nProcessing complete:\n")
	fmt.Printf("  Total files: %d\n", len(results))
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			fmt.Printf("  ✓ %s → %s\n", r.InputFile, r.OutputFile)
		} else {
			fmt.Printf("  ✗ %s: %v\n", r.InputFile, r.Error)
		}
	}
	fmt.Printf("  Successful: %d\n", successCount)
}
