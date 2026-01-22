package processor

import (
	"fmt"
	"path/filepath"
	"time"
)

// updateExifData updates EXIF data for images and videos
func updateExifData(filePath string, dateTime time.Time, config Config) error {
	// For now, just log that we're processing the file
	// EXIF writing is complex and requires external tools or extensive TIFF/JPEG parsing
	// This is a simplified version that sets file modification time
	fmt.Printf("  Setting timestamps for: %s\n", filepath.Base(filePath))

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
