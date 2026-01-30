package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// updateExifData updates EXIF data for images and videos
func updateExifData(filePath string, dateTime time.Time, config Config) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Handle video files (MP4, MOV, M4V, 3GP)
	if ext == ".mp4" || ext == ".mov" || ext == ".m4v" || ext == ".3gp" {
		if config.DryRun {
			if config.Verbose {
				fmt.Printf("  [DRY-RUN] Would update video creation date for: %s\n", filepath.Base(filePath))
			}
			return nil
		}
		err := UpdateVideoMetadata(filePath, dateTime)
		if err != nil {
			return fmt.Errorf("failed to update video metadata: %v", err)
		}
		if config.Verbose {
			fmt.Printf("  Updated video creation date for: %s\n", filepath.Base(filePath))
		}
		return nil
	}

	// Handle JPEG files (EXIF)
	if ext == ".jpg" || ext == ".jpeg" {
		return updateJPEGExif(filePath, dateTime, config)
	}

	// Skip other formats
	if config.Verbose {
		fmt.Printf("  Skipping metadata update for unsupported file type: %s\n", filepath.Base(filePath))
	}
	return nil
}

// updateJPEGExif updates EXIF data for JPEG files
func updateJPEGExif(filePath string, dateTime time.Time, config Config) error {
	// In dry-run mode, skip actual file operations
	if config.DryRun {
		if config.Verbose {
			fmt.Printf("  [DRY-RUN] Would update EXIF DateTimeOriginal for: %s\n", filepath.Base(filePath))
		}
		return nil
	}

	// Read the JPEG file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// Verify it's a valid JPEG
	if len(data) < 2 || data[0] != 0xFF || data[1] != 0xD8 {
		return fmt.Errorf("file is not a valid JPEG")
	}

	// Check if EXIF already exists
	segments, err := ParseJPEGSegments(data)
	if err != nil {
		return fmt.Errorf("failed to parse JPEG segments: %v", err)
	}
	
	_, existingAPP1 := FindAPP1Segment(segments)

	// If EXIF exists and we're not overwriting, skip
	if existingAPP1 != nil && !config.OverwriteExif {
		if config.Verbose {
			fmt.Printf("  EXIF already exists in %s (use -ow to overwrite)\n", filepath.Base(filePath))
		}
		return nil
	}

	// Create EXIF segment
	exifPayload, err := CreateEXIFSegment(dateTime)
	if err != nil {
		return fmt.Errorf("failed to create EXIF segment: %v", err)
	}

	// Insert EXIF segment into JPEG
	newJPEG, err := InsertEXIFSegment(data, exifPayload)
	if err != nil {
		return fmt.Errorf("failed to insert EXIF segment: %v", err)
	}

	// Write the modified JPEG back to file
	// Preserve original file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	err = os.WriteFile(filePath, newJPEG, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	if config.Verbose {
		fmt.Printf("  Updated EXIF DateTimeOriginal for: %s\n", filepath.Base(filePath))
	}
	return nil
}

// isImageFormat checks if the file is an image
func isImageFormat(ext string) bool {
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true,
	}
	return imageExts[ext]
}

// isVideoFormat checks if the file is a video
func isVideoFormat(ext string) bool {
	videoExts := map[string]bool{
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".flv": true, ".m4v": true, ".3gp": true,
	}
	return videoExts[ext]
}
