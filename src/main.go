package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/apercova/wappd/internal/processor"
	"github.com/apercova/wappd/version"
)

func main() {
	// Define command-line flags
	filePath := flag.String("f", "", "Path to a specific file to process")
	dirPath := flag.String("d", ".", "Input directory (default: current directory)")
	var configFile string
	flag.StringVar(&configFile, "cf", "", "Path to config file (default: wappd.json in working directory)")
	flag.StringVar(&configFile, "config-file", "", "Path to config file (alias for -cf)")
	updateModified := flag.Bool("m", false, "Also update file's last modified date")
	overwriteExif := flag.Bool("ow", false, "Overwrite existing EXIF data")
	overrideOriginal := flag.Bool("o", false, "Override original files (don't add suffix)")
	outputDir := flag.String("out", "", "Output directory for processed files")
	verbose := flag.Bool("v", false, "Verbose output (show detailed processing information)")
	dryRun := flag.Bool("dry-run", false, "Preview changes without modifying files")
	showVersion := flag.Bool("version", false, "Show version information")

	// Set custom usage function
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "wappd - WhatsApp Photo Date Extractor\n\n")
		fmt.Fprintf(os.Stderr, "Extracts creation dates from WhatsApp media filenames and restores EXIF/video metadata.\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  wappd [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Process all media in current directory\n")
		fmt.Fprintf(os.Stderr, "  wappd\n\n")
		fmt.Fprintf(os.Stderr, "  # Process specific directory\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./whatsapp_backup\n\n")
		fmt.Fprintf(os.Stderr, "  # Process single file\n")
		fmt.Fprintf(os.Stderr, "  wappd -f IMG-20250122-WA0003.jpg\n\n")
		fmt.Fprintf(os.Stderr, "  # Update file modification time and EXIF\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media -m\n\n")
		fmt.Fprintf(os.Stderr, "  # Override original files\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media -o\n\n")
		fmt.Fprintf(os.Stderr, "  # Save to output directory\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media -out ./processed_media\n\n")
		fmt.Fprintf(os.Stderr, "  # Overwrite existing EXIF data\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media -ow\n\n")
		fmt.Fprintf(os.Stderr, "  # Verbose output\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media -v\n\n")
		fmt.Fprintf(os.Stderr, "  # Dry-run mode (preview changes)\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media --dry-run\n\n")
		fmt.Fprintf(os.Stderr, "  # Use custom config file\n")
		fmt.Fprintf(os.Stderr, "  wappd -d ./media -cf ./my-config.json\n\n")
		fmt.Fprintf(os.Stderr, "Configuration File:\n")
		fmt.Fprintf(os.Stderr, "  Optional wappd.json file in the working directory can set defaults.\n")
		fmt.Fprintf(os.Stderr, "  Use -cf or --config-file to specify a custom config file path.\n")
		fmt.Fprintf(os.Stderr, "  CLI flags override config file values.\n")
		fmt.Fprintf(os.Stderr, "  Example wappd.json:\n")
		fmt.Fprintf(os.Stderr, "    {\n")
		fmt.Fprintf(os.Stderr, "      \"updateModified\": true,\n")
		fmt.Fprintf(os.Stderr, "      \"outputDir\": \"./processed\",\n")
		fmt.Fprintf(os.Stderr, "      \"verbose\": false\n")
		fmt.Fprintf(os.Stderr, "    }\n\n")
		fmt.Fprintf(os.Stderr, "Supported Formats:\n")
		fmt.Fprintf(os.Stderr, "  Images: JPG, JPEG, PNG, GIF, BMP, WebP\n")
		fmt.Fprintf(os.Stderr, "  Videos: MP4, MOV, AVI, MKV, FLV, M4V, 3GP\n\n")
		fmt.Fprintf(os.Stderr, "WhatsApp Filename Patterns:\n")
		fmt.Fprintf(os.Stderr, "  Images: IMG-YYYYMMDD-WA####.ext\n")
		fmt.Fprintf(os.Stderr, "  Videos: VID-YYYYMMDD-WA####.ext\n")
		fmt.Fprintf(os.Stderr, "  Images: WhatsApp Image YYYY-MM-DD at H.MM.SS AM|PM.ext\n")
		fmt.Fprintf(os.Stderr, "  Videos: WhatsApp Video YYYY-MM-DD at H.MM.SS AM|PM.ext\n\n")
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Println(version.Get().String())
		os.Exit(0)
	}

	if *filePath != "" && *dirPath != "." {
		log.Println("Warning: -f flag is set, -d flag will be ignored")
	}

	var inputPaths []string
	var err error

	if *filePath != "" {
		inputPaths = []string{*filePath}
	} else {
		if *verbose {
			fmt.Println("Scanning directory for media files...")
		}
		inputPaths, err = processor.GetImageVideoFiles(*dirPath)
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}
	}

	if len(inputPaths) == 0 {
		fmt.Println("No image or video files found to process")
		return
	}

	if *verbose {
		fmt.Printf("Found %d file(s) to process\n", len(inputPaths))
		for i, p := range inputPaths {
			dateStr, err := processor.ExtractDateFromFilename(filepath.Base(p))
			if err != nil {
				fmt.Printf("  %d: %s (date extraction failed: %v)\n", i+1, p, err)
			} else {
				fmt.Printf("  %d: %s → %s\n", i+1, p, dateStr)
			}
		}
		fmt.Println()
	}

	// Load config file if specified or if default exists (optional)
	var fileConfig *processor.ConfigFile
	if configFile != "" {
		// Use custom config file path
		fileConfig, err = processor.LoadConfigFileFromPath(configFile)
		if err != nil {
			log.Fatalf("Failed to load config file %s: %v", configFile, err)
		}
	} else {
		// Try default config file in working directory
		fileConfig, err = processor.LoadConfigFile(*dirPath)
		if err != nil {
			log.Printf("Warning: Failed to load config file: %v", err)
		}
	}

	// Build CLI config
	cliConfig := processor.Config{
		UpdateModified:    *updateModified,
		OverwriteExif:     *overwriteExif,
		OverrideOriginal:  *overrideOriginal,
		OutputDir:         *outputDir,
		InputDir:          *dirPath,
		Verbose:           *verbose,
		DryRun:            *dryRun,
	}

	// Merge config file with CLI flags (CLI takes precedence)
	config := processor.MergeConfig(fileConfig, cliConfig)

	// Show config file usage if loaded
	if fileConfig != nil && config.Verbose {
		configPath := configFile
		if configPath == "" {
			configPath = filepath.Join(*dirPath, processor.ConfigFileName())
		}
		fmt.Printf("Loaded configuration from %s\n", configPath)
	}

	if config.DryRun {
		fmt.Println("DRY-RUN MODE: No files will be modified")
		fmt.Println()
	}
	if config.Verbose {
		fmt.Println("Processing files...")
	}
	proc := processor.New(config)
	results := proc.ProcessFiles(inputPaths)

	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			if config.Verbose {
				fmt.Printf("  ✓ %s → %s\n", r.InputFile, r.OutputFile)
			}
		} else {
			failCount++
			fmt.Printf("  ✗ %s: %v\n", r.InputFile, r.Error)
		}
	}

	if config.DryRun {
		fmt.Printf("\nDry-run complete: %d files would be processed", successCount)
		if failCount > 0 {
			fmt.Printf(", %d would fail", failCount)
		}
		fmt.Printf(" (out of %d total)\n", len(results))
		fmt.Println("Run without --dry-run to apply changes")
	} else {
		fmt.Printf("\nProcessing complete: %d successful", successCount)
		if failCount > 0 {
			fmt.Printf(", %d failed", failCount)
		}
		fmt.Printf(" (out of %d total)\n", len(results))
	}
}
