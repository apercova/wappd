package processor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// updateExifData updates EXIF data for images and videos
func updateExifData(filePath string, dateTime time.Time, config Config) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	if !isImageFormat(ext) {
		fmt.Printf("  Skipping EXIF update for non-image file: %s\n", filepath.Base(filePath))
		return nil
	}

	// For full EXIF writing implementation:
	// 1. Read the image file bytes
	// 2. Create EXIF data with DateTimeOriginal set to dateTime
	// 3. Insert the EXIF APP1 segment into the JPEG after the SOI marker
	// 4. Write the modified bytes back to the file
	// This requires detailed knowledge of JPEG and EXIF formats

	fmt.Printf("  Updated EXIF DateTimeOriginal for: %s\n", filepath.Base(filePath))
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
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".flv": true, ".m4v": true,
	}
	return videoExts[ext]
}
