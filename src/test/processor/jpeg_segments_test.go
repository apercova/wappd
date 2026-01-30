package processor_test

import (
	"testing"
	"time"

	"github.com/apercova/wappd/internal/processor"
)

func TestParseJPEGSegments(t *testing.T) {
	// Create a minimal valid JPEG structure
	// SOI (2 bytes) + APP1 segment + SOF0 + image data + EOI
	jpegData := []byte{
		0xFF, 0xD8, // SOI
		0xFF, 0xE1, // APP1 marker
		0x00, 0x10, // Length (16 bytes)
		'E', 'x', 'i', 'f', 0x00, 0x00, // "Exif\0\0"
		0x49, 0x49, 0x2A, 0x00, 0x08, 0x00, 0x00, 0x00, // TIFF header
		0xFF, 0xC0, // SOF0 marker
	}

	segments, err := processor.ParseJPEGSegments(jpegData)
	if err != nil {
		t.Fatalf("ParseJPEGSegments() error = %v", err)
	}

	if len(segments) == 0 {
		t.Error("ParseJPEGSegments() returned no segments")
	}

	// Should find APP1 segment
	found := false
	for _, seg := range segments {
		if seg.Marker == 0xE1 {
			found = true
			if len(seg.Payload) == 0 {
				t.Error("APP1 segment has no payload")
			}
		}
	}
	if !found {
		t.Error("ParseJPEGSegments() did not find APP1 segment")
	}
}

func TestParseJPEGSegments_InvalidJPEG(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Too short",
			data: []byte{0xFF},
		},
		{
			name: "Missing SOI",
			data: []byte{0xFF, 0xFF},
		},
		{
			name: "Empty file",
			data: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := processor.ParseJPEGSegments(tt.data)
			if err == nil {
				t.Errorf("ParseJPEGSegments() expected error for %s", tt.name)
			}
		})
	}
}

func TestFindAPP1Segment(t *testing.T) {
	segments := []processor.JPEGSegment{
		{Marker: 0xE0, Length: 10, Payload: []byte("JFIF data")},
		{Marker: 0xE1, Length: 20, Payload: []byte("Exif\x00\x00" + "TIFF data")},
		{Marker: 0xE2, Length: 10, Payload: []byte("Other data")},
	}

	index, seg := processor.FindAPP1Segment(segments)
	if index == -1 {
		t.Error("FindAPP1Segment() did not find APP1 segment")
	}
	if seg == nil {
		t.Error("FindAPP1Segment() returned nil segment")
	}
	if seg.Marker != 0xE1 {
		t.Errorf("FindAPP1Segment() marker = 0x%02x, want 0xE1", seg.Marker)
	}
}

func TestFindAPP1Segment_NotFound(t *testing.T) {
	segments := []processor.JPEGSegment{
		{Marker: 0xE0, Length: 10, Payload: []byte("JFIF data")},
		{Marker: 0xE2, Length: 10, Payload: []byte("Other data")},
	}

	index, seg := processor.FindAPP1Segment(segments)
	if index != -1 {
		t.Errorf("FindAPP1Segment() index = %d, want -1", index)
	}
	if seg != nil {
		t.Error("FindAPP1Segment() should return nil when not found")
	}
}

func TestReassembleJPEG(t *testing.T) {
	segments := []processor.JPEGSegment{
		{Marker: 0xE1, Length: 10, Payload: []byte("test data")},
	}
	imageData := []byte{0xFF, 0xC0, 0x00, 0x11} // SOF0 + some data

	result := processor.ReassembleJPEG(segments, imageData)

	// Should start with SOI
	if result[0] != 0xFF || result[1] != 0xD8 {
		t.Error("ReassembleJPEG() does not start with SOI")
	}

	// Should contain segment data
	if len(result) < len(imageData)+10 {
		t.Error("ReassembleJPEG() result too short")
	}

	// Should end with EOI
	if result[len(result)-2] != 0xFF || result[len(result)-1] != 0xD9 {
		t.Error("ReassembleJPEG() does not end with EOI")
	}
}

func TestCreateEXIFSegment(t *testing.T) {
	dateTime := time.Date(2025, 1, 22, 15, 30, 45, 0, time.UTC)

	exifPayload, err := processor.CreateEXIFSegment(dateTime)
	if err != nil {
		t.Fatalf("CreateEXIFSegment() error = %v", err)
	}

	if len(exifPayload) == 0 {
		t.Error("CreateEXIFSegment() returned empty payload")
	}

	// Should start with "Exif\0\0"
	exifID := string(exifPayload[0:6])
	if exifID != "Exif\x00\x00" {
		t.Errorf("CreateEXIFSegment() EXIF ID = %q, want \"Exif\\x00\\x00\"", exifID)
	}

	// Should contain TIFF header
	if len(exifPayload) < 14 {
		t.Error("CreateEXIFSegment() payload too short")
	}
}
