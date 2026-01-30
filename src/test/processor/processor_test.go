package processor_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apercova/wappd/internal/processor"
)

func TestExtractDateFromFilename_DefaultPatterns(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		// IMG pattern tests
		{
			name:     "WhatsApp image pattern",
			filename: "IMG-20250122-WA0003.jpg",
			want:     "2025-01-22",
			wantErr:  false,
		},
		{
			name:     "WhatsApp image pattern with different extension",
			filename: "IMG-20250122-WA0003.jpeg",
			want:     "2025-01-22",
			wantErr:  false,
		},
		{
			name:     "WhatsApp image pattern with PNG",
			filename: "IMG-20240415-WA0010.png",
			want:     "2024-04-15",
			wantErr:  false,
		},
		// VID pattern tests
		{
			name:     "WhatsApp video pattern",
			filename: "VID-20240415-WA0010.mp4",
			want:     "2024-04-15",
			wantErr:  false,
		},
		{
			name:     "WhatsApp video pattern with MOV",
			filename: "VID-20250122-WA0003.mov",
			want:     "2025-01-22",
			wantErr:  false,
		},
		{
			name:     "WhatsApp video pattern with 3GP",
			filename: "VID-20240415-WA0010.3gp",
			want:     "2024-04-15",
			wantErr:  false,
		},
		// WhatsApp Image with time pattern tests
		{
			name:     "WhatsApp Image with time pattern PM",
			filename: "WhatsApp Image 2025-01-22 at 3.30.45 PM.jpg",
			want:     "2025-01-22T15:30:45",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Image with time pattern AM",
			filename: "WhatsApp Image 2024-04-15 at 10.15.30 AM.jpg",
			want:     "2024-04-15T10:15:30",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Image with time pattern single digit hour",
			filename: "WhatsApp Image 2025-01-22 at 9.05.00 AM.jpg",
			want:     "2025-01-22T09:05:00",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Image with time pattern midnight",
			filename: "WhatsApp Image 2025-01-22 at 12.00.00 AM.jpg",
			want:     "2025-01-22T00:00:00",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Image with time pattern noon",
			filename: "WhatsApp Image 2025-01-22 at 12.00.00 PM.jpg",
			want:     "2025-01-22T12:00:00",
			wantErr:  false,
		},
		// WhatsApp Video with time pattern tests
		{
			name:     "WhatsApp Video with time pattern AM",
			filename: "WhatsApp Video 2024-04-15 at 10.15.30 AM.mp4",
			want:     "2024-04-15T10:15:30",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Video with time pattern PM",
			filename: "WhatsApp Video 2025-01-22 at 3.30.45 PM.mp4",
			want:     "2025-01-22T15:30:45",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Video with time pattern single digit hour",
			filename: "WhatsApp Video 2024-04-15 at 5.20.10 AM.mp4",
			want:     "2024-04-15T05:20:10",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Video with time pattern MOV extension",
			filename: "WhatsApp Video 2025-01-22 at 11.59.59 PM.mov",
			want:     "2025-01-22T23:59:59",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Video with time pattern 3GP extension",
			filename: "WhatsApp Video 2024-04-15 at 10.15.30 AM.3gp",
			want:     "2024-04-15T10:15:30",
			wantErr:  false,
		},
		// Edge cases
		{
			name:     "Filename with path",
			filename: "/path/to/IMG-20250122-WA0003.jpg",
			want:     "2025-01-22",
			wantErr:  false,
		},
		{
			name:     "Filename with Windows path",
			filename: "C:\\Users\\User\\VID-20240415-WA0010.mp4",
			want:     "2024-04-15",
			wantErr:  false,
		},
		{
			name:     "Filename with Windows path 3GP",
			filename: "C:\\Users\\User\\VID-20240415-WA0010.3gp",
			want:     "2024-04-15",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Image with path",
			filename: "/backup/WhatsApp Image 2025-01-22 at 3.30.45 PM.jpg",
			want:     "2025-01-22T15:30:45",
			wantErr:  false,
		},
		// Invalid cases
		{
			name:     "Invalid filename",
			filename: "random-file.jpg",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid IMG pattern",
			filename: "IMG-2025012-WA0003.jpg",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid VID pattern",
			filename: "VID-2024041-WA0010.mp4",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "WhatsApp Image without time",
			filename: "WhatsApp Image 2025-01-22.jpg",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "WhatsApp Video without time",
			filename: "WhatsApp Video 2025-01-22.mp4",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Invalid VID pattern 3GP",
			filename: "VID-2024041-WA0010.3gp",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processor.ExtractDateFromFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractDateFromFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractDateFromFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractDateFromFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  bool
	}{
		{
			name:     "Empty filename",
			filename: "",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Filename without extension",
			filename: "IMG-20250122-WA0003",
			want:     "2025-01-22",
			wantErr:  false,
		},
		{
			name:     "WhatsApp Image without extension",
			filename: "WhatsApp Image 2025-01-22 at 3.30.45 PM",
			want:     "",
			wantErr:  true, // WhatsApp files typically have extensions
		},
		{
			name:     "Case sensitivity - lowercase img",
			filename: "img-20250122-wa0003.jpg",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "Case sensitivity - lowercase vid",
			filename: "vid-20240415-wa0010.mp4",
			want:     "",
			wantErr:  true,
		},
		{
			name:     "WhatsApp Image lowercase",
			filename: "whatsapp image 2025-01-22 at 3.30.45 pm.jpg",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processor.ExtractDateFromFilename(tt.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractDateFromFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractDateFromFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetImageVideoFiles_3GP(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"VID-20240415-WA0010.3gp",
		"VID-20240415-WA0011.mp4",
		"VID-20240415-WA0012.mov",
		"IMG-20240415-WA0010.jpg",
		"document.txt", // Should be ignored
	}

	for _, filename := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Get image and video files
	files, err := processor.GetImageVideoFiles(tmpDir)
	if err != nil {
		t.Fatalf("GetImageVideoFiles() error = %v", err)
	}

	// Verify 3GP file is included
	found3GP := false
	for _, file := range files {
		if filepath.Ext(file) == ".3gp" {
			found3GP = true
			break
		}
	}

	if !found3GP {
		t.Error("GetImageVideoFiles() did not find 3GP file")
	}

	// Verify we got the expected number of media files (4: 3GP, MP4, MOV, JPG)
	if len(files) != 4 {
		t.Errorf("GetImageVideoFiles() returned %d files, want 4", len(files))
	}
}
